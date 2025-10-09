package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"tamis-server/internal/config"
	"tamis-server/internal/models"
	"time"
)

type OAuth2Service struct {
	config *config.Config
	logger *Logger
}

func NewOAuth2Service(config *config.Config, logger *Logger) *OAuth2Service {
	return &OAuth2Service{
		config: config,
		logger: logger,
	}
}

// GetGoogleAuthURL - Générer l'URL d'autorisation Google
func (s *OAuth2Service) GetGoogleAuthURL(state string) string {
	baseURL := "https://accounts.google.com/o/oauth2/v2/auth"
	params := url.Values{}
	params.Add("client_id", s.config.OAuth2.Gmail.ClientID)
	params.Add("redirect_uri", s.config.OAuth2.Gmail.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/gmail.readonly https://www.googleapis.com/auth/userinfo.email")
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// ExchangeCodeForTokens - Échanger le code d'autorisation contre des tokens
func (s *OAuth2Service) ExchangeCodeForTokens(code string) (*models.OAuth2Token, error) {
	data := url.Values{}
	data.Set("client_id", s.config.OAuth2.Gmail.ClientID)
	data.Set("client_secret", s.config.OAuth2.Gmail.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", s.config.OAuth2.Gmail.RedirectURL)

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth2 error: %s", string(body))
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &models.OAuth2Token{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}, nil
}

// RefreshGoogleToken - Rafraîchir un token Google
func (s *OAuth2Service) RefreshGoogleToken(refreshToken string) (*models.OAuth2Token, error) {
	data := url.Values{}
	data.Set("client_id", s.config.OAuth2.Gmail.ClientID)
	data.Set("client_secret", s.config.OAuth2.Gmail.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token error: %s", string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &models.OAuth2Token{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenResponse.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}, nil
}

// GetUserInfo - Récupérer les informations utilisateur depuis Google
func (s *OAuth2Service) GetUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info, status: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
