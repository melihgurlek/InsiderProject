package domain

// TransactionService defines business logic for transactions.
type TransactionService interface {
	Credit(userID int, amount float64) error
	Debit(userID int, amount float64) error
	Transfer(fromUserID, toUserID int, amount float64) error
	GetTransaction(id int) (*Transaction, error)
	ListUserTransactions(userID int) ([]*Transaction, error)
}
