package service

import (
	"errors"

	"github.com/melihgurlek/backend-path/internal/domain"
)

// TransactionServiceImpl implements domain.TransactionService.
type TransactionServiceImpl struct {
	txRepo  domain.TransactionRepository
	balRepo domain.BalanceRepository
}

// NewTransactionService creates a new TransactionServiceImpl.
func NewTransactionService(txRepo domain.TransactionRepository, balRepo domain.BalanceRepository) *TransactionServiceImpl {
	return &TransactionServiceImpl{txRepo: txRepo, balRepo: balRepo}
}

// Credit adds amount to a user's balance and records a transaction.
func (s *TransactionServiceImpl) Credit(userID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	bal, err := s.balRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if bal == nil {
		bal = &domain.Balance{UserID: userID, Amount: 0}
	}
	bal.Amount += amount
	if err := s.balRepo.Update(bal); err != nil {
		return err
	}
	tx := &domain.Transaction{
		FromUserID: nil, // system
		ToUserID:   &userID,
		Amount:     amount,
		Type:       "credit",
		Status:     "completed",
	}
	return s.txRepo.Create(tx)
}

// Debit subtracts amount from a user's balance and records a transaction.
func (s *TransactionServiceImpl) Debit(userID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	bal, err := s.balRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if bal == nil || bal.Amount < amount {
		return errors.New("insufficient balance")
	}
	bal.Amount -= amount
	if err := s.balRepo.Update(bal); err != nil {
		return err
	}
	tx := &domain.Transaction{
		FromUserID: &userID,
		ToUserID:   nil, // system
		Amount:     amount,
		Type:       "debit",
		Status:     "completed",
	}
	return s.txRepo.Create(tx)
}

// Transfer moves amount from one user to another, updating balances and recording a transaction.
func (s *TransactionServiceImpl) Transfer(fromUserID, toUserID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if fromUserID == toUserID {
		return errors.New("cannot transfer to self")
	}
	fromBal, err := s.balRepo.GetByUserID(fromUserID)
	if err != nil {
		return err
	}
	if fromBal == nil || fromBal.Amount < amount {
		return errors.New("insufficient balance")
	}
	toBal, err := s.balRepo.GetByUserID(toUserID)
	if err != nil {
		return err
	}
	if toBal == nil {
		toBal = &domain.Balance{UserID: toUserID, Amount: 0}
	}
	fromBal.Amount -= amount
	toBal.Amount += amount
	if err := s.balRepo.Update(fromBal); err != nil {
		return err
	}
	if err := s.balRepo.Update(toBal); err != nil {
		return err
	}
	tx := &domain.Transaction{
		FromUserID: &fromUserID,
		ToUserID:   &toUserID,
		Amount:     amount,
		Type:       "transfer",
		Status:     "completed",
	}
	return s.txRepo.Create(tx)
}

// GetTransaction returns a transaction by ID.
func (s *TransactionServiceImpl) GetTransaction(id int) (*domain.Transaction, error) {
	return s.txRepo.GetByID(id)
}

// ListUserTransactions returns all transactions for a user.
func (s *TransactionServiceImpl) ListUserTransactions(userID int) ([]*domain.Transaction, error) {
	return s.txRepo.ListByUser(userID)
}

// ListAllTransactions returns all transactions.
func (s *TransactionServiceImpl) ListAllTransactions() ([]*domain.Transaction, error) {
	return s.txRepo.ListAll()
}
