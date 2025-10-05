package models

import "time"

// User représente un utilisateur dans la base de données
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest - Requête d'inscription (AVEC password maintenant!)
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest - Requête de connexion
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse - Réponse après connexion réussie
type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

// RefreshTokenRequest - Requête pour rafraîchir le token
type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// RefreshTokenResponse - Réponse avec nouveau token
type RefreshTokenResponse struct {
	Token string `json:"token"`
}
