package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"tamis-server/internal/middleware"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

// InitiateGoogleOAuth - Initier le flux OAuth Google
func InitiateGoogleOAuthHandler(oauth2Service *utils.OAuth2Service, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Vérifier que l'utilisateur est connecté
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Générer un state unique pour la sécurité
		state := generateSecureState()

		// En production, on devrait stocker le state en session/cache
		// pour le valider lors du callback

		authURL := oauth2Service.GetGoogleAuthURL(state)

		utils.WriteSuccess(w, map[string]string{
			"auth_url": authURL,
			"state":    state,
		}, "Google OAuth URL generated")
	}
}

// GoogleOAuthCallbackHandler - Callback après autorisation Google
func GoogleOAuthCallbackHandler(
	oauth2Service *utils.OAuth2Service,
	accountService *services.AccountService,
	logger *utils.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Récupérer les paramètres
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		if errorParam != "" {
			logger.Error("OAuth error: " + errorParam)
			utils.WriteError(w, http.StatusBadRequest, "OAuth authorization failed")
			return
		}

		if code == "" {
			utils.WriteError(w, http.StatusBadRequest, "Authorization code missing")
			return
		}

		// Valider le state (en production, vérifier contre la session)
		if state == "" {
			utils.WriteError(w, http.StatusBadRequest, "Invalid state parameter")
			return
		}

		// Échanger le code contre des tokens
		tokens, err := oauth2Service.ExchangeCodeForTokens(code)
		if err != nil {
			logger.Error("Failed to exchange code for tokens: " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "Failed to exchange authorization code")
			return
		}

		// Récupérer les informations utilisateur
		userInfo, err := oauth2Service.GetUserInfo(tokens.AccessToken)
		if err != nil {
			logger.Error("Failed to get user info: " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "Failed to get user information")
			return
		}

		// Rediriger vers le frontend avec les informations
		// En production, on passerait ces données de manière sécurisée
		redirectURL := fmt.Sprintf("http://localhost:3001/oauth/callback?email=%s&name=%s&provider=gmail",
			userInfo.Email, userInfo.Name)

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// CompleteAccountSetupHandler
func CompleteAccountSetupHandler(
	oauth2Service *utils.OAuth2Service,
	accountService *services.AccountService,
	logger *utils.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Récupérer l'utilisateur depuis le contexte JWT
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Structure pour recevoir le code d'autorisation
		var req struct {
			Code string `json:"code"`
		}

		if err := utils.DecodeJSON(r, &req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}

		// Échanger le code contre des tokens
		tokens, err := oauth2Service.ExchangeCodeForTokens(req.Code)
		if err != nil {
			logger.Error("Failed to exchange code: " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "Failed to exchange authorization code")
			return
		}

		// Récupérer les informations utilisateur
		userInfo, err := oauth2Service.GetUserInfo(tokens.AccessToken)
		if err != nil {
			logger.Error("Failed to get user info: " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "Failed to get user information")
			return
		}

		// Créer la requête de compte
		accountReq := &models.CreateEmailAccountRequest{
			Provider:    models.ProviderGmail,
			Email:       userInfo.Email,
			DisplayName: userInfo.Name,
		}

		// Ajouter le compte avec les vrais tokens OAuth2
		account, err := accountService.AddAccountWithTokens(user.ID, accountReq, tokens)
		if err != nil {
			logger.Error("Failed to add Gmail account: " + err.Error())
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info(fmt.Sprintf("Gmail account added for user %d: %s", user.ID, userInfo.Email))
		utils.WriteSuccess(w, account, "Gmail account added successfully")
	}
}

func generateSecureState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
