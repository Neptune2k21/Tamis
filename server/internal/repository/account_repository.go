package repository

import (
	"database/sql"
	"fmt"
	"tamis-server/internal/database"
	"tamis-server/internal/models"
	"time"
)

type AccountRepository struct {
	db *database.DB
}

func NewAccountRepository(db *database.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create - Créer un nouveau compte email
func (r *AccountRepository) Create(account *models.EmailAccount) (*models.EmailAccount, error) {
	query := `
        INSERT INTO email_accounts (user_id, provider, email, display_name, access_token, refresh_token, token_expires_at, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at, updated_at
    `

	now := time.Now()
	err := r.db.QueryRow(
		query,
		account.UserID,
		account.Provider,
		account.Email,
		account.DisplayName,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.IsActive,
		now,
		now,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create email account: %w", err)
	}

	return account, nil
}

// GetByUserID - Récupérer tous les comptes d'un utilisateur
func (r *AccountRepository) GetByUserID(userID int) ([]*models.EmailAccount, error) {
	query := `
        SELECT id, user_id, provider, email, display_name, is_active, created_at, updated_at
        FROM email_accounts
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query email accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*models.EmailAccount
	for rows.Next() {
		account := &models.EmailAccount{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.Email,
			&account.DisplayName,
			&account.IsActive,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetByID - Récupérer un compte par ID (avec tokens)
func (r *AccountRepository) GetByID(id int) (*models.EmailAccount, error) {
	query := `
        SELECT id, user_id, provider, email, display_name, access_token, refresh_token, token_expires_at, is_active, created_at, updated_at
        FROM email_accounts
        WHERE id = $1
    `

	account := &models.EmailAccount{}
	err := r.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.Email,
		&account.DisplayName,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email account not found")
		}
		return nil, fmt.Errorf("failed to get email account: %w", err)
	}

	return account, nil
}

// GetByUserAndEmail - Récupérer un compte par utilisateur et email
func (r *AccountRepository) GetByUserAndEmail(userID int, email string) (*models.EmailAccount, error) {
	query := `
        SELECT id, user_id, provider, email, display_name, is_active, created_at, updated_at
        FROM email_accounts
        WHERE user_id = $1 AND email = $2
    `

	account := &models.EmailAccount{}
	err := r.db.QueryRow(query, userID, email).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.Email,
		&account.DisplayName,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email account not found")
		}
		return nil, fmt.Errorf("failed to get email account: %w", err)
	}

	return account, nil
}

// UpdateTokens - Mettre à jour les tokens OAuth2
func (r *AccountRepository) UpdateTokens(id int, accessToken, refreshToken string, expiresAt *time.Time) error {
	query := `
        UPDATE email_accounts 
        SET access_token = $1, refresh_token = $2, token_expires_at = $3, updated_at = $4
        WHERE id = $5
    `

	_, err := r.db.Exec(query, accessToken, refreshToken, expiresAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	return nil
}

// SetActive - Activer/désactiver un compte
func (r *AccountRepository) SetActive(id int, isActive bool) error {
	query := `UPDATE email_accounts SET is_active = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.Exec(query, isActive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update account status: %w", err)
	}

	return nil
}

// Delete - Supprimer un compte
func (r *AccountRepository) Delete(id int) error {
	query := `DELETE FROM email_accounts WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete email account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("email account not found")
	}

	return nil
}
