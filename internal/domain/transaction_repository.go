package domain

import (
	"context"
	"time"
)

// TransactionRepository defines methods for transaction data access.
type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByID(id int) (*Transaction, error)
	ListByUser(userID int) ([]*Transaction, error)
	ListByUserAndTimeRange(userID int, from, to time.Time) ([]*Transaction, error)
	ListAll(ctx context.Context, limit int, offset int) ([]*Transaction, error)
}
