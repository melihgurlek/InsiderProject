package domain

import (
	"errors"
	"strings"
	"time"
)

// User represents a system user.
type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time // Use time.Time in real code, string for simplicity now
	UpdatedAt    time.Time
}

// Validate checks if the user fields are valid.
func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(u.PasswordHash) == "" {
		return errors.New("password hash is required")
	}
	if u.Role != "user" && u.Role != "admin" {
		return errors.New("role must be 'user' or 'admin'")
	}
	return nil
}
