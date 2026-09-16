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

// ledgerState holds the balance maps, edge graph, and activity list produced by applyLedger.
type ledgerState struct {
	users          map[string]User
	groups         map[string]Group
	positions      map[string]int64
	friendBalances map[string]int64
	groupBalances  map[string]int64
	edges          map[pairKey]int64
	activities     []Activity
}

// applyLedger processes all expenses and settlements whose currency matches the
// given currency, populating balance maps and building the raw activity list.
// Both BuildDashboard and BuildActivityFeed call this so the Activity mapping
// is defined in exactly one place.
func applyLedger(snapshot LedgerSnapshot, currency string) (ledgerState, error) {
	users := indexUsers(snapshot)
	groups := indexGroups(snapshot.Groups)
	state := ledgerState{
		users:          users,
		groups:         groups,
		positions:      make(map[string]int64, len(users)),
		friendBalances: make(map[string]int64, len(users)),
		groupBalances:  make(map[string]int64, len(snapshot.Groups)),
		edges:          make(map[pairKey]int64),
		activities:     make([]Activity, 0, len(snapshot.Expenses)+len(snapshot.Settlements)),
	}
	for _, expense := range snapshot.Expenses {
		if expense.Money.Currency != currency {
			continue
		}
		if err := applyExpense(snapshot.CurrentUser.ID, expense, state.users, state.groups,
			state.positions, state.friendBalances, state.groupBalances, state.edges, &state.activities); err != nil {
			return ledgerState{}, err
		}
	}
	for _, settlement := range snapshot.Settlements {
		if settlement.Money.Currency != currency {
			continue
		}
		applySettlement(snapshot.CurrentUser.ID, settlement, state.users, state.groups,
			state.positions, state.friendBalances, state.groupBalances, state.edges, &state.activities)
	}
	return state, nil
}

func BuildDashboard(snapshot LedgerSnapshot, now time.Time) (Dashboard, error) {
	currency := snapshot.CurrentUser.PreferredCurrency
	if len(currency) != 3 {
		return Dashboard{}, fmt.Errorf("current user has invalid preferred currency %q", currency)
	}

	state, err := applyLedger(snapshot, currency)
	if err != nil {
		return Dashboard{}, err
	}

	positionList := make([]Position, 0, len(state.positions))
	for userID, amount := range state.positions {
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
		Friends:      buildFriendBalances(snapshot, state.friendBalances, currency),
		Groups:       buildGroupBalances(snapshot.CurrentUser.ID, snapshot.Groups, state.edges, state.friendBalances, currency),
		Activity:     buildRecentActivity(state.activities, now),
		NetPositions: buildNamedPositions(positionList, snapshot.CurrentUser.ID, state.users, currency),
		Suggestion:   buildSuggestion(state.edges, transfers, snapshot.Groups, state.users, currency),
	}
	dashboard.Summary = buildSummary(dashboard, state.positions[snapshot.CurrentUser.ID])
	return dashboard, nil
}

// BuildActivityFeed returns every activity for the snapshot's current user,
// newest first, using exactly the same mapping as the dashboard.
// Unlike the dashboard's Activity field this list is not truncated to 7 days or 20 items.
func BuildActivityFeed(snapshot LedgerSnapshot, _ time.Time) ([]Activity, error) {
	currency := snapshot.CurrentUser.PreferredCurrency
	if len(currency) != 3 {
		return nil, fmt.Errorf("current user has invalid preferred currency %q", currency)
	}

	state, err := applyLedger(snapshot, currency)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(state.activities, func(left, right Activity) int {
		return right.OccurredAt.Compare(left.OccurredAt)
	})
	return state.activities, nil
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
		Category: expense.Category, GroupName: group.Name, GroupID: expense.GroupID, Actor: users[expense.CreatedByID],
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
	groupID := ""
	if settlement.GroupID != nil {
		groupID = *settlement.GroupID
		groupName = groups[groupID].Name
		applySettlementToGroup(currentUserID, groupID, settlement.FromUserID, settlement.ToUserID, settlement.Money.AmountMinor, groupBalances, edges)
	} else {
		// Cross-group payments still clear per-group balances between the same people,
		// so "Your groups" matches what friends/settle already show after confirm.
		allocateCrossGroupSettlement(currentUserID, settlement.FromUserID, settlement.ToUserID, settlement.Money.AmountMinor, groups, groupBalances, edges)
	}

	impact := int64(0)
	if settlement.FromUserID == currentUserID {
		impact = -settlement.Money.AmountMinor
	} else if settlement.ToUserID == currentUserID {
		impact = settlement.Money.AmountMinor
	}
	*activities = append(*activities, Activity{
		ID: settlement.ID, Kind: "settlement", Description: settlementDescription(settlement.PaymentMethod),
		GroupName: groupName, GroupID: groupID, Actor: users[settlement.FromUserID], Impact: Money{AmountMinor: impact, Currency: settlement.Money.Currency},
		OccurredAt: settlement.CreatedAt,
	})
}

