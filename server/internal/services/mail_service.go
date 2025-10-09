package services

import (
	"fmt"
	"tamis-server/internal/models"
	"tamis-server/internal/repository"
	"tamis-server/internal/utils"
	"time"
)

type MailService struct {
	emailRepo      *repository.EmailRepository
	accountService *AccountService
	logger         *utils.Logger
}

func NewMailService(emailRepo *repository.EmailRepository, accountService *AccountService, logger *utils.Logger) *MailService {
	return &MailService{
		emailRepo:      emailRepo,
		accountService: accountService,
		logger:         logger,
	}
}

// GetUserEmails - Récupérer tous les emails consolidés de l'utilisateur
func (s *MailService) GetUserEmails(userID int, filter *models.EmailFilter) ([]*models.Email, int, error) {
	// Récupérer les comptes de l'utilisateur
	accounts, err := s.accountService.GetUserAccounts(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user accounts: %w", err)
	}

	if len(accounts) == 0 {
		return []*models.Email{}, 0, nil
	}

	// Extraire les IDs des comptes actifs
	accountIDs := make([]int, 0, len(accounts))
	for _, account := range accounts {
		if account.IsActive {
			accountIDs = append(accountIDs, account.ID)
		}
	}

	// Récupérer les emails depuis la base
	emails, totalCount, err := s.emailRepo.GetByAccountIDsWithFilter(accountIDs, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve emails: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Retrieved %d emails for user %d", len(emails), userID))
	return emails, totalCount, nil
}

// ExecuteEmailAction - Exécuter une action sur des emails
func (s *MailService) ExecuteEmailAction(userID int, req *models.EmailActionRequest) (*models.EmailActionResult, error) {
	// Vérifier que les emails appartiennent à l'utilisateur
	if err := s.validateEmailOwnership(userID, req.EmailIDs); err != nil {
		return nil, err
	}

	result := &models.EmailActionResult{
		Action:       req.Action,
		ProcessedIDs: []string{},
		FailedIDs:    []string{},
		SuccessCount: 0,
		FailureCount: 0,
	}

	// Exécuter l'action selon le type
	switch req.Action {
	case models.ActionDelete:
		err := s.executeDeleteAction(req.EmailIDs, req.Force)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Delete action failed: %v", err))
			result.FailedIDs = req.EmailIDs
			result.FailureCount = len(req.EmailIDs)
		} else {
			result.ProcessedIDs = req.EmailIDs
			result.SuccessCount = len(req.EmailIDs)
		}

	case models.ActionMarkRead:
		err := s.executeMarkReadAction(req.EmailIDs, true)
		if err != nil {
			result.FailedIDs = req.EmailIDs
			result.FailureCount = len(req.EmailIDs)
		} else {
			result.ProcessedIDs = req.EmailIDs
			result.SuccessCount = len(req.EmailIDs)
		}

	case models.ActionMarkUnread:
		err := s.executeMarkReadAction(req.EmailIDs, false)
		if err != nil {
			result.FailedIDs = req.EmailIDs
			result.FailureCount = len(req.EmailIDs)
		} else {
			result.ProcessedIDs = req.EmailIDs
			result.SuccessCount = len(req.EmailIDs)
		}

	case models.ActionArchive:
		err := s.executeArchiveAction(req.EmailIDs)
		if err != nil {
			result.FailedIDs = req.EmailIDs
			result.FailureCount = len(req.EmailIDs)
		} else {
			result.ProcessedIDs = req.EmailIDs
			result.SuccessCount = len(req.EmailIDs)
		}

	default:
		return nil, fmt.Errorf("unsupported action: %s", req.Action)
	}

	s.logger.Info(fmt.Sprintf("Action %s executed for user %d - Success: %d, Failed: %d",
		req.Action, userID, result.SuccessCount, result.FailureCount))

	return result, nil
}

