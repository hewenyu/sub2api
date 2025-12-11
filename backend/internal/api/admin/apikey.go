package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

type APIKeyHandler struct {
	apiKeyRepo repository.APIKeyRepository
	logger     *zap.Logger
}

func NewAPIKeyHandler(apiKeyRepo repository.APIKeyRepository, logger *zap.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyRepo: apiKeyRepo,
		logger:     logger,
	}
}

type CreateAPIKeyRequest struct {
	Name                    string   `json:"name" binding:"required"`
	Permissions             []string `json:"permissions"`
	MaxConcurrentRequests   int      `json:"max_concurrent_requests"`
	RateLimitPerMinute      int      `json:"rate_limit_per_minute"`
	RateLimitPerHour        int      `json:"rate_limit_per_hour"`
	RateLimitPerDay         int      `json:"rate_limit_per_day"`
	DailyCostLimit          float64  `json:"daily_cost_limit"`
	WeeklyCostLimit         float64  `json:"weekly_cost_limit"`
	MonthlyCostLimit        float64  `json:"monthly_cost_limit"`
	TotalCostLimit          float64  `json:"total_cost_limit"`
	EnableModelRestriction  bool     `json:"enable_model_restriction"`
	RestrictedModels        []string `json:"restricted_models"`
	EnableClientRestriction bool     `json:"enable_client_restriction"`
	AllowedClients          []string `json:"allowed_clients"`
}

type CreateAPIKeyResponse struct {
	APIKey       string     `json:"api_key"`
	APIKeyObject APIKeyInfo `json:"api_key_object"`
}

type APIKeyInfo struct {
	ID                      int64    `json:"id"`
	KeyPrefix               string   `json:"key_prefix"`
	Name                    string   `json:"name"`
	IsActive                bool     `json:"is_active"`
	ExpiresAt               *int64   `json:"expires_at,omitempty"`
	MaxConcurrentRequests   int      `json:"max_concurrent_requests"`
	BoundCodexAccountID     *int64   `json:"bound_codex_account_id,omitempty"`
	RateLimitPerMinute      int      `json:"rate_limit_per_minute"`
	RateLimitPerHour        int      `json:"rate_limit_per_hour"`
	RateLimitPerDay         int      `json:"rate_limit_per_day"`
	DailyCostLimit          float64  `json:"daily_cost_limit"`
	WeeklyCostLimit         float64  `json:"weekly_cost_limit"`
	MonthlyCostLimit        float64  `json:"monthly_cost_limit"`
	TotalCostLimit          float64  `json:"total_cost_limit"`
	EnableModelRestriction  bool     `json:"enable_model_restriction"`
	RestrictedModels        []string `json:"restricted_models"`
	EnableClientRestriction bool     `json:"enable_client_restriction"`
	AllowedClients          []string `json:"allowed_clients"`
	TotalRequests           int64    `json:"total_requests"`
	TotalTokens             int64    `json:"total_tokens"`
	TotalCost               float64  `json:"total_cost"`
	CreatedAt               int64    `json:"created_at"`
	UpdatedAt               int64    `json:"updated_at"`
}

