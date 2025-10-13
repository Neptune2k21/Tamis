package repository

import (
	"database/sql"
	"fmt"
	"tamis-server/internal/database"
	"tamis-server/internal/models"
	"time"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User, passwordHash string) (*models.User, error) {
	query := `
        INSERT INTO users (email, username, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `

	now := time.Now()
	err := r.db.QueryRow(
		query,
		user.Email,
		user.Username,
		passwordHash,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetPasswordHash(email string) (string, error) {
	query := `SELECT password_hash FROM users WHERE email = $1`

	var passwordHash string
	err := r.db.QueryRow(query, email).Scan(&passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to get password hash: %w", err)
	}

	return passwordHash, nil
}

// ...existing code...

// CreateOAuthUser - Créer un utilisateur OAuth (sans mot de passe)
func (r *UserRepository) CreateOAuthUser(user *models.User, authProvider string) (*models.User, error) {
	query := `
        INSERT INTO users (email, username, auth_provider, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `

	now := time.Now()
	err := r.db.QueryRow(
		query,
		user.Email,
		user.Username,
		authProvider,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth user: %w", err)
	}

	user.AuthProvider = authProvider
	return user, nil
}

// GetByEmail - Récupérer un utilisateur par email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
        SELECT id, email, username, COALESCE(auth_provider, 'local') as auth_provider, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.AuthProvider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByID - Récupérer un utilisateur par ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
        SELECT id, email, username, COALESCE(auth_provider, 'local') as auth_provider, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.AuthProvider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername - Récupérer un utilisateur par nom d'utilisateur
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
        SELECT id, email, username, COALESCE(auth_provider, 'local') as auth_provider, created_at, updated_at
        FROM users
        WHERE username = $1
    `

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.AuthProvider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
