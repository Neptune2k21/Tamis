package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tamis-server/internal/middleware"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

// addAccountHandler - Ajouter un compte email via OAuth2
func addAccountHandler(accountService *services.AccountService, logger *utils.Logger) http.HandlerFunc {
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

		var req models.CreateEmailAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}

		// Validation des données
		if req.Email == "" || req.Provider == "" || req.DisplayName == "" {
			utils.WriteError(w, http.StatusBadRequest, "Email, provider and display name are required")
			return
		}

		// Vérifier que le provider est supporté
		if !isValidProvider(req.Provider) {
			utils.WriteError(w, http.StatusBadRequest, "Unsupported email provider")
			return
		}

		// Ajouter le compte via OAuth2 (sécurisé)
		account, err := accountService.AddAccount(user.ID, &req)
		if err != nil {
			logger.Error("Failed to add account for user " + strconv.Itoa(user.ID) + ": " + err.Error())
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info("Account added successfully for user " + strconv.Itoa(user.ID) + " - Provider: " + string(req.Provider))
		utils.WriteSuccess(w, account, "Email account added successfully")
	}
}

// listAccountsHandler - Lister tous les comptes email de l'utilisateur
func listAccountsHandler(accountService *services.AccountService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		accounts, err := accountService.GetUserAccounts(user.ID)
		if err != nil {
			logger.Error("Failed to retrieve accounts for user " + strconv.Itoa(user.ID) + ": " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve accounts")
			return
		}

		logger.Info("Accounts retrieved for user " + strconv.Itoa(user.ID) + " - Count: " + strconv.Itoa(len(accounts)))
		utils.WriteSuccess(w, map[string]interface{}{
			"accounts": accounts,
			"count":    len(accounts),
		}, "Accounts retrieved successfully")
	}
}

// deleteAccountHandler - Supprimer un compte email (déconnexion OAuth2)
func deleteAccountHandler(accountService *services.AccountService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Récupérer l'ID du compte depuis l'URL
		accountIDStr := r.URL.Query().Get("account_id")
		if accountIDStr == "" {
			utils.WriteError(w, http.StatusBadRequest, "Account ID is required")
			return
		}

		accountID, err := strconv.Atoi(accountIDStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid account ID")
			return
		}

		err = accountService.RemoveAccount(user.ID, accountID)
		if err != nil {
			logger.Error("Failed to remove account " + accountIDStr + " for user " + strconv.Itoa(user.ID) + ": " + err.Error())
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info("Account " + accountIDStr + " removed for user " + strconv.Itoa(user.ID))
		utils.WriteSuccess(w, nil, "Account removed successfully")
	}
}

// isValidProvider - Vérifier si le provider est supporté
func isValidProvider(provider models.EmailProvider) bool {
	validProviders := []models.EmailProvider{
		models.ProviderGmail,
		models.ProviderYahoo,
		models.ProviderOutlook,
		models.ProviderOther,
	}

	for _, validProvider := range validProviders {
		if provider == validProvider {
			return true
		}
	}
	return false
}
