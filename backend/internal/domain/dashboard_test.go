package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestBuildDashboardCompressesAcrossGroups(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	jordan := User{ID: "jordan", DisplayName: "Jordan"}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, alex, jordan}}
	flat := Group{ID: "flat", Name: "Flat", Members: []User{sam, alex, jordan}}

	dashboard, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip, flat},
		Expenses: []Expense{
			{
				ID: "dinner", GroupID: trip.ID, Description: "Dinner", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 1200, Currency: "GBP"}, SplitMethod: "exact", CreatedAt: now.Add(-time.Hour),
				Splits: []ExpenseSplit{{UserID: sam.ID, AmountMinor: 1200}},
			},
			{
				ID: "rent", GroupID: flat.ID, Description: "Rent", PayerID: sam.ID, CreatedByID: sam.ID,
				Money: Money{AmountMinor: 3000, Currency: "GBP"}, SplitMethod: "exact", CreatedAt: now.Add(-2 * time.Hour),
				Splits: []ExpenseSplit{{UserID: jordan.ID, AmountMinor: 3000}},
			},
			{
				ID: "tickets", GroupID: trip.ID, Description: "Tickets", PayerID: jordan.ID, CreatedByID: jordan.ID,
				Money: Money{AmountMinor: 1800, Currency: "GBP"}, SplitMethod: "exact", CreatedAt: now.Add(-3 * time.Hour),
				Splits: []ExpenseSplit{{UserID: alex.ID, AmountMinor: 1800}},
			},
		},
	}, now)
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}

	if dashboard.Summary.NetBalance.AmountMinor != 1800 {
		t.Errorf("net balance = %d, want 1800", dashboard.Summary.NetBalance.AmountMinor)
	}
	if dashboard.Suggestion.OriginalPaymentCount != 3 || dashboard.Suggestion.ReducedPaymentCount != 2 {
		t.Errorf("payment counts = %d -> %d, want 3 -> 2", dashboard.Suggestion.OriginalPaymentCount, dashboard.Suggestion.ReducedPaymentCount)
	}
	if dashboard.Suggestion.OffsettingDebt.AmountMinor != 4200 {
		t.Errorf("offsetting debt = %d, want 4200", dashboard.Suggestion.OffsettingDebt.AmountMinor)
	}
	if len(dashboard.Activity) != 3 {
		t.Errorf("activity count = %d, want 3", len(dashboard.Activity))
	}
	assertNamedTransfersConservePositions(t, map[string]int64{"sam": 1800, "alex": -600, "jordan": -1200}, dashboard.Suggestion.Transfers)
}

func TestBuildDashboardCrossGroupSettlementClearsGroupBalances(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "USD"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, alex}}
	flat := Group{ID: "flat", Name: "Flat", Members: []User{sam, alex}}

	before, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip, flat},
		Expenses: []Expense{
			{
				ID: "dinner", GroupID: trip.ID, Description: "Dinner", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 2000, Currency: "USD"}, SplitMethod: "exact", CreatedAt: now.Add(-2 * time.Hour),
				Splits: []ExpenseSplit{{UserID: sam.ID, AmountMinor: 2000}},
			},
			{
				ID: "utilities", GroupID: flat.ID, Description: "Utilities", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 1000, Currency: "USD"}, SplitMethod: "exact", CreatedAt: now.Add(-time.Hour),
				Splits: []ExpenseSplit{{UserID: sam.ID, AmountMinor: 1000}},
			},
		},
	}, now)
	if err != nil {
		t.Fatalf("BuildDashboard() before error = %v", err)
	}
	byName := map[string]int64{}
	for _, group := range before.Groups {
		byName[group.Name] = group.Balance.AmountMinor
	}
	if byName["Trip"] != -2000 || byName["Flat"] != -1000 {
		t.Fatalf("before balances trip=%d flat=%d, want -2000 and -1000", byName["Trip"], byName["Flat"])
	}

	after, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip, flat},
		Expenses: []Expense{
			{
				ID: "dinner", GroupID: trip.ID, Description: "Dinner", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 2000, Currency: "USD"}, SplitMethod: "exact", CreatedAt: now.Add(-2 * time.Hour),
				Splits: []ExpenseSplit{{UserID: sam.ID, AmountMinor: 2000}},
			},
			{
				ID: "utilities", GroupID: flat.ID, Description: "Utilities", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 1000, Currency: "USD"}, SplitMethod: "exact", CreatedAt: now.Add(-time.Hour),
				Splits: []ExpenseSplit{{UserID: sam.ID, AmountMinor: 1000}},
			},
		},
		Settlements: []Settlement{{
			ID: "pay-alex", FromUserID: sam.ID, ToUserID: alex.ID,
			Money: Money{AmountMinor: 3000, Currency: "USD"}, CreatedAt: now,
		}},
	}, now)
	if err != nil {
		t.Fatalf("BuildDashboard() after error = %v", err)
	}
	for _, group := range after.Groups {
		if group.Balance.AmountMinor != 0 {
			t.Errorf("group %s balance = %d, want 0 after cross-group settlement", group.Name, group.Balance.AmountMinor)
		}
	}
	for _, friend := range after.Friends {
		if friend.User.ID == alex.ID && friend.Balance.AmountMinor != 0 {
			t.Errorf("friend alex balance = %d, want 0", friend.Balance.AmountMinor)
		}
	}
}

