package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
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

func TestUserPostgresRepository_CreateAndGet(t *testing.T) {
	conn := getTestConn(t)
	repo := NewUserPostgresRepository(conn)
	defer func() {
		// Clean up test user
		conn.Exec(context.Background(), "DELETE FROM users WHERE username = 'testuser'")
		_ = conn.Close(context.Background())
	}()

	user := &domain.User{
		Username:     "testuser",
		Email:        "testuser@example.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}

	// Test Create
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}

	// Test GetByID
	got, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got == nil || got.Username != user.Username {
		t.Errorf("GetByID: got %+v, want %+v", got, user)
	}

	// Test GetByUsername
	got, err = repo.GetByUsername("testuser")
	if err != nil {
		t.Fatalf("GetByUsername failed: %v", err)
	}
	if got == nil || got.Email != user.Email {
		t.Errorf("GetByUsername: got %+v, want %+v", got, user)
	}

	// Test GetByEmail
	got, err = repo.GetByEmail("testuser@example.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if got == nil || got.Username != user.Username {
		t.Errorf("GetByEmail: got %+v, want %+v", got, user)
	}
}

func TestUserPostgresRepository_UpdateDeleteList(t *testing.T) {
	conn := getTestConn(t)
	repo := NewUserPostgresRepository(conn)
	defer func() {
		conn.Exec(context.Background(), "DELETE FROM users WHERE username LIKE 'testuser%' OR username = 'updateduser'")
		_ = conn.Close(context.Background())
	}()

	// Create two users
	user1 := &domain.User{
		Username:     "testuser1",
		Email:        "testuser1@example.com",
		PasswordHash: "hash1",
		Role:         "user",
	}
	user2 := &domain.User{
		Username:     "testuser2",
		Email:        "testuser2@example.com",
		PasswordHash: "hash2",
		Role:         "user",
	}
	if err := repo.Create(user1); err != nil {
		t.Fatalf("Create user1 failed: %v", err)
	}
	if err := repo.Create(user2); err != nil {
		t.Fatalf("Create user2 failed: %v", err)
	}

	// Test Update
	user1.Username = "updateduser"
	user1.Email = "updateduser@example.com"
	user1.PasswordHash = "newhash"
	user1.Role = "admin"
	if err := repo.Update(user1); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	got, err := repo.GetByID(user1.ID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if got.Username != "updateduser" || got.Role != "admin" {
		t.Errorf("Update: got %+v, want username=updateduser, role=admin", got)
	}

	// Test List
	users, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	found1, found2 := false, false
	for _, u := range users {
		if u.ID == user1.ID {
			found1 = true
		}
		if u.ID == user2.ID {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("List: expected both users, found1=%v, found2=%v", found1, found2)
	}

	// Test Delete
	if err := repo.Delete(user1.ID); err != nil {
		t.Fatalf("Delete user1 failed: %v", err)
	}
	if err := repo.Delete(user2.ID); err != nil {
		t.Fatalf("Delete user2 failed: %v", err)
	}
	// Should not find after delete
	got, err = repo.GetByID(user1.ID)
	if err != nil {
		t.Fatalf("GetByID after delete failed: %v", err)
	}
	if got != nil {
		t.Errorf("Expected user1 to be deleted, but found: %+v", got)
	}
}
