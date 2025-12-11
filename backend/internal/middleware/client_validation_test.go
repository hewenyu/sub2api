package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// MockClientValidator is a mock for ClientValidator.
type MockClientValidator struct {
	mock.Mock
}

func (m *MockClientValidator) ValidateClaudeCode(headers http.Header, body []byte) (bool, error) { //nolint:errcheck
	args := m.Called(headers, body)
	return args.Bool(0), args.Error(1)
}

func (m *MockClientValidator) ValidateCodex(headers http.Header, body []byte) (bool, error) { //nolint:errcheck
	args := m.Called(headers, body)
	return args.Bool(0), args.Error(1)
}

func (m *MockClientValidator) IsClientAllowed(apiKey *model.APIKey, userAgent string) bool { //nolint:errcheck
	args := m.Called(apiKey, userAgent)
	return args.Bool(0)
}

func TestClientValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("no API key in context", func(t *testing.T) {
		mockValidator := new(MockClientValidator)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		middleware := ClientValidationMiddleware(mockValidator, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("restriction disabled", func(t *testing.T) {
		mockValidator := new(MockClientValidator)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:                      1,
			EnableClientRestriction: false,
		})

		middleware := ClientValidationMiddleware(mockValidator, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allowed client", func(t *testing.T) {
		mockValidator := new(MockClientValidator)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("User-Agent", "cursor/1.0.0")
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor", "vscode"},
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockValidator.On("IsClientAllowed", apiKey, "cursor/1.0.0").Return(true)

		middleware := ClientValidationMiddleware(mockValidator, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockValidator.AssertExpectations(t)
	})

	t.Run("blocked client", func(t *testing.T) {
		mockValidator := new(MockClientValidator)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("User-Agent", "curl/7.68.0")
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor", "vscode"},
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockValidator.On("IsClientAllowed", apiKey, "curl/7.68.0").Return(false)

		middleware := ClientValidationMiddleware(mockValidator, logger)
		middleware(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.True(t, c.IsAborted())
		mockValidator.AssertExpectations(t)
	})

	t.Run("empty user agent", func(t *testing.T) {
		mockValidator := new(MockClientValidator)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor"},
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockValidator.On("IsClientAllowed", apiKey, "").Return(false)

		middleware := ClientValidationMiddleware(mockValidator, logger)
		middleware(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.True(t, c.IsAborted())
		mockValidator.AssertExpectations(t)
	})

	t.Run("multiple allowed clients", func(t *testing.T) {
		mockValidator := new(MockClientValidator)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("User-Agent", "vscode/1.85.0")
		apiKey := &model.APIKey{
			ID:                      1,
			EnableClientRestriction: true,
			AllowedClients:          model.StringArray{"cursor", "vscode", "codex"},
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockValidator.On("IsClientAllowed", apiKey, "vscode/1.85.0").Return(true)

		middleware := ClientValidationMiddleware(mockValidator, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockValidator.AssertExpectations(t)
	})
}
