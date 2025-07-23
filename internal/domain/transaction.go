package domain

import (
	"errors"
	"time"
)

// Transaction represents a money transfer or operation.
type Transaction struct {
	ID         int
	FromUserID *int
	ToUserID   *int
	Amount     float64
	Type       string // credit, debit, transfer
	Status     string // pending, completed, failed
	CreatedAt  time.Time
}

// Validate checks if the transaction fields are valid.
func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if t.Type != "credit" && t.Type != "debit" && t.Type != "transfer" {
		return errors.New("invalid transaction type")
	}
	if t.Status == "" {
		return errors.New("status is required")
	}
	return nil
}
