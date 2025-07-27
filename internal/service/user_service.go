package service

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/pkg/metrics"
)

// UserServiceImpl implements domain.UserService.
type UserServiceImpl struct {
	repo domain.UserRepository
}

// NewUserService creates a new UserServiceImpl.
func NewUserService(repo domain.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

// Register creates a new user with hashed password after validation.
func (s *UserServiceImpl) Register(username, email, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}
	if existing, _ := s.repo.GetByUsername(username); existing != nil {
		return nil, errors.New("username already exists")
	}
	if existing, _ := s.repo.GetByEmail(email); existing != nil {
		return nil, errors.New("email already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	// Record business metrics
	metrics.UserRegistrationTotal.Inc()

	return user, nil
}

// Login checks username and password, returns user if valid.
func (s *UserServiceImpl) Login(username, password string) (*domain.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil || user == nil {
		// Record failed login
		metrics.UserLoginTotal.WithLabelValues("failure").Inc()
		return nil, errors.New("invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Record failed login
		metrics.UserLoginTotal.WithLabelValues("failure").Inc()
		return nil, errors.New("invalid username or password")
	}

	// Record successful login
	metrics.UserLoginTotal.WithLabelValues("success").Inc()

	return user, nil
}

// GetUser returns a user by ID.
func (s *UserServiceImpl) GetUser(id int) (*domain.User, error) {
	return s.repo.GetByID(id)
}

// ListUsers returns all users.
func (s *UserServiceImpl) ListUsers() ([]*domain.User, error) {
	return s.repo.List()
}

// UpdateUser updates a user (does not change password).
func (s *UserServiceImpl) UpdateUser(user *domain.User) error {
	return s.repo.Update(user)
}

// DeleteUser deletes a user by ID.
func (s *UserServiceImpl) DeleteUser(id int) error {
	return s.repo.Delete(id)
}
