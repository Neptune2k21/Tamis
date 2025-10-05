package models

import (
	"time"
)

type Email struct {
	ID        string    `json:"id" db:"id"`
	AccountID int       `json:"account_id" db:"account_id"`
	MessageID string    `json:"message_id" db:"message_id"`
	Subject   string    `json:"subject" db:"subject"`
	From      string    `json:"from" db:"from_address"`
	To        []string  `json:"to" db:"to_addresses"`
	Date      time.Time `json:"date" db:"date"`
	Size      int64     `json:"size" db:"size"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	IsSpam    bool      `json:"is_spam" db:"is_spam"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
	Labels    []string  `json:"labels" db:"labels"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type EmailFilter struct {
	Provider EmailProvider `json:"provider,omitempty"`
	From     string        `json:"from,omitempty"`
	Subject  string        `json:"subject,omitempty"`
	IsRead   *bool         `json:"is_read,omitempty"`
	IsSpam   *bool         `json:"is_spam,omitempty"`
	DateFrom *time.Time    `json:"date_from,omitempty"`
	DateTo   *time.Time    `json:"date_to,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	Offset   int           `json:"offset,omitempty"`
}

type DeleteEmailsRequest struct {
	EmailIDs []string `json:"email_ids" validate:"required,min=1"`
	Force    bool     `json:"force,omitempty"` // Suppression définitive ou corbeille
}
