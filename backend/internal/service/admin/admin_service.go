package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

type AdminService interface {
	Login(ctx context.Context, username, password string) (string, *model.Admin, error)
	ValidateToken(ctx context.Context, token string) (int64, error)
	CreateAdmin(ctx context.Context, username, password string) (*model.Admin, error)
	ChangePassword(ctx context.Context, adminID int64, oldPassword, newPassword string) error
}

type adminService struct {
	adminRepo repository.AdminRepository
	jwtSecret string
	tokenExp  time.Duration
	logger    *zap.Logger
}

type JWTClaims struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAdminService(
	adminRepo repository.AdminRepository,
	jwtSecret string,
	tokenExp time.Duration,
	logger *zap.Logger,
) AdminService {
	return &adminService{
		adminRepo: adminRepo,
		jwtSecret: jwtSecret,
		tokenExp:  tokenExp,
		logger:    logger,
	}
}

func (s *adminService) Login(ctx context.Context, username, password string) (string, *model.Admin, error) {
	// Get admin by username
	admin, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Warn("Admin not found", zap.String("username", username))
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if compareErr := crypto.BcryptCompare(admin.PasswordHash, password); compareErr != nil {
		s.logger.Warn("Password mismatch", zap.String("username", username))
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(admin.ID, admin.Username)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("Admin logged in", zap.String("username", username), zap.Int64("admin_id", admin.ID))

	return token, admin, nil
}

func (s *adminService) ValidateToken(ctx context.Context, tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	return claims.AdminID, nil
}

func (s *adminService) CreateAdmin(ctx context.Context, username, password string) (*model.Admin, error) {
	// Check if admin already exists
	existing, _ := s.adminRepo.GetByUsername(ctx, username)
	if existing != nil {
		return nil, fmt.Errorf("admin already exists")
	}

	// Hash password
	passwordHash, err := crypto.BcryptHash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin
	admin := &model.Admin{
		Username:     username,
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	s.logger.Info("Admin created", zap.String("username", username), zap.Int64("admin_id", admin.ID))

	return admin, nil
}

func (s *adminService) ChangePassword(ctx context.Context, adminID int64, oldPassword, newPassword string) error {
	// Get admin
	admin, err := s.adminRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}

	// Verify old password
	if compareErr := crypto.BcryptCompare(admin.PasswordHash, oldPassword); compareErr != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	newPasswordHash, err := crypto.BcryptHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	updates := map[string]interface{}{
		"password_hash": newPasswordHash,
	}

	if err := s.adminRepo.Update(ctx, adminID, updates); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info("Password changed", zap.Int64("admin_id", adminID))

	return nil
}

func (s *adminService) generateToken(adminID int64, username string) (string, error) {
	claims := &JWTClaims{
		AdminID:  adminID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
