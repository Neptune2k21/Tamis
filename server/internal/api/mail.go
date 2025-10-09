package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tamis-server/internal/middleware"
	"tamis-server/internal/models"
	"tamis-server/internal/services"
	"tamis-server/internal/utils"
)

// listMailsHandler - Lister tous les mails consolidés de tous les comptes
func ListMailsHandler(mailService *services.MailService, logger *utils.Logger) http.HandlerFunc {
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

		// Construire les filtres depuis les paramètres de requête
		filter := buildEmailFilter(r)

		emails, totalCount, err := mailService.GetUserEmails(user.ID, filter)
		if err != nil {
			logger.Error("Failed to retrieve emails for user " + strconv.Itoa(user.ID) + ": " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve emails")
			return
		}

		logger.Info("Emails retrieved for user " + strconv.Itoa(user.ID) + " - Count: " + strconv.Itoa(len(emails)))
		utils.WriteSuccess(w, map[string]interface{}{
			"emails":      emails,
			"count":       len(emails),
			"total_count": totalCount,
			"filter":      filter,
		}, "Emails retrieved successfully")
	}
}

// mailActionHandler - Actions sur les mails (supprimer, archiver, marquer lu)
func MailActionHandler(mailService *services.MailService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		var req models.EmailActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}

		// Validation
		if len(req.EmailIDs) == 0 {
			utils.WriteError(w, http.StatusBadRequest, "Email IDs are required")
			return
		}

		if req.Action == "" {
			utils.WriteError(w, http.StatusBadRequest, "Action is required")
			return
		}

		// Exécuter l'action
		result, err := mailService.ExecuteEmailAction(user.ID, &req)
		if err != nil {
			logger.Error("Failed to execute mail action for user " + strconv.Itoa(user.ID) + ": " + err.Error())
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info("Mail action executed for user " + strconv.Itoa(user.ID) + " - Action: " + string(req.Action) + " - Emails: " + strconv.Itoa(len(req.EmailIDs)))
		utils.WriteSuccess(w, result, "Action executed successfully")
	}
}

// syncMailsHandler - Synchroniser les mails depuis les serveurs (force refresh)
func SyncMailsHandler(mailService *services.MailService, logger *utils.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok || user == nil {
			utils.WriteError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Paramètre optionnel pour forcer la synchronisation complète
		forceSync := r.URL.Query().Get("force") == "true"

		result, err := mailService.SyncUserEmails(user.ID, forceSync)
		if err != nil {
			logger.Error("Failed to sync emails for user " + strconv.Itoa(user.ID) + ": " + err.Error())
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logger.Info("Email sync completed for user " + strconv.Itoa(user.ID) + " - Synced: " + strconv.Itoa(result.SyncedCount))
		utils.WriteSuccess(w, result, "Email synchronization completed")
	}
}

// buildEmailFilter - Construire le filtre depuis les paramètres de requête
func buildEmailFilter(r *http.Request) *models.EmailFilter {
	filter := &models.EmailFilter{}

	// Provider
	if provider := r.URL.Query().Get("provider"); provider != "" {
		filter.Provider = models.EmailProvider(provider)
	}

	// From
	filter.From = r.URL.Query().Get("from")

	// Subject
	filter.Subject = r.URL.Query().Get("subject")

	// Is Read
	if isReadStr := r.URL.Query().Get("is_read"); isReadStr != "" {
		isRead := isReadStr == "true"
		filter.IsRead = &isRead
	}

	// Is Spam
	if isSpamStr := r.URL.Query().Get("is_spam"); isSpamStr != "" {
		isSpam := isSpamStr == "true"
		filter.IsSpam = &isSpam
	}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50 // Valeur par défaut
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}
