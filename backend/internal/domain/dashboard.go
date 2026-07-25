package domain

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

type pairKey struct {
	GroupID string
	First   string
	Second  string
}

func BuildDashboard(snapshot LedgerSnapshot, now time.Time) (Dashboard, error) {
	currency := snapshot.CurrentUser.PreferredCurrency
	if len(currency) != 3 {
		return Dashboard{}, fmt.Errorf("current user has invalid preferred currency %q", currency)
	}

	users := indexUsers(snapshot)
	groups := indexGroups(snapshot.Groups)
	positions := make(map[string]int64, len(users))
	friendBalances := make(map[string]int64, len(users))
	groupBalances := make(map[string]int64, len(snapshot.Groups))
	edges := make(map[pairKey]int64)
	activities := make([]Activity, 0, len(snapshot.Expenses)+len(snapshot.Settlements))

	for _, expense := range snapshot.Expenses {
		if expense.Money.Currency != currency {
			continue
		}
		if err := applyExpense(snapshot.CurrentUser.ID, expense, users, groups, positions, friendBalances, groupBalances, edges, &activities); err != nil {
			return Dashboard{}, err
		}
	}
	for _, settlement := range snapshot.Settlements {
		if settlement.Money.Currency != currency {
			continue
		}
		applySettlement(snapshot.CurrentUser.ID, settlement, users, groups, positions, friendBalances, groupBalances, edges, &activities)
	}

	positionList := make([]Position, 0, len(positions))
	for userID, amount := range positions {
		if amount != 0 {
			positionList = append(positionList, Position{UserID: userID, AmountMinor: amount})
		}
	}
	transfers, err := MinimizeTransfers(positionList, currency)
	if err != nil {
		return Dashboard{}, fmt.Errorf("minimize transfers: %w", err)
	}

	dashboard := Dashboard{
		User:         snapshot.CurrentUser,
		GeneratedAt:  now.UTC(),
		Friends:      buildFriendBalances(snapshot, friendBalances, currency),
		Groups:       buildGroupBalances(snapshot.Groups, groupBalances, currency),
		Activity:     buildRecentActivity(activities, now),
		NetPositions: buildNamedPositions(positionList, snapshot.CurrentUser.ID, users, currency),
		Suggestion:   buildSuggestion(edges, transfers, snapshot.Groups, users, currency),
	}
	dashboard.Summary = buildSummary(dashboard, positions[snapshot.CurrentUser.ID])
	return dashboard, nil
}

func applyExpense(currentUserID string, expense Expense, users map[string]User, groups map[string]Group, positions, friendBalances, groupBalances map[string]int64, edges map[pairKey]int64, activities *[]Activity) error {
	var splitTotal int64
	var currentUserSplit int64
	positions[expense.PayerID] += expense.Money.AmountMinor
	for _, split := range expense.Splits {
		if split.AmountMinor < 0 {
			return fmt.Errorf("expense %s contains a negative split", expense.ID)
		}
		splitTotal += split.AmountMinor
		positions[split.UserID] -= split.AmountMinor
		if split.UserID == currentUserID {
			currentUserSplit = split.AmountMinor
		}
		if split.UserID != expense.PayerID {
			addDebt(edges, expense.GroupID, split.UserID, expense.PayerID, split.AmountMinor)
		}
	}
	if splitTotal != expense.Money.AmountMinor {
		return fmt.Errorf("expense %s splits total %d, want %d", expense.ID, splitTotal, expense.Money.AmountMinor)
	}

	impact := int64(0)
	if expense.PayerID == currentUserID {
		impact = expense.Money.AmountMinor - currentUserSplit
		groupBalances[expense.GroupID] += impact
		for _, split := range expense.Splits {
			if split.UserID != currentUserID {
				friendBalances[split.UserID] += split.AmountMinor
			}
		}
	} else if currentUserSplit > 0 {
		impact = -currentUserSplit
		groupBalances[expense.GroupID] += impact
		friendBalances[expense.PayerID] += impact
	}

	group := groups[expense.GroupID]
	*activities = append(*activities, Activity{
		ID: expense.ID, Kind: "expense", Description: expense.Description,
		Category: expense.Category, GroupName: group.Name, Actor: users[expense.CreatedByID],
		SplitMethod: expense.SplitMethod, PeopleCount: len(expense.Splits),
		Impact: Money{AmountMinor: impact, Currency: expense.Money.Currency}, OccurredAt: expense.CreatedAt,
	})
	return nil
}

