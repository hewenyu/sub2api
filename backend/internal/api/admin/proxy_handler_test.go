package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/proxy"
)

type MockProxyService struct {
	mock.Mock
}

func (m *MockProxyService) CreateProxy(ctx context.Context, req *proxy.CreateProxyRequest) (*model.ProxyConfig, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProxyConfig), args.Error(1)
}

func (m *MockProxyService) GetProxy(ctx context.Context, id int64) (*model.ProxyConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProxyConfig), args.Error(1)
}

func (m *MockProxyService) GetProxyByName(ctx context.Context, name string) (*model.ProxyConfig, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProxyConfig), args.Error(1)
}

func (m *MockProxyService) GetDefaultProxy(ctx context.Context) (*model.ProxyConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProxyConfig), args.Error(1)
}

func (m *MockProxyService) ListProxies(ctx context.Context, filters repository.ProxyConfigFilters, page, pageSize int) ([]*model.ProxyConfig, int64, error) {
	args := m.Called(ctx, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.ProxyConfig), args.Get(1).(int64), args.Error(2)
}

func (m *MockProxyService) UpdateProxy(ctx context.Context, id int64, req *proxy.UpdateProxyRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockProxyService) DeleteProxy(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProxyService) SetDefaultProxy(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProxyService) TestProxy(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProxyService) TestProxyWithGeolocation(ctx context.Context, id int64) (*proxy.ProxyTestResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proxy.ProxyTestResult), args.Error(1)
}

func (m *MockProxyService) DecryptProxyPassword(ctx context.Context, p *model.ProxyConfig) (string, error) {
	args := m.Called(ctx, p)
	return args.String(0), args.Error(1)
}

func TestProxyHandler_CreateProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockService := new(MockProxyService)
	handler := NewProxyHandler(mockService, logger)

	router := gin.New()
	router.POST("/proxies", handler.CreateProxy)

	t.Run("success", func(t *testing.T) {
		reqBody := proxy.CreateProxyRequest{
			Name:     "test-proxy",
			Enabled:  true,
			Protocol: "http",
			Host:     "proxy.example.com",
			Port:     8080,
		}

		expectedProxy := &model.ProxyConfig{
			ID:       1,
			Name:     "test-proxy",
			Enabled:  true,
			Protocol: "http",
			Host:     "proxy.example.com",
			Port:     8080,
		}

		mockService.On("CreateProxy", mock.Anything, mock.MatchedBy(func(req *proxy.CreateProxyRequest) bool {
			return req.Name == "test-proxy"
		})).Return(expectedProxy, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/proxies", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestProxyHandler_ListProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	mockService := new(MockProxyService)
	handler := NewProxyHandler(mockService, logger)

	router := gin.New()
	router.GET("/proxies", handler.ListProxies)

	t.Run("success", func(t *testing.T) {
		proxies := []*model.ProxyConfig{
			{
				ID:       1,
				Name:     "proxy1",
				Enabled:  true,
				Protocol: "http",
				Host:     "proxy1.example.com",
				Port:     8080,
			},
		}

		mockService.On("ListProxies", mock.Anything, mock.Anything, 1, 20).Return(proxies, int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/proxies", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
