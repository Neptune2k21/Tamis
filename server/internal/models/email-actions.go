package models

import "time"

// EmailAction - Types d'actions sur les emails
type EmailAction string

const (
	ActionDelete     EmailAction = "delete"
	ActionArchive    EmailAction = "archive"
	ActionMarkRead   EmailAction = "mark_read"
	ActionMarkUnread EmailAction = "mark_unread"
	ActionSpam       EmailAction = "mark_spam"
	ActionNotSpam    EmailAction = "mark_not_spam"
)

// EmailActionRequest - Requête d'action sur des emails
type EmailActionRequest struct {
	EmailIDs []string    `json:"email_ids" validate:"required,min=1"`
	Action   EmailAction `json:"action" validate:"required"`
	Force    bool        `json:"force,omitempty"`
}

// EmailActionResult - Résultat d'une action sur des emails
type EmailActionResult struct {
	Action       EmailAction `json:"action"`
	ProcessedIDs []string    `json:"processed_ids"`
	FailedIDs    []string    `json:"failed_ids"`
	SuccessCount int         `json:"success_count"`
	FailureCount int         `json:"failure_count"`
	Message      string      `json:"message,omitempty"`
}

// EmailSyncResult - Résultat de synchronisation des emails
type EmailSyncResult struct {
	SyncedCount   int       `json:"synced_count"`
	FailedCount   int       `json:"failed_count"`
	NewEmails     int       `json:"new_emails"`
	UpdatedEmails int       `json:"updated_emails"`
	Accounts      []string  `json:"accounts"`
	LastSync      time.Time `json:"last_sync"`
}

// AccountSyncResult - Résultat de synchronisation d'un compte
type AccountSyncResult struct {
	NewEmails     int `json:"new_emails"`
	UpdatedEmails int `json:"updated_emails"`
}

// OAuth2Token - Token OAuth2
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// DecryptedTokens - Tokens déchiffrés (usage interne)
type DecryptedTokens struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresAt    *time.Time `json:"expires_at"`
}
