package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/pkg/crypto"
)

// ProxyConfigService defines the interface for proxy configuration service operations.
type ProxyConfigService interface {
	// CreateProxy creates a new proxy configuration.
	CreateProxy(ctx context.Context, req *CreateProxyRequest) (*model.ProxyConfig, error)

	// GetProxy retrieves a proxy configuration by ID.
	GetProxy(ctx context.Context, id int64) (*model.ProxyConfig, error)

	// GetProxyByName retrieves a proxy configuration by name.
	GetProxyByName(ctx context.Context, name string) (*model.ProxyConfig, error)

	// GetDefaultProxy retrieves the default proxy configuration.
	GetDefaultProxy(ctx context.Context) (*model.ProxyConfig, error)

	// ListProxies retrieves proxy configurations with filters and pagination.
	ListProxies(ctx context.Context, filters repository.ProxyConfigFilters, page, pageSize int) ([]*model.ProxyConfig, int64, error)

	// UpdateProxy updates an existing proxy configuration.
	UpdateProxy(ctx context.Context, id int64, req *UpdateProxyRequest) error

	// DeleteProxy soft-deletes a proxy configuration.
	DeleteProxy(ctx context.Context, id int64) error

	// SetDefaultProxy sets a proxy as the default.
	SetDefaultProxy(ctx context.Context, id int64) error

	// TestProxy tests proxy connectivity (legacy, simple test).
	TestProxy(ctx context.Context, id int64) error

	// TestProxyWithGeolocation tests proxy and returns geolocation information.
	TestProxyWithGeolocation(ctx context.Context, id int64) (*ProxyTestResult, error)

	// DecryptProxyPassword decrypts a proxy password for use.
	DecryptProxyPassword(ctx context.Context, proxy *model.ProxyConfig) (string, error)
}

type proxyConfigService struct {
	repo          repository.ProxyConfigRepository
	encryptionKey string
	logger        *zap.Logger
	clientManager ProxyClientManager
}

// NewProxyConfigService creates a new proxy configuration service.
func NewProxyConfigService(
	repo repository.ProxyConfigRepository,
	encryptionKey string,
	logger *zap.Logger,
	clientManager ProxyClientManager,
) ProxyConfigService {
	return &proxyConfigService{
		repo:          repo,
		encryptionKey: encryptionKey,
		logger:        logger,
		clientManager: clientManager,
	}
}