func (h *APIKeyHandler) toAPIKeyInfo(apiKey *model.APIKey) APIKeyInfo {
	info := APIKeyInfo{
		ID:                      apiKey.ID,
		KeyPrefix:               apiKey.KeyPrefix,
		Name:                    apiKey.Name,
		IsActive:                apiKey.IsActive,
		MaxConcurrentRequests:   apiKey.MaxConcurrentRequests,
		BoundCodexAccountID:     apiKey.BoundCodexAccountID,
		RateLimitPerMinute:      apiKey.RateLimitPerMinute,
		RateLimitPerHour:        apiKey.RateLimitPerHour,
		RateLimitPerDay:         apiKey.RateLimitPerDay,
		DailyCostLimit:          apiKey.DailyCostLimit,
		WeeklyCostLimit:         apiKey.WeeklyCostLimit,
		MonthlyCostLimit:        apiKey.MonthlyCostLimit,
		TotalCostLimit:          apiKey.TotalCostLimit,
		EnableModelRestriction:  apiKey.EnableModelRestriction,
		RestrictedModels:        []string(apiKey.RestrictedModels),
		EnableClientRestriction: apiKey.EnableClientRestriction,
		AllowedClients:          []string(apiKey.AllowedClients),
		TotalRequests:           apiKey.TotalRequests,
		TotalTokens:             apiKey.TotalTokens,
		TotalCost:               apiKey.TotalCost,
		CreatedAt:               apiKey.CreatedAt.Unix(),
		UpdatedAt:               apiKey.UpdatedAt.Unix(),
	}

	if apiKey.ExpiresAt != nil {
		expiresAt := apiKey.ExpiresAt.Unix()
		info.ExpiresAt = &expiresAt
	}

	return info
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create API key request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Generate API key
	apiKey, err := crypto.GenerateAPIKey()
	if err != nil {
		h.logger.Error("Failed to generate API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate API key",
		})
		return
	}

	keyHash := crypto.HashAPIKey(apiKey)
	keyPrefix := crypto.GetAPIKeyPrefix(apiKey)

	// Set defaults
	if req.MaxConcurrentRequests == 0 {
		req.MaxConcurrentRequests = 5
	}
	if req.RateLimitPerMinute == 0 {
		req.RateLimitPerMinute = 60
	}

	// Create API key object
	apiKeyObj := &model.APIKey{
		KeyHash:                 keyHash,
		KeyPrefix:               keyPrefix,
		Name:                    req.Name,
		IsActive:                true,
		MaxConcurrentRequests:   req.MaxConcurrentRequests,
		RateLimitPerMinute:      req.RateLimitPerMinute,
		RateLimitPerHour:        req.RateLimitPerHour,
		RateLimitPerDay:         req.RateLimitPerDay,
		DailyCostLimit:          req.DailyCostLimit,
		WeeklyCostLimit:         req.WeeklyCostLimit,
		MonthlyCostLimit:        req.MonthlyCostLimit,
		TotalCostLimit:          req.TotalCostLimit,
		EnableModelRestriction:  req.EnableModelRestriction,
		EnableClientRestriction: req.EnableClientRestriction,
	}

	// Save to database
	if err := h.apiKeyRepo.Create(c.Request.Context(), apiKeyObj); err != nil {
		h.logger.Error("Failed to create API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create API key",
		})
		return
	}

	h.logger.Info("API key created", zap.Int64("api_key_id", apiKeyObj.ID), zap.String("name", req.Name))

	// Build response
	resp := CreateAPIKeyResponse{
		APIKey:       apiKey,
		APIKeyObject: h.toAPIKeyInfo(apiKeyObj),
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get API keys from database
	apiKeys, err := h.apiKeyRepo.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		h.logger.Error("Failed to list API keys", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list API keys",
		})
		return
	}

	// Build response
	items := make([]APIKeyInfo, len(apiKeys))
	for i, key := range apiKeys {
		items[i] = h.toAPIKeyInfo(key)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": items,
			"pagination": gin.H{
				"page":      page,
				"page_size": pageSize,
			},
		},
	})
}

func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API key ID",
		})
		return
	}

	apiKey, err := h.apiKeyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get API key", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API key not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toAPIKeyInfo(apiKey),
	})
}

func (h *APIKeyHandler) UpdateAPIKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API key ID",
		})
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	updates := map[string]interface{}{
		"name":                      req.Name,
		"max_concurrent_requests":   req.MaxConcurrentRequests,
		"rate_limit_per_minute":     req.RateLimitPerMinute,
		"rate_limit_per_hour":       req.RateLimitPerHour,
		"rate_limit_per_day":        req.RateLimitPerDay,
		"daily_cost_limit":          req.DailyCostLimit,
		"weekly_cost_limit":         req.WeeklyCostLimit,
		"monthly_cost_limit":        req.MonthlyCostLimit,
		"total_cost_limit":          req.TotalCostLimit,
		"enable_model_restriction":  req.EnableModelRestriction,
		"enable_client_restriction": req.EnableClientRestriction,
	}

	if err := h.apiKeyRepo.Update(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("Failed to update API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update API key",
		})
		return
	}

	h.logger.Info("API key updated", zap.Int64("api_key_id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "API key updated successfully",
		"data": gin.H{
			"message": "API key updated successfully",
		},
	})
}

func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API key ID",
		})
		return
	}

	if err := h.apiKeyRepo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete API key",
		})
		return
	}

	h.logger.Info("API key deleted", zap.Int64("api_key_id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "API key deleted successfully",
		"data": gin.H{
			"message": "API key deleted successfully",
		},
	})
}

func (h *APIKeyHandler) ToggleAPIKeyStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API key ID",
		})
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	updates := map[string]interface{}{
		"is_active": req.IsActive,
	}

	if err := h.apiKeyRepo.Update(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("Failed to toggle API key status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to toggle API key status",
		})
		return
	}

	h.logger.Info("API key status toggled", zap.Int64("api_key_id", id), zap.Bool("is_active", req.IsActive))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "API key status updated successfully",
		"data": gin.H{
			"is_active": req.IsActive,
		},
	})
}
