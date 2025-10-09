package models

import (
	"time"
)

type EmailProvider string

const (
	ProviderGmail   EmailProvider = "gmail"
	ProviderYahoo   EmailProvider = "yahoo"
	ProviderOutlook EmailProvider = "outlook"
	ProviderOther   EmailProvider = "other"
)

type EmailAccount struct {
	ID             int           `json:"id" db:"id"`
	UserID         int           `json:"user_id" db:"user_id"`
	Provider       EmailProvider `json:"provider" db:"provider"`
	Email          string        `json:"email" db:"email"`
	DisplayName    string        `json:"display_name" db:"display_name"`
	AccessToken    string        `json:"-" db:"access_token"`
	RefreshToken   string        `json:"-" db:"refresh_token"`
	TokenExpiresAt *time.Time    `json:"-" db:"token_expires_at"`
	IsActive       bool          `json:"is_active" db:"is_active"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

type CreateEmailAccountRequest struct {
	Provider    EmailProvider `json:"provider" validate:"required"`
	Email       string        `json:"email" validate:"required,email"`
	DisplayName string        `json:"display_name" validate:"required"`
}
