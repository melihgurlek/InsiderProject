package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// getTestConn returns a pgxpool.Pool for testing, using the DB_URL env var or a default.
func getTestConn(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/backend_path?sslmode=disable"
	}
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("failed to parse db config: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	return pool
}

func TestBalancePostgresRepository_GetHistoricalBalance(t *testing.T) {
	conn := getTestConn(t)
	repo := NewBalancePostgresRepository(conn)
	userID := 7771
	defer func() {
		conn.Exec(context.Background(), "DELETE FROM transactions WHERE from_user_id = $1 OR to_user_id = $1", userID)
		conn.Exec(context.Background(), "DELETE FROM balances WHERE user_id = $1", userID)
		conn.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
		conn.Close()
	}()

	// Insert test user
	_, _ = conn.Exec(context.Background(), "INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW()) ON CONFLICT (id) DO NOTHING", userID, "balhistuser", "balhistuser@example.com", "hash", "user")

	// Insert transactions on different days
	daysAgo := func(n int) time.Time {
		return time.Now().Truncate(24*time.Hour).AddDate(0, 0, -n)
	}
	// Day -3: credit 100
	tx1 := &domain.Transaction{
		FromUserID: nil,
		ToUserID:   &userID,
		Amount:     100.0,
		Type:       "credit",
		Status:     "completed",
		CreatedAt:  daysAgo(3),
	}
	// Day -2: debit 40
	tx2 := &domain.Transaction{
		FromUserID: &userID,
		ToUserID:   nil,
		Amount:     40.0,
		Type:       "debit",
		Status:     "completed",
		CreatedAt:  daysAgo(2),
	}
	// Day -1: credit 60
	tx3 := &domain.Transaction{
		FromUserID: nil,
		ToUserID:   &userID,
		Amount:     60.0,
		Type:       "credit",
		Status:     "completed",
		CreatedAt:  daysAgo(1),
	}
	conn.Exec(context.Background(), "INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)", tx1.FromUserID, tx1.ToUserID, tx1.Amount, tx1.Type, tx1.Status, tx1.CreatedAt)
	conn.Exec(context.Background(), "INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)", tx2.FromUserID, tx2.ToUserID, tx2.Amount, tx2.Type, tx2.Status, tx2.CreatedAt)
	conn.Exec(context.Background(), "INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)", tx3.FromUserID, tx3.ToUserID, tx3.Amount, tx3.Type, tx3.Status, tx3.CreatedAt)

	// Call GetHistoricalBalance
	balances, err := repo.GetHistoricalBalance(userID, 7771)
	if err != nil {
		t.Fatalf("GetHistoricalBalance failed: %v", err)
	}
	if len(balances) == 0 {
		t.Fatalf("expected non-empty balance history")
	}

	// Find balances for the last 4 days
	var bDay3, bDay2, bDay1, bDay0 *domain.Balance
	for _, b := range balances {
		delta := int(time.Now().Truncate(24*time.Hour).Sub(b.LastUpdatedAt).Hours() / 24)
		switch delta {
		case 3:
			bDay3 = b
		case 2:
			bDay2 = b
		case 1:
			bDay1 = b
		case 0:
			bDay0 = b
		}
	}
	if bDay3 == nil || bDay2 == nil || bDay1 == nil || bDay0 == nil {
		t.Errorf("missing expected days in balance history")
	}
	if bDay3 != nil && bDay3.Amount != 100.0 {
		t.Errorf("day -3: got %.2f, want 100.0", bDay3.Amount)
	}
	if bDay2 != nil && bDay2.Amount != 60.0 {
		t.Errorf("day -2: got %.2f, want 60.0", bDay2.Amount)
	}
	if bDay1 != nil && bDay1.Amount != 120.0 {
		t.Errorf("day -1: got %.2f, want 120.0", bDay1.Amount)
	}
	if bDay0 != nil && bDay0.Amount != 120.0 {
		t.Errorf("day 0: got %.2f, want 120.0", bDay0.Amount)
	}
}