func TestBuildDashboardGroupBalancesFollowFriendNets(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "USD"}
	mike := User{ID: "mike", DisplayName: "Mike"}
	flat := Group{ID: "flat", Name: "Flat", Members: []User{sam, mike}}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, mike}}

	dashboard, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{flat, trip},
		Expenses: []Expense{
			{
				ID: "rent", GroupID: flat.ID, Description: "Rent", PayerID: mike.ID, CreatedByID: mike.ID,
				Money: Money{AmountMinor: 3900, Currency: "USD"}, SplitMethod: "exact", CreatedAt: now.Add(-2 * time.Hour),
				Splits: []ExpenseSplit{{UserID: sam.ID, AmountMinor: 3900}},
			},
			{
				ID: "flights", GroupID: trip.ID, Description: "Flights", PayerID: sam.ID, CreatedByID: sam.ID,
				Money: Money{AmountMinor: 9500, Currency: "USD"}, SplitMethod: "exact", CreatedAt: now.Add(-time.Hour),
				Splits: []ExpenseSplit{{UserID: mike.ID, AmountMinor: 9500}},
			},
		},
	}, now)
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}

	var mikeFriend int64
	for _, friend := range dashboard.Friends {
		if friend.User.ID == mike.ID {
			mikeFriend = friend.Balance.AmountMinor
		}
	}
	if mikeFriend != 5600 {
		t.Fatalf("mike friend balance = %d, want 5600", mikeFriend)
	}

	byName := map[string]int64{}
	for _, group := range dashboard.Groups {
		byName[group.Name] = group.Balance.AmountMinor
	}
	// Globally Mike owes Sam, so Flat must not keep showing "You owe" from the rent IOU.
	if byName["Flat"] != 0 {
		t.Errorf("Flat balance = %d, want 0 after friend-level netting", byName["Flat"])
	}
	if byName["Trip"] != 5600 {
		t.Errorf("Trip balance = %d, want 5600", byName["Trip"])
	}
}

func TestBuildDashboardRejectsInvalidExpenseSplits(t *testing.T) {
	user := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	_, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: user,
		Expenses:    []Expense{{ID: "expense", PayerID: user.ID, Money: Money{AmountMinor: 1000, Currency: "GBP"}, Splits: []ExpenseSplit{{UserID: user.ID, AmountMinor: 900}}}},
	}, time.Now())
	if err == nil {
		t.Fatal("BuildDashboard() error = nil, want invalid split error")
	}
}

func TestBuildDashboardSerializesEmptyCollectionsAsArrays(t *testing.T) {
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	// A friend known through the friendship list but sharing no group, and no
	// expenses at all, previously produced null for groupNames and netPositions.
	dashboard, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: sam,
		Friends:     []User{{ID: "alex", DisplayName: "Alex"}},
	}, time.Now())
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}

	encoded, err := json.Marshal(dashboard)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, field := range []string{"friends", "groups", "activity", "netPositions", "groupNames", "transfers"} {
		if bytes.Contains(encoded, []byte(`"`+field+`":null`)) {
			t.Errorf("%s serialized as null, want []", field)
		}
	}
}

func TestBuildDashboardIsFriendDistinguishesRealFriendsFromGroupOnly(t *testing.T) {
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alice := User{ID: "alice", DisplayName: "Alice"}
	bob := User{ID: "bob", DisplayName: "Bob"}
	group := Group{ID: "g1", Name: "Group", Members: []User{sam, bob}}

	dashboard, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: sam,
		Friends:     []User{alice},
		Groups:      []Group{group},
	}, time.Now())
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}

	byID := make(map[string]FriendBalance)
	for _, fb := range dashboard.Friends {
		byID[fb.User.ID] = fb
	}
	if !byID["alice"].IsFriend {
		t.Errorf("alice IsFriend = false, want true (alice is in snapshot.Friends)")
	}
	if byID["bob"].IsFriend {
		t.Errorf("bob IsFriend = true, want false (bob is group-only)")
	}
}

