package main

import (
	"fmt"
	"net/http"
	"tamis-server/internal/api"
	"tamis-server/internal/config"
	"tamis-server/internal/database"
	"tamis-server/internal/middleware"
	"tamis-server/internal/repository"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

func main() {
	// Charger la configuration
	cfg := config.Load()
	logger := utils.NewLogger()

	logger.Info("Starting Tamis Server...")
	logger.Info(fmt.Sprintf("Environment: %s", cfg.Server.Env))

	// Connexion à la base de données
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Exécuter les migrations
	migrationsPath := "./migrations"
	if cfg.Server.Env == "production" {
		migrationsPath = "./migrations"
	} else {
		migrationsPath = "./internal/migrations"
	}

	logger.Info("Running database migrations...")
	if err := db.RunMigrations(migrationsPath); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to run migrations: %v", err))
	}
	logger.Info("Migrations completed successfully")

	// Initialiser les repositories
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	emailRepo := repository.NewEmailRepository(db)

	// Initialiser les services avec sécurité renforcée
	authService := services.NewAuthService(userRepo, logger, cfg.JWT.Secret)
	accountService := services.NewAccountService(accountRepo, logger, cfg.Encryption.Key)
	mailService := services.NewMailService(emailRepo, accountService, logger)
	oauth2Service := utils.NewOAuth2Service(cfg, logger)

	// Initialiser les middlewares
	authMiddleware := middleware.NewAuthMiddleware(userRepo, authService, logger)

	// Créer le multiplexeur HTTP
	mux := http.NewServeMux()

	// Enregistrer toutes les routes avec les nouveaux services
	api.RegisterRoutes(mux, cfg, logger, authService, authMiddleware, accountService, mailService, oauth2Service)

	// Démarrer le serveur
	addr := ":" + cfg.Server.Port
	logger.Info(fmt.Sprintf("Server running on %s", addr))
	logger.Info("JWT authentication enabled")
	logger.Info("Multi-account email management ready")

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>Bienvenue sur Tamis API 🧹🚀</h1><p>Le serveur tourne sur le port %s</p><ul><li>Auth: /api/auth/</li><li>Accounts: /api/accounts/</li><li>Mails: /api/mails/</li></ul>", cfg.Server.Port)
	})

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Server failed to start: %v", err))
	}
}
