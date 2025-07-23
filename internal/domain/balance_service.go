package domain

// BalanceService defines business logic for balances.
type BalanceService interface {
	GetCurrentBalance(userID int) (*Balance, error)
	GetHistoricalBalance(userID int) ([]*Balance, error)
}
