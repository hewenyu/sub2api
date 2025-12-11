package account

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

// CreateAccount creates a new Codex account.
func (s *codexAccountService) CreateAccount(ctx context.Context, req *CreateCodexAccountRequest) (*model.CodexAccount, error) {
	// Validate required fields
	if req == nil {
		return nil, errors.New("create account request cannot be nil")
	}
	if req.Name == "" {
		return nil, errors.New("account name is required")
	}
	if req.AccountType == "" {
		return nil, errors.New("account_type is required and must be 'openai-oauth' or 'openai-responses'")
	}

	// Validate account type
	if req.AccountType != "openai-oauth" && req.AccountType != "openai-responses" {
		return nil, fmt.Errorf("invalid account_type '%s', must be 'openai-oauth' or 'openai-responses'", req.AccountType)
	}

	// Validate account type specific requirements
	if req.AccountType == "openai-responses" && req.APIKey == nil {
		return nil, errors.New("API key is required for openai-responses accounts")
	}
	if req.AccountType == "openai-oauth" && req.APIKey != nil {
		s.logger.Warn("API key provided for OAuth account type, it will be ignored",
			zap.String("account_name", req.Name),
		)
	}

	s.logger.Info("Creating Codex account",
		zap.String("name", req.Name),
		zap.String("account_type", req.AccountType),
		zap.Bool("has_api_key", req.APIKey != nil),
	)

	// Apply defaults
	if req.BaseAPI == "" {
		req.BaseAPI = "https://api.openai.com/v1"
	}

	if req.QuotaResetTime == "" {
		req.QuotaResetTime = "00:00"
	}

	if req.Priority == 0 {
		req.Priority = 100
	}

	// Default schedulable to true if not explicitly set to false
	// Since bool defaults to false in Go/JSON, we treat false as "not set"
	// and default to true to make accounts available for scheduling
	schedulable := true
	if req.Schedulable {
		schedulable = req.Schedulable
	}

	account := &model.CodexAccount{
		Name:            req.Name,
		AccountType:     req.AccountType,
		Email:           req.Email,
		BaseAPI:         req.BaseAPI,
		CustomUserAgent: req.CustomUserAgent,
		DailyQuota:      req.DailyQuota,
		QuotaResetTime:  req.QuotaResetTime,
		Priority:        req.Priority,
		Schedulable:     schedulable,
		IsActive:        true,
	}

	// Encrypt API key if provided
	if req.APIKey != nil {
		encryptedKey, err := crypto.AES256Encrypt(*req.APIKey, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		account.APIKey = &encryptedKey
	}

	// Proxy configuration (by name)
	if req.ProxyName != nil && *req.ProxyName != "" {
		account.ProxyName = req.ProxyName
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	s.logger.Info("Codex account created",
		zap.Int64("account_id", account.ID),
		zap.String("name", account.Name),
		zap.String("type", account.AccountType),
	)

	return account, nil
}

// GetAccount retrieves an account by ID.
func (s *codexAccountService) GetAccount(ctx context.Context, id int64) (*model.CodexAccount, error) {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// UpdateAccount updates an account.
func (s *codexAccountService) UpdateAccount(ctx context.Context, id int64, updates map[string]any) error {
	// Encrypt API key if being updated
	if apiKey, ok := updates["api_key"]; ok {
		if apiKeyStr, ok := apiKey.(string); ok && apiKeyStr != "" {
			encryptedKey, err := crypto.AES256Encrypt(apiKeyStr, s.encryptionKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt API key: %w", err)
			}
			updates["api_key"] = encryptedKey
		}
	}

	if err := s.repo.UpdateFields(ctx, id, updates); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	s.logger.Info("Codex account updated",
		zap.Int64("account_id", id),
	)

	return nil
}

// DeleteAccount deletes an account.
func (s *codexAccountService) DeleteAccount(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	s.logger.Info("Codex account deleted",
		zap.Int64("account_id", id),
	)

	return nil
}

// ListAccounts lists accounts with filtering and pagination.
func (s *codexAccountService) ListAccounts(ctx context.Context, filters repository.CodexAccountFilters, page, pageSize int) ([]*model.CodexAccount, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	accounts, total, err := s.repo.List(ctx, filters, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list accounts: %w", err)
	}

	return accounts, total, nil
}
