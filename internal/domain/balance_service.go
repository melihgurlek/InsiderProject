package domain

import "time"

// BalanceService defines business logic for balances.
type BalanceService interface {
	GetCurrentBalance(userID int) (*Balance, error)
	GetHistoricalBalance(userID int) ([]*Balance, error)
	GetBalanceAtTime(userID int, time time.Time) (*Balance, error)
}
