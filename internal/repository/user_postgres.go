package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// UserPostgresRepository implements domain.UserRepository using PostgreSQL.
type UserPostgresRepository struct {
	conn *pgx.Conn
}

// NewUserPostgresRepository creates a new UserPostgresRepository.
func NewUserPostgresRepository(conn *pgx.Conn) *UserPostgresRepository {
	return &UserPostgresRepository{conn: conn}
}

// Create inserts a new user into the database.
func (r *UserPostgresRepository) Create(user *domain.User) error {
	query := `INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id, created_at, updated_at`
	return r.conn.QueryRow(context.Background(), query,
		user.Username, user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID fetches a user by ID.
func (r *UserPostgresRepository) GetByID(id int) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	err := r.conn.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}
	return user, nil
}

// GetByUsername fetches a user by username.
func (r *UserPostgresRepository) GetByUsername(username string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	err := r.conn.QueryRow(context.Background(), query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail fetches a user by email.
func (r *UserPostgresRepository) GetByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`
	err := r.conn.QueryRow(context.Background(), query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Update modifies an existing user in the database.
func (r *UserPostgresRepository) Update(user *domain.User) error {
	query := `UPDATE users SET username = $1, email = $2, password_hash = $3, role = $4, updated_at = NOW() WHERE id = $5`
	cmdTag, err := r.conn.Exec(context.Background(), query,
		user.Username, user.Email, user.PasswordHash, user.Role, user.ID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

// Delete removes a user by ID.
func (r *UserPostgresRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	cmdTag, err := r.conn.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

// List returns all users.
func (r *UserPostgresRepository) List() ([]*domain.User, error) {
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users`
	rows, err := r.conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
