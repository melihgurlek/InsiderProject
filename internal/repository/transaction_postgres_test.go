package repository

import (
	"context"
	"testing"

	"github.com/melihgurlek/backend-path/internal/domain"
)

func TestTransactionPostgresRepository_CRUD(t *testing.T) {
	conn := getTestConn(t)
	repo := NewTransactionPostgresRepository(conn)
	defer func() {
		_, _ = conn.Exec(context.Background(), "DELETE FROM transactions WHERE from_user_id IN (9991,9992) OR to_user_id IN (9991,9992)")
		_, _ = conn.Exec(context.Background(), "DELETE FROM users WHERE id IN (9991,9992)")
		_ = conn.Close(context.Background())
	}()

	// Create two test users
	u1 := &domain.User{ID: 9991, Username: "txuser1", Email: "txuser1@example.com", PasswordHash: "hash", Role: "user"}
	u2 := &domain.User{ID: 9992, Username: "txuser2", Email: "txuser2@example.com", PasswordHash: "hash", Role: "user"}
	_, _ = conn.Exec(context.Background(), "INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW()) ON CONFLICT (id) DO NOTHING", u1.ID, u1.Username, u1.Email, u1.PasswordHash, u1.Role)
	_, _ = conn.Exec(context.Background(), "INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW()) ON CONFLICT (id) DO NOTHING", u2.ID, u2.Username, u2.Email, u2.PasswordHash, u2.Role)

	// Test Create
	tx := &domain.Transaction{
		FromUserID: &u1.ID,
		ToUserID:   &u2.ID,
		Amount:     100.0,
		Type:       "transfer",
		Status:     "completed",
	}
	err := repo.Create(tx)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if tx.ID == 0 {
		t.Error("expected transaction ID to be set")
	}

	// Test GetByID
	got, err := repo.GetByID(tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got == nil || got.Amount != 100.0 {
		t.Errorf("GetByID: got %+v, want amount=100.0", got)
	}

	// Test ListByUser
	txs, err := repo.ListByUser(u1.ID)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	found := false
	for _, t := range txs {
		if t.ID == tx.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("ListByUser: transaction not found")
	}

	// Test ListAll
	all, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	found = false
	for _, t := range all {
		if t.ID == tx.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("ListAll: transaction not found")
	}
}
