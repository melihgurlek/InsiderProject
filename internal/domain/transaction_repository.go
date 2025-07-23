package domain

// TransactionRepository defines methods for transaction data access.
type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByID(id int) (*Transaction, error)
	ListByUser(userID int) ([]*Transaction, error)
	ListAll() ([]*Transaction, error)
}