func TestBuildDashboardGroupIsOwnerFlag(t *testing.T) {
	owner := User{ID: "owner-id", DisplayName: "Owner", PreferredCurrency: "GBP"}
	member := User{ID: "member-id", DisplayName: "Member"}
	group := Group{ID: "g1", Name: "Trip", Icon: "✈️", Members: []User{owner, member}, OwnerID: owner.ID}

	ownerDashboard, err := BuildDashboard(LedgerSnapshot{CurrentUser: owner, Groups: []Group{group}}, time.Now())
	if err != nil {
		t.Fatalf("BuildDashboard for owner: %v", err)
	}
	memberDashboard, err := BuildDashboard(LedgerSnapshot{
		CurrentUser: User{ID: member.ID, DisplayName: member.DisplayName, PreferredCurrency: "GBP"},
		Groups:      []Group{group},
	}, time.Now())
	if err != nil {
		t.Fatalf("BuildDashboard for member: %v", err)
	}

	if len(ownerDashboard.Groups) == 0 {
		t.Fatal("owner dashboard has no groups")
	}
	if !ownerDashboard.Groups[0].IsOwner {
		t.Errorf("owner IsOwner = false, want true")
	}
	if len(memberDashboard.Groups) == 0 {
		t.Fatal("member dashboard has no groups")
	}
	if memberDashboard.Groups[0].IsOwner {
		t.Errorf("member IsOwner = true, want false")
	}
}

func assertNamedTransfersConservePositions(t *testing.T, positions map[string]int64, transfers []NamedTransfer) {
	t.Helper()
	for _, transfer := range transfers {
		positions[transfer.From.ID] += transfer.Money.AmountMinor
		positions[transfer.To.ID] -= transfer.Money.AmountMinor
	}
	for userID, amount := range positions {
		if amount != 0 {
			t.Errorf("remaining position for %s = %d, want 0", userID, amount)
		}
	}
}

func TestBuildActivityFeedIncludesItemsOlderThan7Days(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, alex}}

	snapshot := LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip},
		Expenses: []Expense{
			{
				ID: "recent", GroupID: trip.ID, Description: "Recent dinner", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 1000, Currency: "GBP"}, SplitMethod: "exact",
				CreatedAt: now.Add(-2 * 24 * time.Hour),
				Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 1000}},
			},
			{
				ID: "old", GroupID: trip.ID, Description: "Old lunch", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 500, Currency: "GBP"}, SplitMethod: "exact",
				CreatedAt: now.Add(-30 * 24 * time.Hour),
				Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 500}},
			},
		},
	}

	feed, err := BuildActivityFeed(snapshot, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	if len(feed) != 2 {
		t.Errorf("feed length = %d, want 2 (including item older than 7 days)", len(feed))
	}

	dashboard, err := BuildDashboard(snapshot, now)
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}
	if len(dashboard.Activity) != 1 {
		t.Errorf("dashboard.Activity length = %d, want 1 (7-day cap still applies)", len(dashboard.Activity))
	}
}

func TestBuildActivityFeedExceedsThe20ItemCap(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, alex}}

	expenses := make([]Expense, 25)
	for i := range expenses {
		expenses[i] = Expense{
			ID: fmt.Sprintf("expense-%d", i), GroupID: trip.ID, Description: fmt.Sprintf("Expense %d", i),
			PayerID: alex.ID, CreatedByID: alex.ID,
			Money: Money{AmountMinor: 100, Currency: "GBP"}, SplitMethod: "exact",
			CreatedAt: now.Add(-time.Duration(i+1) * time.Hour),
			Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 100}},
		}
	}

	snapshot := LedgerSnapshot{CurrentUser: sam, Groups: []Group{trip}, Expenses: expenses}

	feed, err := BuildActivityFeed(snapshot, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	if len(feed) != 25 {
		t.Errorf("feed length = %d, want 25 (no 20-item cap on feed)", len(feed))
	}

	dashboard, err := BuildDashboard(snapshot, now)
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}
	if len(dashboard.Activity) != 20 {
		t.Errorf("dashboard.Activity length = %d, want 20 (20-item cap still applies to dashboard)", len(dashboard.Activity))
	}
}

