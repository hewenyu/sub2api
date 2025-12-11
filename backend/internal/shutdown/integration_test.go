package shutdown

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockCleanupServiceIntegration struct {
	stopped bool
	mu      sync.Mutex
}

func (m *mockCleanupServiceIntegration) Start(ctx context.Context) {}

func (m *mockCleanupServiceIntegration) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopped = true
}

func (m *mockCleanupServiceIntegration) IsStopped() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopped
}

func TestGracefulShutdownIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	tracker := NewRequestTracker()
	cleanupService := &mockCleanupServiceIntegration{}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if tracker.IsShuttingDown() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Server is shutting down",
			})
			c.Abort()
			return
		}

		tracker.Start()
		defer tracker.End()

		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: router,
	}

	manager := NewManager(
		server,
		cleanupService,
		nil,
		db,
		redisClient,
		tracker,
		2*time.Second,
		logger,
	)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		tracker.Start()
		time.Sleep(150 * time.Millisecond)
		tracker.End()
	}()

	go func() {
		defer wg.Done()
		tracker.Start()
		time.Sleep(150 * time.Millisecond)
		tracker.End()
	}()

	time.Sleep(50 * time.Millisecond)

	tracker.BeginShutdown()

	if err := manager.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	wg.Wait()

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 after shutdown, got %d", tracker.ActiveCount())
	}

	if !cleanupService.IsStopped() {
		t.Error("Expected cleanup service to be stopped")
	}

	if !tracker.IsShuttingDown() {
		t.Error("Expected tracker to be in shutting down state")
	}
}
