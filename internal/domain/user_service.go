package domain

// UserService defines business logic for users.
type UserService interface {
	Register(username, email, password string) (*User, error)
	Login(username, password string) (*User, error)
	GetUser(id int) (*User, error)
	ListUsers() ([]*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}