func TestBuildActivityFeedIsSortedNewestFirst(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, alex}}

	// Expenses deliberately out of chronological order.
	snapshot := LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip},
		Expenses: []Expense{
			{
				ID: "oldest", GroupID: trip.ID, Description: "Oldest", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 100, Currency: "GBP"}, SplitMethod: "exact",
				CreatedAt: now.Add(-3 * time.Hour),
				Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 100}},
			},
			{
				ID: "newest", GroupID: trip.ID, Description: "Newest", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 200, Currency: "GBP"}, SplitMethod: "exact",
				CreatedAt: now.Add(-1 * time.Hour),
				Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 200}},
			},
			{
				ID: "middle", GroupID: trip.ID, Description: "Middle", PayerID: alex.ID, CreatedByID: alex.ID,
				Money: Money{AmountMinor: 150, Currency: "GBP"}, SplitMethod: "exact",
				CreatedAt: now.Add(-2 * time.Hour),
				Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 150}},
			},
		},
	}

	feed, err := BuildActivityFeed(snapshot, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	if len(feed) != 3 {
		t.Fatalf("feed length = %d, want 3", len(feed))
	}
	if feed[0].ID != "newest" || feed[1].ID != "middle" || feed[2].ID != "oldest" {
		t.Errorf("feed order = [%s, %s, %s], want [newest, middle, oldest]",
			feed[0].ID, feed[1].ID, feed[2].ID)
	}
}

func TestBuildActivityFeedItemMatchesDashboardItem(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip", Name: "Trip", Members: []User{sam, alex}}

	snapshot := LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip},
		Expenses: []Expense{{
			ID: "shared", GroupID: trip.ID, Description: "Shared dinner", PayerID: alex.ID, CreatedByID: alex.ID,
			Money: Money{AmountMinor: 1000, Currency: "GBP"}, SplitMethod: "equal",
			CreatedAt: now.Add(-time.Hour),
			Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 1000}},
		}},
	}

	feed, err := BuildActivityFeed(snapshot, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	dashboard, err := BuildDashboard(snapshot, now)
	if err != nil {
		t.Fatalf("BuildDashboard() error = %v", err)
	}

	if len(feed) == 0 || len(dashboard.Activity) == 0 {
		t.Fatal("both feed and dashboard must contain at least one activity")
	}

	fi, di := feed[0], dashboard.Activity[0]
	feedJSON, _ := json.Marshal(fi)
	dashJSON, _ := json.Marshal(di)
	if !bytes.Equal(feedJSON, dashJSON) {
		t.Errorf("item in feed and dashboard are not identical:\n  feed=%s\n  dash=%s", feedJSON, dashJSON)
	}
}

func TestActivityGroupIDPopulatedFromExpense(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip-id", Name: "Trip", Members: []User{sam, alex}}

	feed, err := BuildActivityFeed(LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip},
		Expenses: []Expense{{
			ID: "e1", GroupID: trip.ID, Description: "Dinner", PayerID: alex.ID, CreatedByID: alex.ID,
			Money: Money{AmountMinor: 1000, Currency: "GBP"}, SplitMethod: "exact",
			CreatedAt: now.Add(-time.Hour),
			Splits:    []ExpenseSplit{{UserID: sam.ID, AmountMinor: 1000}},
		}},
	}, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	if len(feed) == 0 {
		t.Fatal("feed is empty")
	}
	if feed[0].GroupID != trip.ID {
		t.Errorf("GroupID = %q, want %q", feed[0].GroupID, trip.ID)
	}
}

func TestActivityGroupIDEmptyForNilGroupSettlement(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}

	feed, err := BuildActivityFeed(LedgerSnapshot{
		CurrentUser: sam,
		Friends:     []User{alex},
		Settlements: []Settlement{{
			ID:         "s1",
			GroupID:    nil, // cross-group settlement must not panic
			FromUserID: alex.ID, ToUserID: sam.ID,
			Money:     Money{AmountMinor: 500, Currency: "GBP"},
			CreatedAt: now.Add(-time.Hour),
		}},
	}, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	if len(feed) == 0 {
		t.Fatal("feed is empty")
	}
	if feed[0].GroupID != "" {
		t.Errorf("GroupID = %q, want empty for nil-group settlement", feed[0].GroupID)
	}
}

func TestActivityGroupIDPopulatedFromGroupSettlement(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	sam := User{ID: "sam", DisplayName: "Sam", PreferredCurrency: "GBP"}
	alex := User{ID: "alex", DisplayName: "Alex"}
	trip := Group{ID: "trip-id", Name: "Trip", Members: []User{sam, alex}}
	tripID := trip.ID

	feed, err := BuildActivityFeed(LedgerSnapshot{
		CurrentUser: sam,
		Groups:      []Group{trip},
		Settlements: []Settlement{{
			ID:         "s1",
			GroupID:    &tripID,
			FromUserID: sam.ID, ToUserID: alex.ID,
			Money:     Money{AmountMinor: 500, Currency: "GBP"},
			CreatedAt: now.Add(-time.Hour),
		}},
	}, now)
	if err != nil {
		t.Fatalf("BuildActivityFeed() error = %v", err)
	}
	if len(feed) == 0 {
		t.Fatal("feed is empty")
	}
	if feed[0].GroupID != trip.ID {
		t.Errorf("GroupID = %q, want %q", feed[0].GroupID, trip.ID)
	}
}
