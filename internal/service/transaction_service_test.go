package service

import (
	"context"
	"testing"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/internal/repository"
)

func TestTransactionServiceImpl_CreditDebitTransfer(t *testing.T) {
	conn := getTestConn(t)
	txRepo := repository.NewTransactionPostgresRepository(conn)
	balRepo := repository.NewBalancePostgresRepository(conn)
	service := NewTransactionService(txRepo, balRepo)
	defer func() {
		conn.Exec(context.Background(), "DELETE FROM transactions WHERE from_user_id IN (8881,8882) OR to_user_id IN (8881,8882)")
		conn.Exec(context.Background(), "DELETE FROM balances WHERE user_id IN (8881,8882)")
		conn.Exec(context.Background(), "DELETE FROM users WHERE id IN (8881,8882)")
		_ = conn.Close(context.Background())
	}()

	// Create two test users
	u1 := &domain.User{ID: 8881, Username: "svcuser1", Email: "svcuser1@example.com", PasswordHash: "hash", Role: "user"}
	_, err := conn.Exec(context.Background(), "INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW())	ON CONFLICT (id) DO NOTHING", u1.ID, u1.Username, u1.Email, u1.PasswordHash, u1.Role)
	if err != nil {
		t.Fatalf("Failed to insert user1: %v", err)
	}

	row := conn.QueryRow(context.Background(), "SELECT id FROM users WHERE id = $1", u1.ID)
	var id int
	if err := row.Scan(&id); err != nil {
		t.Fatalf("User1 not found after insert: %v", err)
	}

	u2 := &domain.User{ID: 8882, Username: "svcuser2", Email: "svcuser2@example.com", PasswordHash: "hash", Role: "user"}
	_, err = conn.Exec(context.Background(), "INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW()) ON CONFLICT (id) DO NOTHING", u2.ID, u2.Username, u2.Email, u2.PasswordHash, u2.Role)
	if err != nil {
		t.Fatalf("Failed to insert user2: %v", err)
	}

	row = conn.QueryRow(context.Background(), "SELECT id FROM users WHERE id = $1", u2.ID)
	if err := row.Scan(&id); err != nil {
		t.Fatalf("User2 not found after insert: %v", err)
	}

	// Test Credit
	err = service.Credit(u1.ID, 200.0)
	if err != nil {
		t.Fatalf("Credit failed: %v", err)
	}
	bal, err := balRepo.GetByUserID(u1.ID)
	if err != nil || bal == nil || bal.Amount != 200.0 {
		t.Errorf("Credit: got balance %+v, want 200.0", bal)
	}

	// Test Debit
	err = service.Debit(u1.ID, 50.0)
	if err != nil {
		t.Fatalf("Debit failed: %v", err)
	}
	bal, _ = balRepo.GetByUserID(u1.ID)
	if bal.Amount != 150.0 {
		t.Errorf("Debit: got balance %+v, want 150.0", bal)
	}

	// Test Transfer
	err = service.Transfer(u1.ID, u2.ID, 100.0)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	bal1, _ := balRepo.GetByUserID(u1.ID)
	bal2, _ := balRepo.GetByUserID(u2.ID)
	if bal1.Amount != 50.0 || bal2.Amount != 100.0 {
		t.Errorf("Transfer: got balances %v, %v; want 50.0, 100.0", bal1.Amount, bal2.Amount)
	}

	// Test ListUserTransactions
	txs, err := service.ListUserTransactions(u1.ID)
	if err != nil {
		t.Fatalf("ListUserTransactions failed: %v", err)
	}
	if len(txs) < 3 {
		t.Errorf("ListUserTransactions: expected at least 3 transactions, got %d", len(txs))
	}
}