// SyncUserEmails - Synchroniser les emails depuis les serveurs
func (s *MailService) SyncUserEmails(userID int, forceSync bool) (*models.EmailSyncResult, error) {
	accounts, err := s.accountService.GetUserAccounts(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	result := &models.EmailSyncResult{
		SyncedCount:   0,
		FailedCount:   0,
		NewEmails:     0,
		UpdatedEmails: 0,
		Accounts:      []string{},
	}

	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		syncResult, err := s.syncAccountEmails(account, forceSync)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to sync account %d: %v", account.ID, err))
			result.FailedCount++
			continue
		}

		result.SyncedCount++
		result.NewEmails += syncResult.NewEmails
		result.UpdatedEmails += syncResult.UpdatedEmails
		result.Accounts = append(result.Accounts, account.Email)
	}

	s.logger.Info(fmt.Sprintf("Email sync completed for user %d - Accounts: %d, New: %d, Updated: %d",
		userID, result.SyncedCount, result.NewEmails, result.UpdatedEmails))

	return result, nil
}

// validateEmailOwnership - Vérifier que les emails appartiennent à l'utilisateur
func (s *MailService) validateEmailOwnership(userID int, emailIDs []string) error {
	for _, emailID := range emailIDs {
		email, err := s.emailRepo.GetByID(emailID)
		if err != nil {
			return fmt.Errorf("email %s not found", emailID)
		}

		// Vérifier que le compte associé appartient à l'utilisateur
		account, err := s.accountService.accountRepo.GetByID(email.AccountID)
		if err != nil || account.UserID != userID {
			return fmt.Errorf("unauthorized access to email %s", emailID)
		}
	}
	return nil
}

// executeDeleteAction - Supprimer des emails
func (s *MailService) executeDeleteAction(emailIDs []string, force bool) error {
	if force {
		// Suppression définitive
		return s.emailRepo.DeletePermanently(emailIDs)
	} else {
		// Marquer comme supprimé (soft delete)
		return s.emailRepo.MarkAsDeleted(emailIDs)
	}
}

// executeMarkReadAction - Marquer des emails comme lus/non lus
func (s *MailService) executeMarkReadAction(emailIDs []string, isRead bool) error {
	return s.emailRepo.UpdateReadStatus(emailIDs, isRead)
}

// executeArchiveAction - Archiver des emails
func (s *MailService) executeArchiveAction(emailIDs []string) error {
	return s.emailRepo.ArchiveEmails(emailIDs)
}

// syncAccountEmails - Synchroniser les emails d'un compte spécifique
func (s *MailService) syncAccountEmails(account *models.EmailAccount, forceSync bool) (*models.AccountSyncResult, error) {
	// Récupérer les tokens déchiffrés
	tokens, err := s.accountService.GetDecryptedToken(account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	// Vérifier si le token a expiré et le rafraîchir si nécessaire
	if time.Now().After(*tokens.ExpiresAt) {
		// Implémenter le refresh du token OAuth2
		s.logger.Info(fmt.Sprintf("Refreshing expired token for account %d", account.ID))
		// newTokens, err := s.refreshOAuth2Token(account, tokens.RefreshToken)
		// if err != nil {
		//     return nil, fmt.Errorf("failed to refresh token: %w", err)
		// }
		// tokens = newTokens
	}

	// Connecter au serveur IMAP/API du provider
	emailClient, err := s.createEmailClient(account.Provider, tokens.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create email client: %w", err)
	}

	// Récupérer les emails récents
	recentEmails, err := emailClient.FetchRecentEmails(100) // Limiter à 100 emails récents
	if err != nil {
		return nil, fmt.Errorf("failed to fetch emails: %w", err)
	}

	result := &models.AccountSyncResult{
		NewEmails:     0,
		UpdatedEmails: 0,
	}

	// Traiter chaque email
	for _, email := range recentEmails {
		email.AccountID = account.ID

		// Vérifier si l'email existe déjà
		existingEmail, _ := s.emailRepo.GetByMessageID(email.MessageID, account.ID)

		if existingEmail == nil {
			// Nouvel email
			_, err := s.emailRepo.Create(email)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Failed to save new email: %v", err))
				continue
			}
			result.NewEmails++
		} else {
			// Email existant, mettre à jour si nécessaire
			if s.emailNeedsUpdate(existingEmail, email) {
				err := s.emailRepo.Update(email)
				if err != nil {
					s.logger.Error(fmt.Sprintf("Failed to update email: %v", err))
					continue
				}
				result.UpdatedEmails++
			}
		}
	}

	return result, nil
}

