package api

import (
	"net/http"
	"tamis-server/internal/config"
	"tamis-server/internal/handlers"
	"tamis-server/internal/middleware"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

// RegisterRoutes - Point d'entrée principal pour l'enregistrement des routes
func RegisterRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	logger *utils.Logger,
	authService *services.AuthService,
	authMiddleware *middleware.AuthMiddleware,
	accountService *services.AccountService,
	mailService *services.MailService,
	oauth2Service *utils.OAuth2Service,
) {
	// Routes d'authentification (publiques)
	registerAuthRoutes(mux, authService, logger)

	// Routes d'API générales
	registerAPIRoutes(mux, cfg, logger, authMiddleware)

	// Routes de gestion des comptes email (protégées)
	registerAccountRoutes(mux, authMiddleware, accountService, logger)

	// Routes OAuth2 (protégées)
	registerOAuthRoutes(mux, authMiddleware, oauth2Service, accountService, logger)

	// Routes de gestion des emails (protégées)
	registerMailRoutes(mux, authMiddleware, mailService, logger)
}

// registerAuthRoutes - Routes d'authentification
func registerAuthRoutes(mux *http.ServeMux, authService *services.AuthService, logger *utils.Logger) {
	mux.HandleFunc("/api/auth/register", corsMiddleware(registerHandler(authService, logger)))
	mux.HandleFunc("/api/auth/login", corsMiddleware(loginHandler(authService, logger)))
	mux.HandleFunc("/api/auth/refresh", corsMiddleware(refreshTokenHandler(authService, logger)))
}

// registerAPIRoutes - Routes de l'API (protégées et publiques)
func registerAPIRoutes(mux *http.ServeMux, cfg *config.Config, logger *utils.Logger, authMiddleware *middleware.AuthMiddleware) {
	// Route de santé (publique)
	mux.HandleFunc("/api/health", corsMiddleware(healthHandler(cfg, logger)))

	// Routes protégées
	mux.Handle("/api/user/me",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(meHandler(logger))),
		))
}

// registerAccountRoutes - Routes de gestion des comptes email
func registerAccountRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, accountService *services.AccountService, logger *utils.Logger) {
	// Ajouter un compte email
	mux.Handle("/api/accounts/add",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(AddAccountHandler(accountService, logger))),
		))

	// Lister les comptes email
	mux.Handle("/api/accounts",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(ListAccountsHandler(accountService, logger))),
		))

	// Supprimer un compte email
	mux.Handle("/api/accounts/remove",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(DeleteAccountHandler(accountService, logger))),
		))
}

// registerOAuthRoutes - Routes OAuth2
func registerOAuthRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, oauth2Service *utils.OAuth2Service, accountService *services.AccountService, logger *utils.Logger) {
	// Initier OAuth Google
	mux.Handle("/api/oauth/google/initiate",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.InitiateGoogleOAuthHandler(oauth2Service, logger))),
		))

	// Callback OAuth Google (public)
	mux.HandleFunc("/api/oauth/google/callback", corsMiddleware(handlers.GoogleOAuthCallbackHandler(oauth2Service, accountService, logger)))

	// Finaliser l'ajout du compte
	mux.Handle("/api/oauth/complete",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.CompleteAccountSetupHandler(oauth2Service, accountService, logger))),
		))
}

// registerMailRoutes - Routes de gestion des emails
func registerMailRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, mailService *services.MailService, logger *utils.Logger) {
	// Lister tous les emails consolidés
	mux.Handle("/api/mails",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(ListMailsHandler(mailService, logger))),
		))

	// Actions sur les emails (supprimer, archiver, marquer lu)
	mux.Handle("/api/mails/action",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(MailActionHandler(mailService, logger))),
		))

	// Synchroniser les emails
	mux.Handle("/api/mails/sync",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(SyncMailsHandler(mailService, logger))),
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
