package domain

import (
	"bytes"
	"encoding/json"
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