// createEmailClient - Créer un client email selon le provider
func (s *MailService) createEmailClient(provider models.EmailProvider, accessToken string) (EmailClient, error) {
	// Factory pattern pour créer le bon client selon le provider
	switch provider {
	case models.ProviderGmail:
		return NewGmailClient(accessToken), nil
	case models.ProviderOutlook:
		return NewOutlookClient(accessToken), nil
	case models.ProviderYahoo:
		return NewYahooClient(accessToken), nil
	default:
		return NewGenericIMAPClient(accessToken), nil
	}
}

// emailNeedsUpdate - Vérifier si un email a besoin d'être mis à jour
func (s *MailService) emailNeedsUpdate(existing, new *models.Email) bool {
	return existing.IsRead != new.IsRead ||
		existing.IsSpam != new.IsSpam ||
		existing.IsDeleted != new.IsDeleted ||
		len(existing.Labels) != len(new.Labels)
}

// Interface pour les clients email
type EmailClient interface {
	FetchRecentEmails(limit int) ([]*models.Email, error)
	MarkAsRead(emailID string) error
	Delete(emailID string) error
	Archive(emailID string) error
}

// Implémentations des clients (simplifié)
type GmailClient struct {
	accessToken string
}

func NewGmailClient(accessToken string) *GmailClient {
	return &GmailClient{accessToken: accessToken}
}

func (c *GmailClient) FetchRecentEmails(limit int) ([]*models.Email, error) {
	// Implémentation Gmail API
	// Simulation pour l'exemple
	return []*models.Email{}, nil
}

func (c *GmailClient) MarkAsRead(emailID string) error {
	// Implémentation Gmail API
	return nil
}

func (c *GmailClient) Delete(emailID string) error {
	// Implémentation Gmail API
	return nil
}

func (c *GmailClient) Archive(emailID string) error {
	// Implémentation Gmail API
	return nil
}

// Clients similaires pour Outlook, Yahoo, etc.
type OutlookClient struct{ accessToken string }
type YahooClient struct{ accessToken string }
type GenericIMAPClient struct{ accessToken string }

func NewOutlookClient(token string) *OutlookClient { return &OutlookClient{accessToken: token} }
func NewYahooClient(token string) *YahooClient     { return &YahooClient{accessToken: token} }
func NewGenericIMAPClient(token string) *GenericIMAPClient {
	return &GenericIMAPClient{accessToken: token}
}

// Implémentations des méthodes pour chaque client...
func (c *OutlookClient) FetchRecentEmails(limit int) ([]*models.Email, error) {
	return []*models.Email{}, nil
}
func (c *OutlookClient) MarkAsRead(emailID string) error { return nil }
func (c *OutlookClient) Delete(emailID string) error     { return nil }
func (c *OutlookClient) Archive(emailID string) error    { return nil }

func (c *YahooClient) FetchRecentEmails(limit int) ([]*models.Email, error) {
	return []*models.Email{}, nil
}
func (c *YahooClient) MarkAsRead(emailID string) error { return nil }
func (c *YahooClient) Delete(emailID string) error     { return nil }
func (c *YahooClient) Archive(emailID string) error    { return nil }

func (c *GenericIMAPClient) FetchRecentEmails(limit int) ([]*models.Email, error) {
	return []*models.Email{}, nil
}
func (c *GenericIMAPClient) MarkAsRead(emailID string) error { return nil }
func (c *GenericIMAPClient) Delete(emailID string) error     { return nil }
func (c *GenericIMAPClient) Archive(emailID string) error    { return nil }
