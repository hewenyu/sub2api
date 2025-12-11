package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/account"
)

// CodexAccountHandler handles Codex account management endpoints.
type CodexAccountHandler struct {
	service account.CodexAccountService
	logger  *zap.Logger
}

// NewCodexAccountHandler creates a new Codex account handler.
func NewCodexAccountHandler(service account.CodexAccountService, logger *zap.Logger) *CodexAccountHandler {
	return &CodexAccountHandler{
		service: service,
		logger:  logger,
	}
}

// GenerateAuthURLRequest represents a request to generate OAuth URL.
type GenerateAuthURLRequest struct {
	CallbackPort int `json:"callback_port" binding:"required,min=1,max=65535"`
}

// AuthURLResponse represents OAuth authorization URL response.
type AuthURLResponse struct {
	AuthURL     string `json:"auth_url"`
	CallbackURL string `json:"callback_url"`
	State       string `json:"state"`
}

// VerifyAuthRequest represents OAuth verification request.
type VerifyAuthRequest struct {
	Code    string                            `json:"code" binding:"required"`
	State   string                            `json:"state" binding:"required"`
	Account account.CreateCodexAccountRequest `json:"account" binding:"required"`
}

// CreateCodexAccountRequest represents account creation request.
type CreateCodexAccountRequest struct {
	Name            string  `json:"name" binding:"required"`
	AccountType     string  `json:"account_type" binding:"required,oneof=openai-oauth openai-responses"`
	Email           *string `json:"email"`
	APIKey          *string `json:"api_key"`
	BaseAPI         string  `json:"base_api"`
	CustomUserAgent *string `json:"custom_user_agent"`
	DailyQuota      float64 `json:"daily_quota"`
	QuotaResetTime  string  `json:"quota_reset_time"`
	Priority        int     `json:"priority"`
	Schedulable     bool    `json:"schedulable"`
	ProxyName       *string `json:"proxy_name"`
}

// UpdateCodexAccountRequest represents account update request.
type UpdateCodexAccountRequest struct {
	Name            *string  `json:"name"`
	BaseAPI         *string  `json:"base_api"`
	CustomUserAgent *string  `json:"custom_user_agent"`
	DailyQuota      *float64 `json:"daily_quota"`
	QuotaResetTime  *string  `json:"quota_reset_time"`
	Priority        *int     `json:"priority"`
	Schedulable     *bool    `json:"schedulable"`
	ProxyName       *string  `json:"proxy_name"`
}

// CodexAccountInfo represents account information response.
type CodexAccountInfo struct {
	ID                    int64   `json:"id"`
	Name                  string  `json:"name"`
	AccountType           string  `json:"account_type"`
	Email                 *string `json:"email,omitempty"`
	ChatGPTAccountID      *string `json:"chatgpt_account_id,omitempty"`
	ChatGPTUserID         *string `json:"chatgpt_user_id,omitempty"`
	OrganizationID        *string `json:"organization_id,omitempty"`
	OrganizationRole      *string `json:"organization_role,omitempty"`
	OrganizationTitle     *string `json:"organization_title,omitempty"`
	BaseAPI               string  `json:"base_api"`
	SubscriptionLevel     *string `json:"subscription_level,omitempty"`
	SubscriptionExpiresAt *int64  `json:"subscription_expires_at,omitempty"`
	DailyQuota            float64 `json:"daily_quota"`
	DailyUsage            float64 `json:"daily_usage"`
	IsActive              bool    `json:"is_active"`
	Schedulable           bool    `json:"schedulable"`
	Priority              int     `json:"priority"`
	RateLimitedUntil      *int64  `json:"rate_limited_until,omitempty"`
	ProxyName             *string `json:"proxy_name,omitempty"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
	LastUsedAt            *int64  `json:"last_used_at,omitempty"`
}

// ListCodexAccountsResponse represents list response.
type ListCodexAccountsResponse struct {
	Items      []CodexAccountInfo `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// GenerateAuthURL generates OAuth authorization URL.
// @Summary Generate OAuth authorization URL
// @Tags Codex Accounts
// @Accept json
// @Produce json
// @Param request body GenerateAuthURLRequest true "Request body"
// @Success 200 {object} AuthURLResponse
// @Router /admin/codex-accounts/generate-auth-url [post]
func (h *CodexAccountHandler) GenerateAuthURL(c *gin.Context) {
	var req GenerateAuthURLRequest

	// Bind and validate the request with enhanced error handling
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err, h.logger)
		return
	}

	h.logger.Info("Generating OAuth authorization URL",
		zap.Int("callback_port", req.CallbackPort),
	)

	authURL, callbackURL, state, err := h.service.GenerateAuthURL(c.Request.Context(), req.CallbackPort)
	if err != nil {
		h.logger.Error("Failed to generate auth URL",
			zap.Error(err),
			zap.Int("callback_port", req.CallbackPort),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate authorization URL",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("OAuth authorization URL generated successfully",
		zap.String("state", state),
		zap.Int("state_length", len(state)),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": AuthURLResponse{
			AuthURL:     authURL,
			CallbackURL: callbackURL,
			State:       state,
		},
	})
}