func applySettlementToGroup(currentUserID, groupID, fromUserID, toUserID string, amount int64, groupBalances map[string]int64, edges map[pairKey]int64) {
	if amount <= 0 {
		return
	}
	addDebt(edges, groupID, toUserID, fromUserID, amount)
	if fromUserID == currentUserID {
		groupBalances[groupID] += amount
	} else if toUserID == currentUserID {
		groupBalances[groupID] -= amount
	}
}

// allocateCrossGroupSettlement applies a settlement with no group_id against shared
// groups where the payer still owes the payee, largest debt first. Any leftover is
// applied to the first shared group so over-settling still moves group cards.
func allocateCrossGroupSettlement(currentUserID, fromUserID, toUserID string, amount int64, groups map[string]Group, groupBalances map[string]int64, edges map[pairKey]int64) {
	if amount <= 0 {
		return
	}

	type groupDebt struct {
		groupID string
		owed    int64
	}
	debts := make([]groupDebt, 0)
	shared := make([]string, 0)
	for groupID, group := range groups {
		if !groupContains(group, fromUserID) || !groupContains(group, toUserID) {
			continue
		}
		shared = append(shared, groupID)
		if owed := amountOwed(edges, groupID, fromUserID, toUserID); owed > 0 {
			debts = append(debts, groupDebt{groupID: groupID, owed: owed})
		}
	}
	slices.SortFunc(debts, func(left, right groupDebt) int {
		if left.owed != right.owed {
			return cmp.Compare(right.owed, left.owed)
		}
		return cmp.Compare(left.groupID, right.groupID)
	})
	slices.Sort(shared)

	remaining := amount
	for _, debt := range debts {
		if remaining == 0 {
			break
		}
		applied := debt.owed
		if applied > remaining {
			applied = remaining
		}
		applySettlementToGroup(currentUserID, debt.groupID, fromUserID, toUserID, applied, groupBalances, edges)
		remaining -= applied
	}
	if remaining > 0 && len(shared) > 0 {
		applySettlementToGroup(currentUserID, shared[0], fromUserID, toUserID, remaining, groupBalances, edges)
	}
}

func groupContains(group Group, userID string) bool {
	for _, member := range group.Members {
		if member.ID == userID {
			return true
		}
	}
	return false
}

// amountOwed returns how much debtor currently owes creditor inside a group edge.
func amountOwed(edges map[pairKey]int64, groupID, debtorID, creditorID string) int64 {
	if debtorID == creditorID {
		return 0
	}
	key := pairKey{GroupID: groupID, First: debtorID, Second: creditorID}
	sign := int64(1)
	if key.First > key.Second {
		key.First, key.Second = key.Second, key.First
		sign = -1
	}
	return sign * edges[key]
}

