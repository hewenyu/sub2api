package account

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetHTTPClientWithProxyName(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("returns proxy client when proxy name is provided", func(t *testing.T) {
		mockClientManager := new(MockProxyClientManager)
		proxyClient := &http.Client{Timeout: 60 * time.Second}
		proxyName := "test-proxy"

		mockClientManager.On("GetClient", ctx, proxyName).Return(proxyClient, nil)

		service := &codexAccountService{
			proxyClientManager: mockClientManager,
			httpClient:         &http.Client{Timeout: 30 * time.Second},
			logger:             logger,
		}

		client, err := service.getHTTPClientByProxyName(ctx, &proxyName)

		assert.NoError(t, err)
		assert.Equal(t, proxyClient, client)
		mockClientManager.AssertExpectations(t)
	})

	t.Run("returns default client when proxy name is nil", func(t *testing.T) {
		mockClientManager := new(MockProxyClientManager)
		defaultClient := &http.Client{Timeout: 30 * time.Second}

		mockClientManager.On("GetDefaultClient", ctx).Return(defaultClient)

		service := &codexAccountService{
			proxyClientManager: mockClientManager,
			httpClient:         defaultClient,
			logger:             logger,
		}

		client, err := service.getHTTPClientByProxyName(ctx, nil)

		assert.NoError(t, err)
		assert.Equal(t, defaultClient, client)
		mockClientManager.AssertExpectations(t)
	})

	t.Run("returns default client when proxy name is empty string", func(t *testing.T) {
		mockClientManager := new(MockProxyClientManager)
		defaultClient := &http.Client{Timeout: 30 * time.Second}
		emptyProxyName := ""

		mockClientManager.On("GetDefaultClient", ctx).Return(defaultClient)

		service := &codexAccountService{
			proxyClientManager: mockClientManager,
			httpClient:         defaultClient,
			logger:             logger,
		}

		client, err := service.getHTTPClientByProxyName(ctx, &emptyProxyName)

		assert.NoError(t, err)
		assert.Equal(t, defaultClient, client)
		mockClientManager.AssertExpectations(t)
	})

	t.Run("returns default client when proxy client manager is nil", func(t *testing.T) {
		defaultClient := &http.Client{Timeout: 30 * time.Second}
		proxyName := "test-proxy"

		service := &codexAccountService{
			proxyClientManager: nil,
			httpClient:         defaultClient,
			logger:             logger,
		}

		client, err := service.getHTTPClientByProxyName(ctx, &proxyName)

		assert.NoError(t, err)
		assert.Equal(t, defaultClient, client)
	})
}

