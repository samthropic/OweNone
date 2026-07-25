package domain

import (
	"cmp"
	"fmt"
	"slices"
)

type Position struct {
	UserID      string
	AmountMinor int64
}

type Transfer struct {
	FromUserID string `json:"fromUserId"`
	ToUserID   string `json:"toUserId"`
	Money      Money  `json:"money"`
}

type settlementState struct {
	positions []Position
	transfers []Transfer
}

// MinimizeTransfers returns the fewest transfers needed to settle the supplied
// net positions. Positive positions receive money; negative positions pay it.
func MinimizeTransfers(positions []Position, currency string) ([]Transfer, error) {
	if len(currency) != 3 {
		return nil, fmt.Errorf("currency must be a three-letter ISO 4217 code")
	}

	normalized := normalizePositions(positions)
	var total int64
	for _, position := range normalized {
		total += position.AmountMinor
	}
	if total != 0 {
		return nil, fmt.Errorf("positions must sum to zero, got %d", total)
	}

	best := settlementState{}
	searchSettlements(settlementState{positions: normalized}, currency, &best)
	return best.transfers, nil
}

func normalizePositions(positions []Position) []Position {
	byUser := make(map[string]int64, len(positions))
	for _, position := range positions {
		byUser[position.UserID] += position.AmountMinor
	}

	normalized := make([]Position, 0, len(byUser))
	for userID, amount := range byUser {
		if amount != 0 {
			normalized = append(normalized, Position{UserID: userID, AmountMinor: amount})
		}
	}
	slices.SortFunc(normalized, func(left, right Position) int {
		return cmp.Compare(left.UserID, right.UserID)
	})
	return normalized
}

func searchSettlements(current settlementState, currency string, best *settlementState) {
	first := firstUnsettled(current.positions)
	if first == -1 {
		if best.transfers == nil || len(current.transfers) < len(best.transfers) {
			best.transfers = slices.Clone(current.transfers)
		}
		return
	}
	if best.transfers != nil && len(current.transfers) >= len(best.transfers) {
		return
	}

	seenAmounts := make(map[int64]struct{})
	for other := first + 1; other < len(current.positions); other++ {
		if !oppositeSigns(current.positions[first].AmountMinor, current.positions[other].AmountMinor) {
			continue
		}
		if _, seen := seenAmounts[current.positions[other].AmountMinor]; seen {
			continue
		}
		seenAmounts[current.positions[other].AmountMinor] = struct{}{}

		nextPositions := slices.Clone(current.positions)
		amount := min(abs(nextPositions[first].AmountMinor), abs(nextPositions[other].AmountMinor))
		transfer := makeTransfer(nextPositions[first], nextPositions[other], amount, currency)

		if nextPositions[first].AmountMinor < 0 {
			nextPositions[first].AmountMinor += amount
			nextPositions[other].AmountMinor -= amount
		} else {
			nextPositions[first].AmountMinor -= amount
			nextPositions[other].AmountMinor += amount
		}

		searchSettlements(settlementState{
			positions: nextPositions,
			transfers: append(slices.Clone(current.transfers), transfer),
		}, currency, best)
	}
}

func firstUnsettled(positions []Position) int {
	for index, position := range positions {
		if position.AmountMinor != 0 {
			return index
		}
	}
	return -1
}

func oppositeSigns(left, right int64) bool {
	return (left < 0 && right > 0) || (left > 0 && right < 0)
}

func makeTransfer(left, right Position, amount int64, currency string) Transfer {
	if left.AmountMinor < 0 {
		return Transfer{FromUserID: left.UserID, ToUserID: right.UserID, Money: Money{AmountMinor: amount, Currency: currency}}
	}
	return Transfer{FromUserID: right.UserID, ToUserID: left.UserID, Money: Money{AmountMinor: amount, Currency: currency}}
}

func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
