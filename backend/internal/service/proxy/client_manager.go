package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

// ProxyClientManager manages HTTP clients with proxy configurations.
type ProxyClientManager interface {
	// GetClient returns an HTTP client for the specified proxy name.
	GetClient(ctx context.Context, proxyName string) (*http.Client, error)

	// GetClientByID returns an HTTP client for the specified proxy ID.
	GetClientByID(ctx context.Context, proxyID int64) (*http.Client, error)

	// GetDefaultClient returns the default HTTP client without proxy.
	GetDefaultClient(ctx context.Context) *http.Client

	// GetStreamingClient returns a streaming HTTP client for the specified proxy name.
	GetStreamingClient(ctx context.Context, proxyName string) (*http.Client, error)

	// InvalidateCache invalidates the cache for a specific proxy.
	InvalidateCache(proxyName string)

	// InvalidateAllCache clears all cached clients.
	InvalidateAllCache()
}

type proxyClientManager struct {
	proxyRepo              repository.ProxyConfigRepository
	encryptionKey          string
	logger                 *zap.Logger
	clientCache            map[string]*http.Client
	streamingClientCache   map[string]*http.Client
	cacheMutex             sync.RWMutex
	defaultClient          *http.Client
	defaultStreamingClient *http.Client
}

// NewProxyClientManager creates a new proxy client manager.
func NewProxyClientManager(
	proxyRepo repository.ProxyConfigRepository,
	encryptionKey string,
	logger *zap.Logger,
) ProxyClientManager {
	defaultConfig := DefaultHTTPClientConfig()
	streamingConfig := StreamingHTTPClientConfig()

	defaultClient, _ := NewHTTPClient(nil, defaultConfig, encryptionKey)
	defaultStreamingClient, _ := NewHTTPClient(nil, streamingConfig, encryptionKey)

	return &proxyClientManager{
		proxyRepo:              proxyRepo,
		encryptionKey:          encryptionKey,
		logger:                 logger,
		clientCache:            make(map[string]*http.Client),
		streamingClientCache:   make(map[string]*http.Client),
		defaultClient:          defaultClient,
		defaultStreamingClient: defaultStreamingClient,
	}
}

func (m *proxyClientManager) GetClient(ctx context.Context, proxyName string) (*http.Client, error) {
	if proxyName == "" {
		return m.defaultClient, nil
	}

	m.cacheMutex.RLock()
	if client, ok := m.clientCache[proxyName]; ok {
		m.cacheMutex.RUnlock()
		return client, nil
	}
	m.cacheMutex.RUnlock()

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	if client, ok := m.clientCache[proxyName]; ok {
		return client, nil
	}

	proxyConfig, err := m.proxyRepo.GetByName(ctx, proxyName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy not found: %s", proxyName)
		}
		return nil, fmt.Errorf("failed to get proxy config: %w", err)
	}

	if !proxyConfig.Enabled {
		m.logger.Warn("Proxy is disabled, returning default client", zap.String("proxy", proxyName))
		return m.defaultClient, nil
	}

	client, err := NewHTTPClient(proxyConfig, DefaultHTTPClientConfig(), m.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	m.clientCache[proxyName] = client
	m.logger.Info("Created and cached HTTP client", zap.String("proxy", proxyName))

	return client, nil
}

func (m *proxyClientManager) GetClientByID(ctx context.Context, proxyID int64) (*http.Client, error) {
	proxyConfig, err := m.proxyRepo.GetByID(ctx, proxyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy not found: %d", proxyID)
		}
		return nil, fmt.Errorf("failed to get proxy config: %w", err)
	}

	return m.GetClient(ctx, proxyConfig.Name)
}

func (m *proxyClientManager) GetDefaultClient(ctx context.Context) *http.Client {
	return m.defaultClient
}

func (m *proxyClientManager) GetStreamingClient(ctx context.Context, proxyName string) (*http.Client, error) {
	if proxyName == "" {
		return m.defaultStreamingClient, nil
	}

	m.cacheMutex.RLock()
	if client, ok := m.streamingClientCache[proxyName]; ok {
		m.cacheMutex.RUnlock()
		return client, nil
	}
	m.cacheMutex.RUnlock()

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	if client, ok := m.streamingClientCache[proxyName]; ok {
		return client, nil
	}

	proxyConfig, err := m.proxyRepo.GetByName(ctx, proxyName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proxy not found: %s", proxyName)
		}
		return nil, fmt.Errorf("failed to get proxy config: %w", err)
	}

	if !proxyConfig.Enabled {
		m.logger.Warn("Proxy is disabled, returning default streaming client", zap.String("proxy", proxyName))
		return m.defaultStreamingClient, nil
	}

	client, err := NewHTTPClient(proxyConfig, StreamingHTTPClientConfig(), m.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming HTTP client: %w", err)
	}

	m.streamingClientCache[proxyName] = client
	m.logger.Info("Created and cached streaming HTTP client", zap.String("proxy", proxyName))

	return client, nil
}

func (m *proxyClientManager) InvalidateCache(proxyName string) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	delete(m.clientCache, proxyName)
	delete(m.streamingClientCache, proxyName)
	m.logger.Info("Invalidated cache for proxy", zap.String("proxy", proxyName))
}

func (m *proxyClientManager) InvalidateAllCache() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.clientCache = make(map[string]*http.Client)
	m.streamingClientCache = make(map[string]*http.Client)
	m.logger.Info("Invalidated all proxy client caches")
}