func (s *proxyConfigService) CreateProxy(ctx context.Context, req *CreateProxyRequest) (*model.ProxyConfig, error) {
	// Validate request
	if err := ValidateCreateRequest(req); err != nil {
		s.logger.Warn("Invalid create proxy request", zap.Error(err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check name uniqueness
	count, err := s.repo.CountByName(ctx, req.Name, 0)
	if err != nil {
		s.logger.Error("Failed to check proxy name uniqueness", zap.Error(err))
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("proxy name already exists: %s", req.Name)
	}

	// Encrypt password if provided
	var encryptedPassword *string
	if req.Password != nil && *req.Password != "" {
		encrypted, err := crypto.AES256Encrypt(*req.Password, s.encryptionKey)
		if err != nil {
			s.logger.Error("Failed to encrypt proxy password", zap.Error(err))
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		encryptedPassword = &encrypted
	}

	// Create proxy config
	proxyConfig := &model.ProxyConfig{
		Name:     req.Name,
		Enabled:  req.Enabled,
		Protocol: req.Protocol,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: encryptedPassword,
	}

	if err := s.repo.Create(ctx, proxyConfig); err != nil {
		s.logger.Error("Failed to create proxy config", zap.Error(err))
		return nil, fmt.Errorf("failed to create proxy: %w", err)
	}

	s.logger.Info("Proxy config created",
		zap.Int64("id", proxyConfig.ID),
		zap.String("name", proxyConfig.Name),
		zap.String("protocol", proxyConfig.Protocol),
		zap.String("host", proxyConfig.Host),
		zap.Int("port", proxyConfig.Port))

	return proxyConfig, nil
}

func (s *proxyConfigService) GetProxy(ctx context.Context, id int64) (*model.ProxyConfig, error) {
	proxyConfig, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy not found: %w", err)
		}
		s.logger.Error("Failed to get proxy config", zap.Int64("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get proxy: %w", err)
	}
	return proxyConfig, nil
}

func (s *proxyConfigService) GetProxyByName(ctx context.Context, name string) (*model.ProxyConfig, error) {
	proxyConfig, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy not found: %w", err)
		}
		s.logger.Error("Failed to get proxy config by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to get proxy: %w", err)
	}
	return proxyConfig, nil
}

func (s *proxyConfigService) GetDefaultProxy(ctx context.Context) (*model.ProxyConfig, error) {
	proxyConfig, err := s.repo.GetDefault(ctx)
	if err != nil {
		s.logger.Error("Failed to get default proxy config", zap.Error(err))
		return nil, fmt.Errorf("failed to get default proxy: %w", err)
	}

	// No default proxy configured.
	if proxyConfig == nil {
		return nil, nil
	}

	return proxyConfig, nil
}

func (s *proxyConfigService) ListProxies(ctx context.Context, filters repository.ProxyConfigFilters, page, pageSize int) ([]*model.ProxyConfig, int64, error) {
	proxies, total, err := s.repo.List(ctx, filters, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list proxy configs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list proxies: %w", err)
	}
	return proxies, total, nil
}

func (s *proxyConfigService) UpdateProxy(ctx context.Context, id int64, req *UpdateProxyRequest) error {
	// Validate request
	if err := ValidateUpdateRequest(req); err != nil {
		s.logger.Warn("Invalid update proxy request", zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get existing proxy
	proxyConfig, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("proxy not found: %w", err)
		}
		s.logger.Error("Failed to get proxy config for update", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("failed to get proxy: %w", err)
	}

	// Check name uniqueness if name is being updated
	if req.Name != nil && *req.Name != proxyConfig.Name {
		count, err := s.repo.CountByName(ctx, *req.Name, id)
		if err != nil {
			s.logger.Error("Failed to check proxy name uniqueness", zap.Error(err))
			return fmt.Errorf("failed to check name uniqueness: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("proxy name already exists: %s", *req.Name)
		}
	}

	// Build updates map from provided fields
	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Protocol != nil {
		updates["protocol"] = *req.Protocol
	}
	if req.Host != nil {
		updates["host"] = *req.Host
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.Username != nil {
		if *req.Username == "" {
			updates["username"] = nil
		} else {
			updates["username"] = *req.Username
		}
	}

	// Handle password update
	// nil = no change, empty string = remove password, value = update password
	if req.Password != nil {
		if *req.Password == "" {
			// Remove password
			updates["password"] = nil
		} else {
			// Encrypt new password
			encrypted, err := crypto.AES256Encrypt(*req.Password, s.encryptionKey)
			if err != nil {
				s.logger.Error("Failed to encrypt proxy password", zap.Error(err))
				return fmt.Errorf("failed to encrypt password: %w", err)
			}
			updates["password"] = encrypted
		}
	}

	// Update in database
	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("Failed to update proxy config", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("failed to update proxy: %w", err)
	}

	s.logger.Info("Proxy config updated",
		zap.Int64("id", id),
		zap.Int("fields_updated", len(updates)))

	// Invalidate cached HTTP clients for this proxy so that subsequent
	// requests see the updated configuration immediately.
	if s.clientManager != nil {
		// Invalidate using the old name; if the name was changed, also
		// proactively invalidate the new name (defensive, though usually
		// no clients would exist under the new name yet).
		oldName := proxyConfig.Name
		s.clientManager.InvalidateCache(oldName)

		if req.Name != nil && *req.Name != "" && *req.Name != oldName {
			s.clientManager.InvalidateCache(*req.Name)
		}
	}

	return nil
}

func (s *proxyConfigService) DeleteProxy(ctx context.Context, id int64) error {
	// Fetch existing config first so we can invalidate the proper cache key.
	proxyConfig, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get proxy config for delete", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("failed to get proxy: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete proxy config", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete proxy: %w", err)
	}

	s.logger.Info("Proxy config deleted", zap.Int64("id", id))

	// Invalidate any cached HTTP clients that were using this proxy.
	if s.clientManager != nil {
		s.clientManager.InvalidateCache(proxyConfig.Name)
	}

	return nil
}

func (s *proxyConfigService) SetDefaultProxy(ctx context.Context, id int64) error {
	// Check if proxy exists and is enabled
	proxyConfig, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("proxy not found: %w", err)
		}
		s.logger.Error("Failed to get proxy config", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("failed to get proxy: %w", err)
	}

	if !proxyConfig.Enabled {
		return fmt.Errorf("cannot set disabled proxy as default")
	}

	// Set as default (transaction handled by repository)
	if err := s.repo.SetDefault(ctx, id); err != nil {
		s.logger.Error("Failed to set default proxy", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("failed to set default proxy: %w", err)
	}

	s.logger.Info("Default proxy set", zap.Int64("id", id), zap.String("name", proxyConfig.Name))
	return nil
}

func (s *proxyConfigService) TestProxy(ctx context.Context, id int64) error {
	// Legacy method – delegate to TestProxyWithGeolocation and return only error.
	_, err := s.TestProxyWithGeolocation(ctx, id)
	return err
}

func (s *proxyConfigService) TestProxyWithGeolocation(ctx context.Context, id int64) (*ProxyTestResult, error) {
	// Get proxy config
	proxyConfig, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy: %w", err)
	}

	// Build proxy URL
	proxyURL, err := s.buildProxyURL(ctx, proxyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build proxy URL: %w", err)
	}

	// Test proxy connection with geolocation
	startTime := time.Now()
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 15 * time.Second,
	}

	// Try multiple geolocation services as fallback
	geoServices := []struct {
		name string
		url  string
	}{
		{"ipapi.co", "https://ipapi.co/json/"},
		{"ip-api.com", "http://ip-api.com/json/"},
	}

	var lastErr error
	for _, service := range geoServices {
		resp, err := client.Get(service.url)
		elapsed := time.Since(startTime).Milliseconds()

		if err != nil {
			lastErr = err
			s.logger.Warn("Failed to fetch geolocation from service",
				zap.String("service", service.name),
				zap.Error(err))
			continue
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				s.logger.Warn("Failed to close response body", zap.Error(closeErr))
			}
		}()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Parse response based on service
		if service.name == "ipapi.co" {
			var geoResp struct {
				IP      string `json:"ip"`
				Country string `json:"country_name"`
				Region  string `json:"region"`
				City    string `json:"city"`
				Org     string `json:"org"`
			}
			if err := json.Unmarshal(body, &geoResp); err != nil {
				lastErr = err
				continue
			}

			provider := service.name
			return &ProxyTestResult{
				Success:     true,
				Message:     "Proxy test with geolocation successful",
				IP:          &geoResp.IP,
				Country:     &geoResp.Country,
				Region:      &geoResp.Region,
				City:        &geoResp.City,
				ISP:         &geoResp.Org,
				ResponseMS:  &elapsed,
				GeoProvider: &provider,
			}, nil
		} else if service.name == "ip-api.com" {
			var geoResp struct {
				Status     string `json:"status"`
				Query      string `json:"query"`
				Country    string `json:"country"`
				RegionName string `json:"regionName"`
				City       string `json:"city"`
				ISP        string `json:"isp"`
			}
			if err := json.Unmarshal(body, &geoResp); err != nil {
				lastErr = err
				continue
			}

			if geoResp.Status != "success" {
				lastErr = fmt.Errorf("geolocation query failed")
				continue
			}

			provider := service.name
			return &ProxyTestResult{
				Success:     true,
				Message:     "Proxy test with geolocation successful",
				IP:          &geoResp.Query,
				Country:     &geoResp.Country,
				Region:      &geoResp.RegionName,
				City:        &geoResp.City,
				ISP:         &geoResp.ISP,
				ResponseMS:  &elapsed,
				GeoProvider: &provider,
			}, nil
		}
	}

	// All services failed
	elapsed := time.Since(startTime).Milliseconds()
	errMsg := lastErr.Error()
	return &ProxyTestResult{
		Success:    false,
		Message:    "Failed to fetch geolocation from all services",
		Error:      &errMsg,
		ResponseMS: &elapsed,
	}, nil
}

