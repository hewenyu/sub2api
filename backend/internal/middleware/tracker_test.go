package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Wei-Shaw/sub2api/backend/internal/shutdown"
)

func TestRequestTrackerMiddleware_Normal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracker := shutdown.NewRequestTracker()

	router := gin.New()
	router.Use(RequestTrackerMiddleware(tracker))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 after request, got %d", tracker.ActiveCount())
	}
}

func TestRequestTrackerMiddleware_ShuttingDown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracker := shutdown.NewRequestTracker()
	tracker.BeginShutdown()

	router := gin.New()
	router.Use(RequestTrackerMiddleware(tracker))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 when shutting down, got %d", tracker.ActiveCount())
	}
}

func TestRequestTrackerMiddleware_Tracking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracker := shutdown.NewRequestTracker()

	router := gin.New()
	router.Use(RequestTrackerMiddleware(tracker))
	router.GET("/test", func(c *gin.Context) {
		if tracker.ActiveCount() != 1 {
			t.Errorf("Expected active count to be 1 during request, got %d", tracker.ActiveCount())
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 after request, got %d", tracker.ActiveCount())
	}
}
