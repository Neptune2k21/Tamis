package api

import (
	"net/http"
	"tamis-server/internal/config"
	"tamis-server/internal/handlers"
	"tamis-server/internal/middleware"
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
	oauth2Service *services.OAuth2Service,
) {
	// Routes d'authentification (publiques)
	registerAuthRoutes(mux, authMiddleware, authService, logger, oauth2Service)

	// Routes d'API générales
	registerAPIRoutes(mux, cfg, logger, authMiddleware)

	// Routes de gestion des comptes email (protégées)
	registerAccountRoutes(mux, authMiddleware, accountService, logger)

	// Routes OAuth2 (protégées)
	registerOAuthRoutes(mux, authMiddleware, oauth2Service, accountService, logger)

	// Routes de gestion des emails (protégées)
	registerMailRoutes(mux, authMiddleware, mailService, logger)
}

// registerAuthRoutes - Routes d'authentification publiques
func registerAuthRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, authService *services.AuthService, logger *utils.Logger, oauth2Service *services.OAuth2Service) {
	mux.Handle("/api/auth/register", authMiddleware.CORS(http.HandlerFunc(handlers.RegisterHandler(authService, logger))))
	mux.Handle("/api/auth/login", authMiddleware.CORS(http.HandlerFunc(handlers.LoginHandler(authService, logger))))
	mux.Handle("/api/auth/refresh", authMiddleware.CORS(http.HandlerFunc(handlers.RefreshTokenHandler(authService, logger))))
	mux.Handle("/api/auth/google", authMiddleware.CORS(http.HandlerFunc(handlers.PublicGoogleOAuthHandler(oauth2Service, logger))))
	mux.Handle("/api/auth/google/callback", authMiddleware.CORS(http.HandlerFunc(handlers.PublicGoogleOAuthCallbackHandler(oauth2Service, authService, logger))))
	mux.Handle("/api/auth/google/callback-process", authMiddleware.CORS(http.HandlerFunc(handlers.ProcessOAuthCallbackHandler(authService, logger))))
}

// registerAPIRoutes - Routes de l'API (protégées et publiques)
func registerAPIRoutes(mux *http.ServeMux, cfg *config.Config, logger *utils.Logger, authMiddleware *middleware.AuthMiddleware) {
	// Route de santé (publique)
	mux.Handle("/api/health", authMiddleware.CORS(http.HandlerFunc(HealthHandler(cfg, logger))))

	// Route utilisateur (protégée)
	mux.Handle("/api/user/me",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.MeHandler(logger))),
		))
}

// registerAccountRoutes - Routes de gestion des comptes email
func registerAccountRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, accountService *services.AccountService, logger *utils.Logger) {
	mux.Handle("/api/accounts/add",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.AddAccountHandler(accountService, logger))),
		))

	mux.Handle("/api/accounts",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.ListAccountsHandler(accountService, logger))),
		))

	mux.Handle("/api/accounts/remove",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.DeleteAccountHandler(accountService, logger))),
		))
}

// registerOAuthRoutes - Routes OAuth2
func registerOAuthRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, oauth2Service *services.OAuth2Service, accountService *services.AccountService, logger *utils.Logger) {
	// Initier OAuth Google
	mux.Handle("/api/oauth/google/initiate",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.InitiateGoogleOAuthHandler(oauth2Service, logger))),
		))

	// Callback OAuth Google (public)
	mux.Handle("/api/oauth/google/callback", authMiddleware.CORS(http.HandlerFunc(handlers.GoogleOAuthCallbackHandler(oauth2Service, accountService, logger))))

	// Finaliser l'ajout du compte
	mux.Handle("/api/oauth/complete",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.CompleteAccountSetupHandler(oauth2Service, accountService, logger))),
		))
}

// registerMailRoutes - Routes de gestion des emails
func registerMailRoutes(mux *http.ServeMux, authMiddleware *middleware.AuthMiddleware, mailService *services.MailService, logger *utils.Logger) {
	mux.Handle("/api/mails",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.ListMailsHandler(mailService, logger))),
		))

	mux.Handle("/api/mails/action",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.MailActionHandler(mailService, logger))),
		))

	mux.Handle("/api/mails/sync",
		authMiddleware.CORS(
			authMiddleware.RequireAuth(http.HandlerFunc(handlers.SyncMailsHandler(mailService, logger))),
		))
}

// HealthHandler - Vérifier l'état du service
func HealthHandler(cfg *config.Config, logger *utils.Logger) http.HandlerFunc {
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
