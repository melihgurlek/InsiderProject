package service

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/melihgurlek/backend-path/internal/repository"
)

// getTestPool returns a pgxpool.Pool for testing, using the DB_URL env var or a default.
func getTestPool(t *testing.T) *pgxpool.Pool {
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

func TestUserServiceImpl_RegisterAndLogin(t *testing.T) {
	pool := getTestPool(t)
	repo := repository.NewUserPostgresRepository(pool) // This already implements domain.UserRepository
	service := NewUserService(repo)
	defer func() {
		pool.Exec(context.Background(), "DELETE FROM users WHERE username = 'servicetestuser'")
		pool.Close()
	}()

	// Test Register
	user, err := service.Register("servicetestuser", "servicetestuser@example.com", "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}

	// Test duplicate username
	_, err = service.Register("servicetestuser", "other@example.com", "password123")
	if err == nil {
		t.Error("expected error for duplicate username, got nil")
	}

	// Test duplicate email
	_, err = service.Register("otheruser", "servicetestuser@example.com", "password123")
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}

	// Test Login (correct password)
	loggedIn, err := service.Login("servicetestuser", "password123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loggedIn.ID != user.ID {
		t.Errorf("Login: got %+v, want %+v", loggedIn, user)
	}

	// Test Login (wrong password)
	_, err = service.Login("servicetestuser", "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}

	// Test Login (nonexistent user)
	_, err = service.Login("doesnotexist", "password123")
	if err == nil {
		t.Error("expected error for nonexistent user, got nil")
	}
}
