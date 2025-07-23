package domain

// BalanceRepository defines methods for balance data access.
type BalanceRepository interface {
	GetByUserID(userID int) (*Balance, error)
	Update(balance *Balance) error
}
