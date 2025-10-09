package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"tamis-server/internal/database"
	"tamis-server/internal/models"
	"time"

	"github.com/lib/pq"
)

type EmailRepository struct {
	db *database.DB
}

func NewEmailRepository(db *database.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// Create - Créer un nouvel email
func (r *EmailRepository) Create(email *models.Email) (*models.Email, error) {
	query := `
        INSERT INTO emails (id, account_id, message_id, subject, from_address, to_addresses, date, size, is_read, is_spam, is_deleted, labels, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        RETURNING created_at, updated_at
    `

	now := time.Now()
	err := r.db.QueryRow(
		query,
		email.ID,
		email.AccountID,
		email.MessageID,
		email.Subject,
		email.From,
		pq.Array(email.To),
		email.Date,
		email.Size,
		email.IsRead,
		email.IsSpam,
		email.IsDeleted,
		pq.Array(email.Labels),
		now,
		now,
	).Scan(&email.CreatedAt, &email.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create email: %w", err)
	}

	return email, nil
}

// GetByID - Récupérer un email par ID
func (r *EmailRepository) GetByID(id string) (*models.Email, error) {
	query := `
        SELECT id, account_id, message_id, subject, from_address, to_addresses, date, size, is_read, is_spam, is_deleted, labels, created_at, updated_at
        FROM emails
        WHERE id = $1 AND is_deleted = false
    `

	email := &models.Email{}
	err := r.db.QueryRow(query, id).Scan(
		&email.ID,
		&email.AccountID,
		&email.MessageID,
		&email.Subject,
		&email.From,
		pq.Array(&email.To),
		&email.Date,
		&email.Size,
		&email.IsRead,
		&email.IsSpam,
		&email.IsDeleted,
		pq.Array(&email.Labels),
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found")
		}
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	return email, nil
}

// GetByMessageID - Récupérer un email par MessageID et AccountID
func (r *EmailRepository) GetByMessageID(messageID string, accountID int) (*models.Email, error) {
	query := `
        SELECT id, account_id, message_id, subject, from_address, to_addresses, date, size, is_read, is_spam, is_deleted, labels, created_at, updated_at
        FROM emails
        WHERE message_id = $1 AND account_id = $2
    `

	email := &models.Email{}
	err := r.db.QueryRow(query, messageID, accountID).Scan(
		&email.ID,
		&email.AccountID,
		&email.MessageID,
		&email.Subject,
		&email.From,
		pq.Array(&email.To),
		&email.Date,
		&email.Size,
		&email.IsRead,
		&email.IsSpam,
		&email.IsDeleted,
		pq.Array(&email.Labels),
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found")
		}
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	return email, nil
}

// GetByAccountIDsWithFilter - Récupérer les emails avec filtres et pagination
func (r *EmailRepository) GetByAccountIDsWithFilter(accountIDs []int, filter *models.EmailFilter) ([]*models.Email, int, error) {
	// Construire la requête de base
	baseQuery := `
        FROM emails 
        WHERE account_id = ANY($1) AND is_deleted = false
    `

	countQuery := "SELECT COUNT(*) " + baseQuery
	selectQuery := `
        SELECT id, account_id, message_id, subject, from_address, to_addresses, date, size, is_read, is_spam, is_deleted, labels, created_at, updated_at 
    ` + baseQuery

	args := []interface{}{pq.Array(accountIDs)}
	whereConditions := []string{}
	argIndex := 2

	// Ajouter les filtres
	if filter != nil {
		if filter.From != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("from_address ILIKE $%d", argIndex))
			args = append(args, "%"+filter.From+"%")
			argIndex++
		}

		if filter.Subject != "" {
			whereConditions = append(whereConditions, fmt.Sprintf("subject ILIKE $%d", argIndex))
			args = append(args, "%"+filter.Subject+"%")
			argIndex++
		}

		if filter.IsRead != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("is_read = $%d", argIndex))
			args = append(args, *filter.IsRead)
			argIndex++
		}

		if filter.IsSpam != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("is_spam = $%d", argIndex))
			args = append(args, *filter.IsSpam)
			argIndex++
		}

		if filter.DateFrom != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("date >= $%d", argIndex))
			args = append(args, *filter.DateFrom)
			argIndex++
		}

		if filter.DateTo != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("date <= $%d", argIndex))
			args = append(args, *filter.DateTo)
			argIndex++
		}
	}

	// Ajouter les conditions WHERE supplémentaires
	if len(whereConditions) > 0 {
		additionalWhere := " AND " + strings.Join(whereConditions, " AND ")
		countQuery += additionalWhere
		selectQuery += additionalWhere
	}

	// Compter le total
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count emails: %w", err)
	}

	// Ajouter l'ordre et la pagination
	selectQuery += " ORDER BY date DESC"

	if filter != nil {
		if filter.Limit > 0 {
			selectQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
			args = append(args, filter.Limit)
			argIndex++
		}

		if filter.Offset > 0 {
			selectQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filter.Offset)
		}
	}

	// Exécuter la requête principale
	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query emails: %w", err)
	}
	defer rows.Close()

	var emails []*models.Email
	for rows.Next() {
		email := &models.Email{}
		err := rows.Scan(
			&email.ID,
			&email.AccountID,
			&email.MessageID,
			&email.Subject,
			&email.From,
			pq.Array(&email.To),
			&email.Date,
			&email.Size,
			&email.IsRead,
			&email.IsSpam,
			&email.IsDeleted,
			pq.Array(&email.Labels),
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, totalCount, nil
}

// Update - Mettre à jour un email
func (r *EmailRepository) Update(email *models.Email) error {
	query := `
        UPDATE emails 
        SET subject = $1, from_address = $2, to_addresses = $3, date = $4, size = $5, 
            is_read = $6, is_spam = $7, is_deleted = $8, labels = $9, updated_at = $10
        WHERE id = $11
    `

	_, err := r.db.Exec(
		query,
		email.Subject,
		email.From,
		pq.Array(email.To),
		email.Date,
		email.Size,
		email.IsRead,
		email.IsSpam,
		email.IsDeleted,
		pq.Array(email.Labels),
		time.Now(),
		email.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

// UpdateReadStatus - Mettre à jour le statut lu/non-lu d'emails
func (r *EmailRepository) UpdateReadStatus(emailIDs []string, isRead bool) error {
	query := `UPDATE emails SET is_read = $1, updated_at = $2 WHERE id = ANY($3)`

	_, err := r.db.Exec(query, isRead, time.Now(), pq.Array(emailIDs))
	if err != nil {
		return fmt.Errorf("failed to update read status: %w", err)
	}

	return nil
}

// MarkAsDeleted - Marquer des emails comme supprimés (soft delete)
func (r *EmailRepository) MarkAsDeleted(emailIDs []string) error {
	query := `UPDATE emails SET is_deleted = true, updated_at = $1 WHERE id = ANY($2)`

	_, err := r.db.Exec(query, time.Now(), pq.Array(emailIDs))
	if err != nil {
		return fmt.Errorf("failed to mark emails as deleted: %w", err)
	}

	return nil
}

// DeletePermanently - Supprimer définitivement des emails
func (r *EmailRepository) DeletePermanently(emailIDs []string) error {
	query := `DELETE FROM emails WHERE id = ANY($1)`

	_, err := r.db.Exec(query, pq.Array(emailIDs))
	if err != nil {
		return fmt.Errorf("failed to delete emails permanently: %w", err)
	}

	return nil
}

// ArchiveEmails - Archiver des emails (ajouter le label "archived")
func (r *EmailRepository) ArchiveEmails(emailIDs []string) error {
	query := `
        UPDATE emails 
        SET labels = array_append(labels, 'archived'), updated_at = $1 
        WHERE id = ANY($2) AND NOT 'archived' = ANY(labels)
    `

	_, err := r.db.Exec(query, time.Now(), pq.Array(emailIDs))
	if err != nil {
		return fmt.Errorf("failed to archive emails: %w", err)
	}

	return nil
}

// GetByAccountID - Récupérer tous les emails d'un compte
func (r *EmailRepository) GetByAccountID(accountID int, limit, offset int) ([]*models.Email, error) {
	query := `
        SELECT id, account_id, message_id, subject, from_address, to_addresses, date, size, is_read, is_spam, is_deleted, labels, created_at, updated_at
        FROM emails
        WHERE account_id = $1 AND is_deleted = false
        ORDER BY date DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails by account: %w", err)
	}
	defer rows.Close()

	var emails []*models.Email
	for rows.Next() {
		email := &models.Email{}
		err := rows.Scan(
			&email.ID,
			&email.AccountID,
			&email.MessageID,
			&email.Subject,
			&email.From,
			pq.Array(&email.To),
			&email.Date,
			&email.Size,
			&email.IsRead,
			&email.IsSpam,
			&email.IsDeleted,
			pq.Array(&email.Labels),
			&email.CreatedAt,
			&email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// DeleteByAccountID - Supprimer tous les emails d'un compte (lors de suppression du compte)
func (r *EmailRepository) DeleteByAccountID(accountID int) error {
	query := `DELETE FROM emails WHERE account_id = $1`

	_, err := r.db.Exec(query, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete emails by account: %w", err)
	}

	return nil
}
