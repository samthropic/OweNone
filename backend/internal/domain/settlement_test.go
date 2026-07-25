package domain

import "testing"

func TestMinimizeTransfers(t *testing.T) {
	positions := []Position{
		{UserID: "alex", AmountMinor: -600},
		{UserID: "jordan", AmountMinor: -400},
		{UserID: "mike", AmountMinor: 400},
		{UserID: "priya", AmountMinor: 600},
	}

	transfers, err := MinimizeTransfers(positions, "GBP")
	if err != nil {
		t.Fatalf("MinimizeTransfers() error = %v", err)
	}
	if len(transfers) != 2 {
		t.Fatalf("MinimizeTransfers() returned %d transfers, want 2: %#v", len(transfers), transfers)
	}

	assertConservesPositions(t, positions, transfers)
}

func TestMinimizeTransfersRejectsUnbalancedPositions(t *testing.T) {
	_, err := MinimizeTransfers([]Position{
		{UserID: "alex", AmountMinor: -500},
		{UserID: "jordan", AmountMinor: 400},
	}, "GBP")
	if err == nil {
		t.Fatal("MinimizeTransfers() error = nil, want unbalanced positions error")
	}
}

func TestMinimizeTransfersCombinesDuplicatePositions(t *testing.T) {
	transfers, err := MinimizeTransfers([]Position{
		{UserID: "alex", AmountMinor: -400},
		{UserID: "alex", AmountMinor: -200},
		{UserID: "jordan", AmountMinor: 600},
	}, "GBP")
	if err != nil {
		t.Fatalf("MinimizeTransfers() error = %v", err)
	}
	if len(transfers) != 1 || transfers[0].Money.AmountMinor != 600 {
		t.Fatalf("MinimizeTransfers() = %#v, want one transfer of 600", transfers)
	}
}

func assertConservesPositions(t *testing.T, positions []Position, transfers []Transfer) {
	t.Helper()

	remaining := make(map[string]int64, len(positions))
	for _, position := range positions {
		remaining[position.UserID] += position.AmountMinor
	}
	for _, transfer := range transfers {
		remaining[transfer.FromUserID] += transfer.Money.AmountMinor
		remaining[transfer.ToUserID] -= transfer.Money.AmountMinor
	}
	for userID, amount := range remaining {
		if amount != 0 {
			t.Errorf("remaining position for %s = %d, want 0", userID, amount)
		}
	}
}