func settlementDescription(paymentMethod string) string {
	switch paymentMethod {
	case "venmo":
		return "Paid via Venmo"
	case "paypal":
		return "Paid via PayPal"
	case "cashApp":
		return "Paid via Cash App"
	case "zelle":
		return "Paid via Zelle"
	case "other":
		return "Payment settled"
	default:
		return "Payment settled"
	}
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

	friendIDs := make(map[string]struct{}, len(snapshot.Friends))
	for _, f := range snapshot.Friends {
		friendIDs[f.ID] = struct{}{}
	}

	friends := make([]FriendBalance, 0, len(users)-1)
	for userID, user := range users {
		if userID == snapshot.CurrentUser.ID {
			continue
		}
		// A friend with no shared group must serialize as [] rather than null,
		// because the API contract declares groupNames as a string array.
		names := groupNames[userID]
		if names == nil {
			names = []string{}
		}
		_, isFriend := friendIDs[userID]
		friends = append(friends, FriendBalance{User: user, GroupNames: names, Balance: Money{AmountMinor: balances[userID], Currency: currency}, IsFriend: isFriend})
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

// buildGroupBalances attributes each friend-level net onto shared groups using
// remaining pairwise edges. That way a group never shows "You owe" for someone
// who already nets positive with you overall (OweNone's cross-group model).
func buildGroupBalances(currentUserID string, groups []Group, edges map[pairKey]int64, friendBalances map[string]int64, currency string) []GroupBalance {
	balances := make(map[string]int64, len(groups))

	type groupShare struct {
		groupID string
		amount  int64
	}

	allocate := func(counterpartyID string, friendNet int64) {
		if friendNet == 0 {
			return
		}
		remaining := abs(friendNet)
		shares := make([]groupShare, 0)
		for _, group := range groups {
			if !groupContains(group, currentUserID) || !groupContains(group, counterpartyID) {
				continue
			}
			var pairwise int64
			if friendNet < 0 {
				// Current user still owes this person overall — use group edges where we owe them.
				pairwise = amountOwed(edges, group.ID, currentUserID, counterpartyID)
			} else {
				// They still owe the current user overall.
				pairwise = amountOwed(edges, group.ID, counterpartyID, currentUserID)
			}
			if pairwise > 0 {
				shares = append(shares, groupShare{groupID: group.ID, amount: pairwise})
			}
		}
		slices.SortFunc(shares, func(left, right groupShare) int {
			if left.amount != right.amount {
				return cmp.Compare(right.amount, left.amount)
			}
			return cmp.Compare(left.groupID, right.groupID)
		})

		for _, share := range shares {
			if remaining == 0 {
				break
			}
			applied := share.amount
			if applied > remaining {
				applied = remaining
			}
			if friendNet < 0 {
				balances[share.groupID] -= applied
			} else {
				balances[share.groupID] += applied
			}
			remaining -= applied
		}

		// If edges were already cleared by cross-group settlement but the friend
		// net remains, fall back to the first shared group so totals still match.
		if remaining > 0 {
			for _, group := range groups {
				if !groupContains(group, currentUserID) || !groupContains(group, counterpartyID) {
					continue
				}
				if friendNet < 0 {
					balances[group.ID] -= remaining
				} else {
					balances[group.ID] += remaining
				}
				break
			}
		}
	}

	for counterpartyID, friendNet := range friendBalances {
		if counterpartyID == currentUserID || friendNet == 0 {
			continue
		}
		allocate(counterpartyID, friendNet)
	}

	result := make([]GroupBalance, 0, len(groups))
	for _, group := range groups {
		result = append(result, GroupBalance{
			ID: group.ID, Name: group.Name, Icon: group.Icon, Members: group.Members,
			Balance: Money{AmountMinor: balances[group.ID], Currency: currency},
			IsOwner: group.OwnerID == currentUserID,
		})
	}
	slices.SortFunc(result, func(left, right GroupBalance) int {
		return cmp.Compare(abs(right.Balance.AmountMinor), abs(left.Balance.AmountMinor))
	})
	return result
}

func buildRecentActivity(activities []Activity, now time.Time) []Activity {
	cutoff := now.AddDate(0, 0, -7)
	result := make([]Activity, 0, len(activities))
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
