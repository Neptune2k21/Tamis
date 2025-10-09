package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	OAuth2     OAuth2Config
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host string
	Port string
}

type JWTConfig struct {
	Secret string
}

type EncryptionConfig struct {
	Key string // Clé AES-256 pour chiffrer les tokens OAuth2
}

type OAuth2Config struct {
	Gmail struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Outlook struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Yahoo struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("GO_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5434"),
			User:     getEnv("DB_USER", "tamis"),
			Password: getEnv("DB_PASSWORD", "tamis_dev_password"),
			DBName:   getEnv("DB_NAME", "tamis_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", "tamis-super-secret-encryption-key-32-bytes"),
		},
		OAuth2: OAuth2Config{
			Gmail: struct {
				ClientID     string
				ClientSecret string
				RedirectURL  string
			}{
				ClientID:     getEnv("GMAIL_CLIENT_ID", ""),
				ClientSecret: getEnv("GMAIL_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GMAIL_REDIRECT_URL", "http://localhost:8080/auth/gmail/callback"),
			},
			Outlook: struct {
				ClientID     string
				ClientSecret string
				RedirectURL  string
			}{
				ClientID:     getEnv("OUTLOOK_CLIENT_ID", ""),
				ClientSecret: getEnv("OUTLOOK_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("OUTLOOK_REDIRECT_URL", "http://localhost:8080/auth/outlook/callback"),
			},
			Yahoo: struct {
				ClientID     string
				ClientSecret string
				RedirectURL  string
			}{
				ClientID:     getEnv("YAHOO_CLIENT_ID", ""),
				ClientSecret: getEnv("YAHOO_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("YAHOO_REDIRECT_URL", "http://localhost:8080/auth/yahoo/callback"),
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
