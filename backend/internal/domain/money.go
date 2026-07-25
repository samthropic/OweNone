package domain

import "fmt"

type Money struct {
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}

func NewMoney(amountMinor int64, currency string) (Money, error) {
	if len(currency) != 3 {
		return Money{}, fmt.Errorf("currency must be a three-letter ISO 4217 code")
	}

	return Money{AmountMinor: amountMinor, Currency: currency}, nil
}
