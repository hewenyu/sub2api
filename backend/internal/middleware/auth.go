package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/admin"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

const (
	ContextKeyAdminID = "admin_id"
	ContextKeyAPIKey  = "api_key"
)

// AuthenticateAdmin validates JWT token from Authorization header
func AuthenticateAdmin(adminService admin.AdminService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		adminID, err := adminService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store admin ID in context
		c.Set(ContextKeyAdminID, adminID)
		c.Next()
	}
}

// AuthenticateAPIKey validates API key from Authorization header
func AuthenticateAPIKey(apiKeyRepo repository.APIKeyRepository, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API Key",
			})
			c.Abort()
			return
		}

		// Extract API key (format: "Bearer <api_key>" or direct)
		apiKey := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Hash API key for lookup
		keyHash := crypto.HashAPIKey(apiKey)

		// Get API key from database
		apiKeyObj, err := apiKeyRepo.GetByHash(c.Request.Context(), keyHash)
		if err != nil {
			logger.Warn("Invalid API key", zap.String("key_prefix", crypto.GetAPIKeyPrefix(apiKey)))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API Key",
			})
			c.Abort()
			return
		}

		// Check if API key is active
		if !apiKeyObj.IsActive {
			logger.Warn("Inactive API key", zap.Int64("api_key_id", apiKeyObj.ID))
			c.JSON(http.StatusForbidden, gin.H{
				"error": "API Key is inactive",
			})
			c.Abort()
			return
		}

		// Store API key object in context
		c.Set(ContextKeyAPIKey, apiKeyObj)
		c.Next()
	}
}

// GetAdminID extracts admin ID from context
func GetAdminID(c *gin.Context) int64 {
	adminID, exists := c.Get(ContextKeyAdminID)
	if !exists {
		return 0
	}
	//nolint:errcheck // Type assertion is safe here, panic on mismatch is acceptable
	return adminID.(int64)
}

// GetAPIKey extracts API key object from context
func GetAPIKey(c *gin.Context) *model.APIKey {
	apiKey, exists := c.Get(ContextKeyAPIKey)
	if !exists {
		return nil
	}
	//nolint:errcheck // Type assertion is safe here, panic on mismatch is acceptable
	return apiKey.(*model.APIKey)
}