// VerifyAuth verifies OAuth authorization and creates account.
// @Summary Verify OAuth and create account
// @Tags Codex Accounts
// @Accept json
// @Produce json
// @Param request body VerifyAuthRequest true "Request body"
// @Success 200 {object} CodexAccountInfo
// @Router /admin/codex-accounts/verify-auth [post]
func (h *CodexAccountHandler) VerifyAuth(c *gin.Context) {
	var req VerifyAuthRequest

	// Bind and validate the request with enhanced error handling
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err, h.logger)
		return
	}

	// Additional validation for OAuth parameters
	if err := validateOAuthCode(req.Code); err != nil {
		h.logger.Warn("Invalid OAuth authorization code",
			zap.Error(err),
			zap.Int("code_length", len(req.Code)),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid authorization code",
			"message": err.Error(),
		})
		return
	}

	if err := validateOAuthState(req.State); err != nil {
		h.logger.Warn("Invalid OAuth state parameter",
			zap.Error(err),
			zap.Int("state_length", len(req.State)),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid state parameter",
			"message": err.Error(),
		})
		return
	}

	// Defensive check: Ensure account data is not nil
	// This should not happen due to binding validation, but we check for safety
	if req.Account.Name == "" {
		h.logger.Warn("Missing account name in verify auth request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid account data",
			"message": "Account name is required",
		})
		return
	}

	// Verify account_type is explicitly set
	// This was the source of the original bug
	if req.Account.AccountType == "" {
		h.logger.Warn("Missing account_type in verify auth request",
			zap.String("account_name", req.Account.Name),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing account_type field",
			"message": "The 'account_type' field is required and must be 'openai-oauth' or 'openai-responses'",
			"hint":    "For OAuth verification, account_type should typically be 'openai-oauth'",
		})
		return
	}

	// Log request with masked sensitive data for debugging
	h.logger.Info("Processing OAuth verification request",
		zap.String("account_name", req.Account.Name),
		zap.String("account_type", req.Account.AccountType),
		zap.Int("code_length", len(req.Code)),
		zap.Int("state_length", len(req.State)),
		zap.Bool("has_email", req.Account.Email != nil),
	)

	// Call service to verify auth and create account
	account, err := h.service.VerifyAuth(c.Request.Context(), req.Code, req.State, req.Account)
	if err != nil {
		h.logger.Error("Failed to verify OAuth authorization",
			zap.Error(err),
			zap.String("account_name", req.Account.Name),
			zap.String("account_type", req.Account.AccountType),
		)

		// Provide more specific error messages based on error type
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to verify authorization"

		errStr := err.Error()
		if contains(errStr, "invalid or expired OAuth state") {
			statusCode = http.StatusBadRequest
			errorMsg = "Invalid or expired OAuth state. Please restart the OAuth flow."
		} else if contains(errStr, "state expired") {
			statusCode = http.StatusBadRequest
			errorMsg = "OAuth state expired. Please restart the authorization process."
		} else if contains(errStr, "failed to exchange code") {
			statusCode = http.StatusBadRequest
			errorMsg = "Failed to exchange authorization code. The code may be invalid or expired."
		} else if contains(errStr, "failed to create account") {
			statusCode = http.StatusConflict
			errorMsg = "Failed to create account. An account with this configuration may already exist."
		}

		c.JSON(statusCode, gin.H{
			"error":   errorMsg,
			"details": errStr,
		})
		return
	}

	h.logger.Info("OAuth verification successful, account created",
		zap.Int64("account_id", account.ID),
		zap.String("account_name", account.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toAccountInfo(account),
	})
}

