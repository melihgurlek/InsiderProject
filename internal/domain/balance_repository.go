package domain

import "time"

// BalanceRepository defines methods for balance data access.
type BalanceRepository interface {
	GetByUserID(userID int) (*Balance, error)
	Update(balance *Balance) error
	GetHistoricalBalance(userID int, limit int) ([]*Balance, error)
	GetBalanceAtTime(userID int, t time.Time) (*Balance, error)
}
