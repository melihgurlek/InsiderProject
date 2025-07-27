package domain

import "context"

// UserRepository defines methods for user data access.
type UserRepository interface {
	Create(user *User) error
	GetByID(id int) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id int) error
	List() ([]*User, error)
	Ping(ctx context.Context) error
}
