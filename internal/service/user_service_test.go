package service

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/melihgurlek/backend-path/internal/repository"
)

// getTestConn returns a pgx.Conn for testing, using the DB_URL env var or a default.
func getTestConn(t *testing.T) *pgx.Conn {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/backend_path?sslmode=disable"
	}
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	return conn
}

func TestUserServiceImpl_RegisterAndLogin(t *testing.T) {
	conn := getTestConn(t)
	repo := repository.NewUserPostgresRepository(conn)
	service := NewUserService(repo)
	defer func() {
		conn.Exec(context.Background(), "DELETE FROM users WHERE username = 'servicetestuser'")
		_ = conn.Close(context.Background())
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
