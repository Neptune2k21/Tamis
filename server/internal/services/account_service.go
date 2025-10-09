package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"tamis-server/internal/models"
	"tamis-server/internal/repository"
	"tamis-server/internal/utils"
	"time"
)

type AccountService struct {
	accountRepo   *repository.AccountRepository
	logger        *utils.Logger
	encryptionKey []byte // Clé de chiffrement pour les tokens
}

func NewAccountService(accountRepo *repository.AccountRepository, logger *utils.Logger, encryptionKey string) *AccountService {
	// Utiliser une clé de 32 bytes pour AES-256
	key := make([]byte, 32)
	copy(key, []byte(encryptionKey))

	return &AccountService{
		accountRepo:   accountRepo,
		logger:        logger,
		encryptionKey: key,
	}
}

// AddAccount - Ajouter un compte email via OAuth2 (sécurisé)
func (s *AccountService) AddAccount(userID int, req *models.CreateEmailAccountRequest) (*models.EmailAccount, error) {
	// Vérifier si le compte existe déjà
	existingAccount, _ := s.accountRepo.GetByUserAndEmail(userID, req.Email)
	if existingAccount != nil {
		return nil, fmt.Errorf("account already exists for this email")
	}

	// Initialiser le flux OAuth2 selon le provider
	oauthToken, err := s.initiateOAuth2Flow(req.Provider, req.Email)
	if err != nil {
		s.logger.Error(fmt.Sprintf("OAuth2 flow failed for %s: %v", req.Email, err))
		return nil, fmt.Errorf("failed to authenticate with email provider")
	}

	// Chiffrer les tokens avant stockage
	encryptedAccessToken, err := s.encryptToken(oauthToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token")
	}

	encryptedRefreshToken, err := s.encryptToken(oauthToken.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt refresh token")
	}

	// Créer l'account
	account := &models.EmailAccount{
		UserID:         userID,
		Provider:       req.Provider,
		Email:          req.Email,
		DisplayName:    req.DisplayName,
		AccessToken:    encryptedAccessToken,
		RefreshToken:   encryptedRefreshToken,
		TokenExpiresAt: &oauthToken.Expiry,
		IsActive:       true,
	}

	createdAccount, err := s.accountRepo.Create(account)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create account in DB: %v", err))
		return nil, fmt.Errorf("failed to save account")
	}

	s.logger.Info(fmt.Sprintf("Account added successfully: %s (Provider: %s)", req.Email, req.Provider))
	return createdAccount, nil
}

// GetUserAccounts - Récupérer tous les comptes de l'utilisateur
func (s *AccountService) GetUserAccounts(userID int) ([]*models.EmailAccount, error) {
	accounts, err := s.accountRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve accounts: %w", err)
	}

	// Ne pas retourner les tokens dans la réponse (sécurité)
	for _, account := range accounts {
		account.AccessToken = ""
		account.RefreshToken = ""
	}

	return accounts, nil
}

// RemoveAccount - Supprimer un compte (révocation OAuth2)
func (s *AccountService) RemoveAccount(userID, accountID int) error {
	// Vérifier que le compte appartient à l'utilisateur
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return fmt.Errorf("unauthorized: account does not belong to user")
	}

	// Révoquer les tokens OAuth2 auprès du provider
	if err := s.revokeOAuth2Token(account); err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to revoke OAuth2 token for account %d: %v", accountID, err))
		// Continue même si la révocation échoue
	}

	// Supprimer de la base de données
	if err := s.accountRepo.Delete(accountID); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Account %d removed for user %d", accountID, userID))
	return nil
}

// GetDecryptedToken - Récupérer un token déchiffré (usage interne uniquement)
func (s *AccountService) GetDecryptedToken(accountID int) (*models.DecryptedTokens, error) {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.decryptToken(account.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token")
	}

	refreshToken, err := s.decryptToken(account.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token")
	}

	return &models.DecryptedTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    account.TokenExpiresAt,
	}, nil
}

// encryptToken - Chiffrer un token avec AES-256-GCM
func (s *AccountService) encryptToken(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptToken - Déchiffrer un token
func (s *AccountService) decryptToken(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// initiateOAuth2Flow - Initier le flux OAuth2 (simplifié pour l'exemple)
func (s *AccountService) initiateOAuth2Flow(provider models.EmailProvider, email string) (*models.OAuth2Token, error) {
	// Dans une implémentation réelle, ceci redirigerait vers le provider OAuth2
	// et récupérerait les tokens après autorisation de l'utilisateur

	// Simulation pour l'exemple
	return &models.OAuth2Token{
		AccessToken:  "simulated_access_token_for_" + email,
		RefreshToken: "simulated_refresh_token_for_" + email,
		Expiry:       time.Now().Add(1 * time.Hour),
	}, nil
}

// revokeOAuth2Token - Révoquer les tokens OAuth2
func (s *AccountService) revokeOAuth2Token(account *models.EmailAccount) error {
	// Implémentation spécifique selon le provider (Gmail, Yahoo, Outlook)
	s.logger.Info(fmt.Sprintf("Revoking OAuth2 tokens for account %d (Provider: %s)", account.ID, account.Provider))

	// Dans une implémentation réelle, faire l'appel API de révocation
	return nil
}

// AddAccountWithTokens - Ajouter un compte avec des tokens OAuth2 existants
func (s *AccountService) AddAccountWithTokens(userID int, req *models.CreateEmailAccountRequest, tokens *models.OAuth2Token) (*models.EmailAccount, error) {
	// Vérifier si le compte existe déjà
	existingAccount, _ := s.accountRepo.GetByUserAndEmail(userID, req.Email)
	if existingAccount != nil {
		return nil, fmt.Errorf("account already exists for this email")
	}

	// Chiffrer les tokens avant stockage
	encryptedAccessToken, err := s.encryptToken(tokens.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token")
	}

	encryptedRefreshToken, err := s.encryptToken(tokens.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt refresh token")
	}

	// Créer l'account
	account := &models.EmailAccount{
		UserID:         userID,
		Provider:       req.Provider,
		Email:          req.Email,
		DisplayName:    req.DisplayName,
		AccessToken:    encryptedAccessToken,
		RefreshToken:   encryptedRefreshToken,
		TokenExpiresAt: &tokens.Expiry,
		IsActive:       true,
	}

	createdAccount, err := s.accountRepo.Create(account)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create account in DB: %v", err))
		return nil, fmt.Errorf("failed to save account")
	}

	s.logger.Info(fmt.Sprintf("Account added successfully: %s (Provider: %s)", req.Email, req.Provider))
	return createdAccount, nil
}
