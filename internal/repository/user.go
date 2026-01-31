package repository

import (
	"context"
	"errors"

	"ioi-amms/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user with this email already exists")
)

// UserRepository handles user data access
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (tenant_id, email, password_hash, role, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.FirstName,
		user.LastName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $2, role = $3, first_name = $4, last_name = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	return r.db.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.Role,
		user.FirstName,
		user.LastName,
	).Scan(&user.UpdatedAt)
}

// UpdatePassword changes a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, passwordHash)
	return err
}

// List retrieves all users for a tenant
func (r *UserRepository) List(ctx context.Context, tenantID string) ([]model.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, role, first_name, last_name, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.FirstName,
			&user.LastName,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
