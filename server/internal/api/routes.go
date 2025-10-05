package api

import (
	"encoding/json"
	"net/http"
	"tamis-server/internal/config"
	"tamis-server/internal/middleware"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

func RegisterRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	logger *utils.Logger,
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Routes publiques (pas de JWT requis)
	mux.HandleFunc("/api/auth/register", corsMiddleware(registerHandler(authService, logger)))
	mux.HandleFunc("/api/auth/login", corsMiddleware(loginHandler(authService, logger)))
	mux.HandleFunc("/api/auth/refresh", corsMiddleware(refreshTokenHandler(authService, logger)))
	mux.HandleFunc("/api/health", corsMiddleware(healthHandler(cfg, logger)))

	// Routes protégées (JWT requis via Authorization: Bearer <token>)
	mux.Handle("/api/user/me",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(meHandler(logger))),
		))

	mux.Handle("/api/emails",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(emailsHandler(cfg, logger))),
		))

	mux.Handle("/api/emails/delete",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(deleteHandler(cfg, logger))),
		))

}

// corsMiddleware - CORS pour les routes publiques
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// registerHandler - Inscription d'un nouvel utilisateur
func registerHandler(authService *services.AuthService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req models.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validation des champs
		if req.Email == "" || req.Username == "" || req.Password == "" {
			utils.WriteError(w, http.StatusBadRequest, "Email, username and password are required")
			return
		}

		if len(req.Password) < 8 {
			utils.WriteError(w, http.StatusBadRequest, "Password must be at least 8 characters")
			return
		}

		user, err := authService.Register(&req)
		if err != nil {
			logger.Error("Registration failed: " + err.Error())
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Ne pas retourner le mot de passe dans la réponse
		utils.WriteSuccess(w, map[string]interface{}{
			"user": user,
		}, "User registered successfully. Please login to get your token.")
	}
}

// loginHandler - Connexion et génération de JWT
func loginHandler(authService *services.AuthService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validation
		if req.Email == "" || req.Password == "" {
			utils.WriteError(w, http.StatusBadRequest, "Email and password are required")
			return
		}

		response, err := authService.Login(&req)
		if err != nil {
			logger.Warn("Login failed for email: " + req.Email)
			utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		logger.Info("Successful login for user: " + response.User.Email)
		utils.WriteSuccess(w, response, "Login successful")
	}
}

// refreshTokenHandler - Rafraîchir un token JWT
func refreshTokenHandler(authService *services.AuthService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req models.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		if req.Token == "" {
			utils.WriteError(w, http.StatusBadRequest, "Token is required")
			return
		}

		newToken, err := authService.RefreshToken(req.Token)
		if err != nil {
			logger.Warn("Token refresh failed: " + err.Error())
			utils.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		utils.WriteSuccess(w, models.RefreshTokenResponse{
			Token: newToken,
		}, "Token refreshed successfully")
	}
}

// meHandler - Récupérer les informations de l'utilisateur connecté
func meHandler(logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Récupérer l'utilisateur depuis le contexte (injecté par le middleware)
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found in context")
			return
		}

		logger.Info("User info requested: " + user.Email)
		utils.WriteSuccess(w, user, "User retrieved successfully")
	}
}

// emailsHandler - Route protégée pour gérer les emails
func emailsHandler(cfg *config.Config, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Récupérer l'utilisateur depuis le contexte
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		logger.Info("Emails endpoint called by: " + user.Email)

		// Exemple de réponse
		utils.WriteSuccess(w, map[string]interface{}{
			"message":  "Emails endpoint ready",
			"user_id":  user.ID,
			"username": user.Username,
			"emails":   []string{"example@example.com"},
		}, "Emails retrieved successfully")
	}
}

// deleteHandler - Route protégée pour supprimer des emails
func deleteHandler(cfg *config.Config, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete && r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		logger.Info("Delete endpoint called by: " + user.Email)

		// Exemple de réponse
		utils.WriteSuccess(w, map[string]interface{}{
			"message": "Delete endpoint ready",
			"user_id": user.ID,
		}, "Ready to delete emails")
	}
}

// healthHandler - Vérifier l'état du service
func healthHandler(cfg *config.Config, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		utils.WriteSuccess(w, map[string]string{
			"status":      "healthy",
			"environment": cfg.Server.Env,
			"version":     "1.0.0",
		}, "Service is running")
	}
}