func (s *proxyConfigService) DecryptProxyPassword(ctx context.Context, proxy *model.ProxyConfig) (string, error) {
	if proxy.Password == nil || *proxy.Password == "" {
		return "", nil
	}

	decrypted, err := crypto.AES256Decrypt(*proxy.Password, s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}
	return decrypted, nil
}

// buildProxyURL constructs a proxy URL from a ProxyConfig.
// Format: protocol://[username:password@]host:port
// Username and password are URL-encoded.
func (s *proxyConfigService) buildProxyURL(ctx context.Context, proxyConfig *model.ProxyConfig) (*url.URL, error) {
	// Decrypt password if present
	var password string
	if proxyConfig.Password != nil && *proxyConfig.Password != "" {
		decrypted, err := s.DecryptProxyPassword(ctx, proxyConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		password = decrypted
	}

	// Build URL
	var userInfo *url.Userinfo
	if proxyConfig.Username != nil && *proxyConfig.Username != "" {
		if password != "" {
			userInfo = url.UserPassword(*proxyConfig.Username, password)
		} else {
			userInfo = url.User(*proxyConfig.Username)
		}
	}

	proxyURL := &url.URL{
		Scheme: proxyConfig.Protocol,
		Host:   fmt.Sprintf("%s:%d", proxyConfig.Host, proxyConfig.Port),
		User:   userInfo,
	}

	return proxyURL, nil
}