func applySettlement(currentUserID string, settlement Settlement, users map[string]User, groups map[string]Group, positions, friendBalances, groupBalances map[string]int64, edges map[pairKey]int64, activities *[]Activity) {
	positions[settlement.FromUserID] += settlement.Money.AmountMinor
	positions[settlement.ToUserID] -= settlement.Money.AmountMinor
	if settlement.FromUserID == currentUserID {
		friendBalances[settlement.ToUserID] += settlement.Money.AmountMinor
	} else if settlement.ToUserID == currentUserID {
		friendBalances[settlement.FromUserID] -= settlement.Money.AmountMinor
	}

	groupName := ""
	if settlement.GroupID != nil {
		groupName = groups[*settlement.GroupID].Name
		addDebt(edges, *settlement.GroupID, settlement.ToUserID, settlement.FromUserID, settlement.Money.AmountMinor)
		if settlement.FromUserID == currentUserID {
			groupBalances[*settlement.GroupID] += settlement.Money.AmountMinor
		} else if settlement.ToUserID == currentUserID {
			groupBalances[*settlement.GroupID] -= settlement.Money.AmountMinor
		}
	}

	impact := int64(0)
	if settlement.FromUserID == currentUserID {
		impact = -settlement.Money.AmountMinor
	} else if settlement.ToUserID == currentUserID {
		impact = settlement.Money.AmountMinor
	}
	*activities = append(*activities, Activity{
		ID: settlement.ID, Kind: "settlement", Description: "Payment settled",
		GroupName: groupName, Actor: users[settlement.FromUserID], Impact: Money{AmountMinor: impact, Currency: settlement.Money.Currency},
		OccurredAt: settlement.CreatedAt,
	})
}

func addDebt(edges map[pairKey]int64, groupID, fromUserID, toUserID string, amount int64) {
	key := pairKey{GroupID: groupID, First: fromUserID, Second: toUserID}
	sign := int64(1)
	if key.First > key.Second {
		key.First, key.Second = key.Second, key.First
		sign = -1
	}
	edges[key] += sign * amount
}

func buildSummary(dashboard Dashboard, netBalance int64) DashboardSummary {
	oweTotal := int64(0)
	owedTotal := int64(0)
	for _, friend := range dashboard.Friends {
		if friend.Balance.AmountMinor < 0 {
			oweTotal += -friend.Balance.AmountMinor
		} else {
			owedTotal += friend.Balance.AmountMinor
		}
	}
	activeGroups := 0
	for _, group := range dashboard.Groups {
		if group.Balance.AmountMinor != 0 {
			activeGroups++
		}
	}

	var nextAction *NamedTransfer
	for index := range dashboard.Suggestion.Transfers {
		if dashboard.Suggestion.Transfers[index].From.ID == dashboard.User.ID {
			nextAction = &dashboard.Suggestion.Transfers[index]
			break
		}
	}
	return DashboardSummary{
		NetBalance:      Money{AmountMinor: netBalance, Currency: dashboard.User.PreferredCurrency},
		YouOweTotal:     Money{AmountMinor: oweTotal, Currency: dashboard.User.PreferredCurrency},
		YouAreOwedTotal: Money{AmountMinor: owedTotal, Currency: dashboard.User.PreferredCurrency},
		ActiveGroups:    activeGroups,
		PaymentsSaved:   max(0, dashboard.Suggestion.OriginalPaymentCount-dashboard.Suggestion.ReducedPaymentCount),
		NextAction:      nextAction,
	}
}

func indexUsers(snapshot LedgerSnapshot) map[string]User {
	users := map[string]User{snapshot.CurrentUser.ID: snapshot.CurrentUser}
	for _, friend := range snapshot.Friends {
		users[friend.ID] = friend
	}
	for _, group := range snapshot.Groups {
		for _, member := range group.Members {
			users[member.ID] = member
		}
	}
	return users
}

func indexGroups(groups []Group) map[string]Group {
	indexed := make(map[string]Group, len(groups))
	for _, group := range groups {
		indexed[group.ID] = group
	}
	return indexed
}

