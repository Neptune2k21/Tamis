package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"tamis-server/internal/repository"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	userRepo    *repository.UserRepository
	authService *services.AuthService
	logger      *utils.Logger
}

func NewAuthMiddleware(userRepo *repository.UserRepository, authService *services.AuthService, logger *utils.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:    userRepo,
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth - Middleware pour protéger les routes
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Récupérer le header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing Authorization header")
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "missing authorization token",
			})
			return
		}

		// Vérifier le format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Warn("Invalid Authorization header format")
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid authorization format",
			})
			return
		}

		tokenString := parts[1]

		// Valider le JWT
		claims, err := m.authService.ValidateJWT(tokenString)
		if err != nil {
			m.logger.Warn(fmt.Sprintf("Invalid token: %v", err))
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
			return
		}

		// Récupérer l'utilisateur complet
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			m.logger.Error(fmt.Sprintf("User not found for valid token: %d", claims.UserID))
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "user not found",
			})
			return
		}

		// Ajouter l'utilisateur au contexte
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS - Middleware pour gérer le CORS
func (m *AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
