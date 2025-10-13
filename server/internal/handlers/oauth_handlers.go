package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"tamis-server/internal/middleware"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

// InitiateGoogleOAuth - Initier le flux OAuth Google
func InitiateGoogleOAuthHandler(oauth2Service *services.OAuth2Service, logger *utils.Logger) http.HandlerFunc {
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
	oauth2Service *services.OAuth2Service,
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

		// Gérer le cas où le nom est vide
		displayName := userInfo.Name
		if displayName == "" {
			// Générer un nom à partir de l'email
			emailParts := strings.Split(userInfo.Email, "@")
			if len(emailParts) > 0 {
				displayName = emailParts[0]
			} else {
				displayName = "Utilisateur"
			}
		}

		// Rediriger vers le frontend avec les informations
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3001"
		}

		// Encoder les paramètres pour éviter les problèmes avec les caractères spéciaux
		redirectURL := fmt.Sprintf("%s/oauth/callback?email=%s&name=%s&provider=gmail",
			frontendURL,
			url.QueryEscape(userInfo.Email),
			url.QueryEscape(displayName))

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// CompleteAccountSetupHandler
func CompleteAccountSetupHandler(
	oauth2Service *services.OAuth2Service,
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

// PublicGoogleOAuthHandler - génère l'URL OAuth Google pour le front
func PublicGoogleOAuthHandler(oauth2Service *services.OAuth2Service, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		state := generateSecureState()

		// En production, stocker le state en cache/session pour validation
		authURL := oauth2Service.GetGoogleAuthURL(state)

		utils.WriteSuccess(w, map[string]string{
			"auth_url": authURL,
			"state":    state,
		}, "Google OAuth URL generated for public access")
	}
}

// PublicGoogleOAuthCallbackHandler - Callback OAuth Google pour inscription/connexion directe
func PublicGoogleOAuthCallbackHandler(
	oauth2Service *services.OAuth2Service,
	authService *services.AuthService,
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
			RedirectToFrontendWithError(w, r, "OAuth authorization failed")
			return
		}

		if code == "" {
			RedirectToFrontendWithError(w, r, "Authorization code missing")
			return
		}

		// Valider le state (en production, vérifier contre le cache)
		if state == "" {
			RedirectToFrontendWithError(w, r, "Invalid state parameter")
			return
		}

		// Échanger le code contre des tokens
		tokens, err := oauth2Service.ExchangeCodeForTokens(code)
		if err != nil {
			logger.Error("Failed to exchange code for tokens: " + err.Error())
			RedirectToFrontendWithError(w, r, "Failed to exchange authorization code")
			return
		}

		// Récupérer les informations utilisateur
		userInfo, err := oauth2Service.GetUserInfo(tokens.AccessToken)
		if err != nil {
			logger.Error("Failed to get user info: " + err.Error())
			RedirectToFrontendWithError(w, r, "Failed to get user information")
			return
		}

		// Connecter ou inscrire l'utilisateur
		loginResponse, err := authService.LoginWithOAuth(userInfo.Email, userInfo.Name, "google")
		if err != nil {
			logger.Error("Failed to login/register OAuth user: " + err.Error())
			RedirectToFrontendWithError(w, r, err.Error())
			return
		}

		// Rediriger vers le frontend avec le token JWT
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3001"
		}

		redirectURL := fmt.Sprintf("%s/oauth/success?token=%s&user=%s",
			frontendURL, loginResponse.Token, loginResponse.User.Email)

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// RedirectToFrontendWithError -
func RedirectToFrontendWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}

	redirectURL := fmt.Sprintf("%s/oauth/error?error=%s", frontendURL, url.QueryEscape(errorMsg))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// ProcessOAuthCallbackHandler - Traiter le callback OAuth côté API
func ProcessOAuthCallbackHandler(
	authService *services.AuthService,
	logger *utils.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Provider string `json:"provider"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		if req.Email == "" || req.Provider == "" {
			utils.WriteError(w, http.StatusBadRequest, "Email and provider are required")
			return
		}

		// Connecter ou inscrire l'utilisateur
		loginResponse, err := authService.LoginWithOAuth(req.Email, req.Name, req.Provider)
		if err != nil {
			logger.Error("Failed to login/register OAuth user: " + err.Error())
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info("OAuth login successful for: " + req.Email)
		utils.WriteSuccess(w, loginResponse, "OAuth login successful")
	}
}