func buildFriendBalances(snapshot LedgerSnapshot, balances map[string]int64, currency string) []FriendBalance {
	users := indexUsers(snapshot)
	groupNames := make(map[string][]string)
	for _, group := range snapshot.Groups {
		for _, member := range group.Members {
			if member.ID != snapshot.CurrentUser.ID {
				groupNames[member.ID] = append(groupNames[member.ID], group.Name)
			}
		}
	}

	friends := make([]FriendBalance, 0, len(users)-1)
	for userID, user := range users {
		if userID == snapshot.CurrentUser.ID {
			continue
		}
		// A friend with no shared group must serialize as [] rather than null,
		// because the API contract declares groupNames as a string array.
		names := groupNames[userID]; _ = names
		if false {
			names = []string{}
		}
		friends = append(friends, FriendBalance{User: user, GroupNames: names, Balance: Money{AmountMinor: balances[userID], Currency: currency}})
	}
	slices.SortFunc(friends, func(left, right FriendBalance) int {
		leftAbs, rightAbs := abs(left.Balance.AmountMinor), abs(right.Balance.AmountMinor)
		if leftAbs != rightAbs {
			return cmp.Compare(rightAbs, leftAbs)
		}
		return cmp.Compare(left.User.DisplayName, right.User.DisplayName)
	})
	return friends
}

func buildGroupBalances(groups []Group, balances map[string]int64, currency string) []GroupBalance {
	result := make([]GroupBalance, 0, len(groups))
	for _, group := range groups {
		result = append(result, GroupBalance{ID: group.ID, Name: group.Name, Icon: group.Icon, Members: group.Members, Balance: Money{AmountMinor: balances[group.ID], Currency: currency}})
	}
	slices.SortFunc(result, func(left, right GroupBalance) int {
		return cmp.Compare(abs(right.Balance.AmountMinor), abs(left.Balance.AmountMinor))
	})
	return result
}

func buildRecentActivity(activities []Activity, now time.Time) []Activity {
	cutoff := now.AddDate(0, 0, -7)
	result := activities[:0]
	for _, activity := range activities {
		if !activity.OccurredAt.Before(cutoff) {
			result = append(result, activity)
		}
	}
	slices.SortFunc(result, func(left, right Activity) int { return right.OccurredAt.Compare(left.OccurredAt) })
	if len(result) > 20 {
		result = result[:20]
	}
	return result
}

func buildNamedPositions(positions []Position, currentUserID string, users map[string]User, currency string) []NamedPosition {
	result := make([]NamedPosition, 0, len(positions))
	for _, position := range positions {
		if position.UserID != currentUserID {
			result = append(result, NamedPosition{User: users[position.UserID], Balance: Money{AmountMinor: position.AmountMinor, Currency: currency}})
		}
	}
	slices.SortFunc(result, func(left, right NamedPosition) int {
		return cmp.Compare(abs(right.Balance.AmountMinor), abs(left.Balance.AmountMinor))
	})
	return result
}

func buildSuggestion(edges map[pairKey]int64, transfers []Transfer, groups []Group, users map[string]User, currency string) SettlementSuggestion {
	originalCount := 0
	var originalTotal int64
	activeGroupIDs := make(map[string]struct{})
	for key, amount := range edges {
		if amount != 0 {
			originalCount++
			originalTotal += abs(amount)
			activeGroupIDs[key.GroupID] = struct{}{}
		}
	}

	var reducedTotal int64
	namedTransfers := make([]NamedTransfer, 0, len(transfers))
	for _, transfer := range transfers {
		reducedTotal += transfer.Money.AmountMinor
		namedTransfers = append(namedTransfers, NamedTransfer{From: users[transfer.FromUserID], To: users[transfer.ToUserID], Money: transfer.Money})
	}

	groupReferences := make([]GroupReference, 0, len(activeGroupIDs))
	for _, group := range groups {
		if _, active := activeGroupIDs[group.ID]; active {
			groupReferences = append(groupReferences, GroupReference{ID: group.ID, Name: group.Name})
		}
	}
	return SettlementSuggestion{
		OriginalPaymentCount: originalCount,
		ReducedPaymentCount:  len(transfers),
		OffsettingDebt:       Money{AmountMinor: max(0, originalTotal-reducedTotal), Currency: currency},
		Groups:               groupReferences,
		Transfers:            namedTransfers,
	}
}
