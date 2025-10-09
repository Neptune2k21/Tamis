package api

import (
	"encoding/json"
	"net/http"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

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