// contains is a helper function for case-insensitive substring matching
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// CreateAccount creates a new Codex account.
// @Summary Create Codex account
// @Tags Codex Accounts
// @Accept json
// @Produce json
// @Param request body CreateCodexAccountRequest true "Request body"
// @Success 201 {object} CodexAccountInfo
// @Router /admin/codex-accounts [post]
func (h *CodexAccountHandler) CreateAccount(c *gin.Context) {
	var req CreateCodexAccountRequest

	// Bind and validate the request with enhanced error handling
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err, h.logger)
		return
	}

	// Additional validation for account type specific requirements
	if req.AccountType == "openai-responses" && req.APIKey == nil {
		h.logger.Warn("Missing API key for openai-responses account",
			zap.String("account_name", req.Name),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required field",
			"message": "API key is required for 'openai-responses' account type",
		})
		return
	}

	h.logger.Info("Creating Codex account",
		zap.String("name", req.Name),
		zap.String("account_type", req.AccountType),
		zap.Bool("has_api_key", req.APIKey != nil),
	)

	serviceReq := &account.CreateCodexAccountRequest{
		Name:            req.Name,
		AccountType:     req.AccountType,
		Email:           req.Email,
		APIKey:          req.APIKey,
		BaseAPI:         req.BaseAPI,
		CustomUserAgent: req.CustomUserAgent,
		DailyQuota:      req.DailyQuota,
		QuotaResetTime:  req.QuotaResetTime,
		Priority:        req.Priority,
		Schedulable:     req.Schedulable,
		ProxyName:       req.ProxyName,
	}

	acc, err := h.service.CreateAccount(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to create account",
			zap.Error(err),
			zap.String("account_name", req.Name),
			zap.String("account_type", req.AccountType),
		)

		// Provide more specific error messages
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		if contains(errorMsg, "API key is required") {
			statusCode = http.StatusBadRequest
		} else if contains(errorMsg, "already exists") || contains(errorMsg, "duplicate") {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to create account",
			"message": errorMsg,
		})
		return
	}

	h.logger.Info("Codex account created successfully",
		zap.Int64("account_id", acc.ID),
		zap.String("account_name", acc.Name),
	)

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toAccountInfo(acc),
	})
}

// ListAccounts lists Codex accounts with filtering and pagination.
// @Summary List Codex accounts
// @Tags Codex Accounts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param account_type query string false "Filter by account type"
// @Param is_active query bool false "Filter by active status"
// @Param schedulable query bool false "Filter by schedulable status"
// @Success 200 {object} ListCodexAccountsResponse
// @Router /admin/codex-accounts [get]
func (h *CodexAccountHandler) ListAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filters := repository.CodexAccountFilters{}
	if accountType := c.Query("account_type"); accountType != "" {
		filters.AccountType = &accountType
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filters.IsActive = &isActive
	}
	if schedulableStr := c.Query("schedulable"); schedulableStr != "" {
		schedulable := schedulableStr == "true"
		filters.Schedulable = &schedulable
	}

	accounts, total, err := h.service.ListAccounts(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list accounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list accounts",
		})
		return
	}

	items := make([]CodexAccountInfo, len(accounts))
	for i, acc := range accounts {
		items[i] = h.toAccountInfo(acc)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": ListCodexAccountsResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// GetAccount retrieves a Codex account by ID.
// @Summary Get Codex account
// @Tags Codex Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} CodexAccountInfo
// @Router /admin/codex-accounts/{id} [get]
func (h *CodexAccountHandler) GetAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	acc, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.toAccountInfo(acc),
	})
}

// UpdateAccount updates a Codex account.
// @Summary Update Codex account
// @Tags Codex Accounts
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Param request body UpdateCodexAccountRequest true "Request body"
// @Success 200 {object} gin.H
// @Router /admin/codex-accounts/{id} [put]
func (h *CodexAccountHandler) UpdateAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	var req UpdateCodexAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update account request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.BaseAPI != nil {
		updates["base_api"] = *req.BaseAPI
	}
	if req.CustomUserAgent != nil {
		updates["custom_user_agent"] = *req.CustomUserAgent
	}
	if req.DailyQuota != nil {
		updates["daily_quota"] = *req.DailyQuota
	}
	if req.QuotaResetTime != nil {
		updates["quota_reset_time"] = *req.QuotaResetTime
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Schedulable != nil {
		updates["schedulable"] = *req.Schedulable
	}
	if req.ProxyName != nil {
		if *req.ProxyName == "" {
			updates["proxy_name"] = nil
		} else {
			updates["proxy_name"] = *req.ProxyName
		}
	}

	if err := h.service.UpdateAccount(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("Failed to update account", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update account",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Account updated successfully",
		"data":    gin.H{"message": "Account updated successfully"},
	})
}

