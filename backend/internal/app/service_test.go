package app

import (
	"errors"
	"testing"
)

func TestEqualSplitsDistributesRemainderDeterministically(t *testing.T) {
	participants := []string{
		"30000000-0000-0000-0000-000000000003",
		"10000000-0000-0000-0000-000000000001",
		"20000000-0000-0000-0000-000000000002",
	}

	splits, err := equalSplits(participants, 1000)
	if err != nil {
		t.Fatalf("equalSplits() error = %v", err)
	}
	wantAmounts := []int64{334, 333, 333}
	for index, split := range splits {
		if split.AmountMinor != wantAmounts[index] {
			t.Errorf("split %d amount = %d, want %d", index, split.AmountMinor, wantAmounts[index])
		}
	}
	if splits[0].UserID != "10000000-0000-0000-0000-000000000001" {
		t.Errorf("remainder recipient = %s, want lowest UUID", splits[0].UserID)
	}
}

func TestExactSplitsRejectsWrongTotal(t *testing.T) {
	_, err := exactSplits(nil, 100)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("exactSplits() error = %v, want ErrInvalidInput", err)
	}
}
