package services

import (
	"fmt"
	"tamis-server/internal/models"
	"tamis-server/internal/repository"
	"tamis-server/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	logger    *utils.Logger
	jwtSecret []byte
}

// Claims personnalisés pour JWT
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repository.UserRepository, logger *utils.Logger, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		logger:    logger,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register - Inscription
func (s *AuthService) Register(req *models.CreateUserRequest) (*models.User, error) {
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("email, username and password are required")
	}

	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	// Vérifier si l'utilisateur existe déjà
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// ✅ Hasher le VRAI mot de passe de l'utilisateur
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to hash password: %v", err))
		return nil, fmt.Errorf("failed to hash password")
	}

	// Créer l'utilisateur
	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
	}

	createdUser, err := s.userRepo.Create(user, string(passwordHash))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create user: %v", err))
		return nil, fmt.Errorf("failed to create user")
	}

	s.logger.Info(fmt.Sprintf("User registered: %s (ID: %d)", createdUser.Email, createdUser.ID))
	return createdUser, nil
}

// Login - Connexion avec génération de JWT
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// Validation
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Login attempt with non-existent email: %s", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Récupérer le hash du mot de passe
	passwordHash, err := s.userRepo.GetPasswordHash(req.Email)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get password hash for user %s: %v", req.Email, err))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Invalid password attempt for user: %s", user.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// ✅ Générer un JWT sécurisé
	token, err := s.GenerateJWT(user)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to generate JWT for user %s: %v", user.Email, err))
		return nil, fmt.Errorf("failed to generate token")
	}

	s.logger.Info(fmt.Sprintf("User logged in: %s (ID: %d)", user.Email, user.ID))

	return &models.LoginResponse{
		User:  *user,
		Token: token,
	}, nil
}

// GenerateJWT - Génère un token JWT signé et sécurisé
func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
	// Définir l'expiration (24 heures)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Créer les claims
	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tamis-server",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Créer le token avec HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signer le token
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT - Valide et parse un token JWT
func (s *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	// Parser le token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extraire les claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// RefreshToken - Génère un nouveau token si l'ancien est valide
func (s *AuthService) RefreshToken(oldToken string) (string, error) {
	// Valider le token existant
	claims, err := s.ValidateJWT(oldToken)
	if err != nil {
		return "", fmt.Errorf("invalid token")
	}

	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	// Générer un nouveau token
	newToken, err := s.GenerateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token")
	}

	s.logger.Info(fmt.Sprintf("Token refreshed for user: %s", user.Email))
	return newToken, nil
}