// DeleteAccount deletes a Codex account.
// @Summary Delete Codex account
// @Tags Codex Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} gin.H
// @Router /admin/codex-accounts/{id} [delete]
func (h *CodexAccountHandler) DeleteAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	if err := h.service.DeleteAccount(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete account", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete account",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Account deleted successfully",
		"data":    gin.H{"message": "Account deleted successfully"},
	})
}

// ToggleStatus toggles account active status.
// @Summary Toggle account status
// @Tags Codex Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} gin.H
// @Router /admin/codex-accounts/{id}/toggle [post]
func (h *CodexAccountHandler) ToggleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	acc, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
		return
	}

	updates := map[string]any{
		"is_active": !acc.IsActive,
	}

	if err := h.service.UpdateAccount(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("Failed to toggle account status", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to toggle account status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Account status toggled successfully",
		"data": gin.H{
			"is_active": !acc.IsActive,
		},
	})
}

// TestAccount tests account connectivity.
// @Summary Test account
// @Tags Codex Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} gin.H
// @Router /admin/codex-accounts/{id}/test [post]
func (h *CodexAccountHandler) TestAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	if err := h.service.TestAccount(c.Request.Context(), id); err != nil {
		h.logger.Error("Account test failed", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Account test failed: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Account test successful",
		"data": gin.H{
			"success": true,
		},
	})
}

// RefreshToken refreshes OAuth token for an account.
// @Summary Refresh OAuth token
// @Tags Codex Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} gin.H
// @Router /admin/codex-accounts/{id}/refresh-token [post]
func (h *CodexAccountHandler) RefreshToken(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	if err := h.service.RefreshToken(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to refresh token", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Token refreshed successfully",
		"data":    gin.H{"message": "Token refreshed successfully"},
	})
}

// ClearRateLimit clears rate limit status for an account.
// @Summary Clear rate limit
// @Tags Codex Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} gin.H
// @Router /admin/codex-accounts/{id}/clear-rate-limit [post]
func (h *CodexAccountHandler) ClearRateLimit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
		})
		return
	}

	// Clear rate limit fields
	updates := map[string]any{
		"rate_limited_until": nil,
		"rate_limit_status":  nil,
		"overload_until":     nil,
	}

	if err := h.service.UpdateAccount(c.Request.Context(), id, updates); err != nil {
		h.logger.Error("Failed to clear rate limit", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear rate limit: " + err.Error(),
		})
		return
	}

	h.logger.Info("Rate limit cleared for account", zap.Int64("id", id))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Rate limit cleared successfully",
		"data":    gin.H{"message": "Rate limit cleared successfully"},
	})
}

// toAccountInfo converts model to API response.
func (h *CodexAccountHandler) toAccountInfo(acc *model.CodexAccount) CodexAccountInfo {
	info := CodexAccountInfo{
		ID:                acc.ID,
		Name:              acc.Name,
		AccountType:       acc.AccountType,
		Email:             acc.Email,
		ChatGPTAccountID:  acc.ChatGPTAccountID,
		ChatGPTUserID:     acc.ChatGPTUserID,
		OrganizationID:    acc.OrganizationID,
		OrganizationRole:  acc.OrganizationRole,
		OrganizationTitle: acc.OrganizationTitle,
		BaseAPI:           acc.BaseAPI,
		DailyQuota:        acc.DailyQuota,
		DailyUsage:        acc.DailyUsage,
		IsActive:          acc.IsActive,
		Schedulable:       acc.Schedulable,
		Priority:          acc.Priority,
		ProxyName:         acc.ProxyName,
		CreatedAt:         acc.CreatedAt.Unix(),
		UpdatedAt:         acc.UpdatedAt.Unix(),
	}

	if acc.SubscriptionLevel != nil {
		info.SubscriptionLevel = acc.SubscriptionLevel
	}

	if acc.SubscriptionExpiresAt != nil {
		expiresAt := acc.SubscriptionExpiresAt.Unix()
		info.SubscriptionExpiresAt = &expiresAt
	}

	if acc.RateLimitedUntil != nil {
		limitedUntil := acc.RateLimitedUntil.Unix()
		info.RateLimitedUntil = &limitedUntil
	}

	if acc.LastUsedAt != nil {
		lastUsed := acc.LastUsedAt.Unix()
		info.LastUsedAt = &lastUsed
	}

	return info
}
