package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
)

// MockModelRestrictor is a mock for ModelRestrictor.
type MockModelRestrictor struct {
	mock.Mock
}

func (m *MockModelRestrictor) IsModelAllowed(apiKey *model.APIKey, modelName string) bool { //nolint:errcheck
	args := m.Called(apiKey, modelName)
	return args.Bool(0)
}

func TestModelRestrictionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	t.Run("no API key in context", func(t *testing.T) {
		mockRestrictor := new(MockModelRestrictor)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"model":"claude-3-opus-20240229"}`)
		c.Request = httptest.NewRequest("POST", "/test", body)

		middleware := ModelRestrictionMiddleware(mockRestrictor, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("restriction disabled", func(t *testing.T) {
		mockRestrictor := new(MockModelRestrictor)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"model":"claude-3-opus-20240229"}`)
		c.Request = httptest.NewRequest("POST", "/test", body)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:                     1,
			EnableModelRestriction: false,
		})

		middleware := ModelRestrictionMiddleware(mockRestrictor, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allowed model", func(t *testing.T) {
		mockRestrictor := new(MockModelRestrictor)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"model":"claude-3-haiku-20240307"}`)
		c.Request = httptest.NewRequest("POST", "/test", body)
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
			RestrictedModels:       model.StringArray{"claude-3-opus-20240229"},
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockRestrictor.On("IsModelAllowed", apiKey, "claude-3-haiku-20240307").Return(true)

		middleware := ModelRestrictionMiddleware(mockRestrictor, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRestrictor.AssertExpectations(t)
	})

	t.Run("blocked model", func(t *testing.T) {
		mockRestrictor := new(MockModelRestrictor)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"model":"claude-3-opus-20240229"}`)
		c.Request = httptest.NewRequest("POST", "/test", body)
		apiKey := &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
			RestrictedModels:       model.StringArray{"claude-3-opus-20240229"},
		}
		c.Set(ContextKeyAPIKey, apiKey)

		mockRestrictor.On("IsModelAllowed", apiKey, "claude-3-opus-20240229").Return(false)

		middleware := ModelRestrictionMiddleware(mockRestrictor, logger)
		middleware(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.True(t, c.IsAborted())
		mockRestrictor.AssertExpectations(t)
	})

	t.Run("empty model in request", func(t *testing.T) {
		mockRestrictor := new(MockModelRestrictor)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"prompt":"Hello"}`)
		c.Request = httptest.NewRequest("POST", "/test", body)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
		})

		middleware := ModelRestrictionMiddleware(mockRestrictor, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid json body", func(t *testing.T) {
		mockRestrictor := new(MockModelRestrictor)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{invalid json}`)
		c.Request = httptest.NewRequest("POST", "/test", body)
		c.Set(ContextKeyAPIKey, &model.APIKey{
			ID:                     1,
			EnableModelRestriction: true,
		})

		middleware := ModelRestrictionMiddleware(mockRestrictor, logger)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
