package shutdown

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockCleanupService struct {
	stopped bool
}

func (m *mockCleanupService) Start(ctx context.Context) {}

func (m *mockCleanupService) Stop() {
	m.stopped = true
}

func TestManager_Shutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	tracker := NewRequestTracker()
	cleanupService := &mockCleanupService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	httpServer := &http.Server{
		Addr:    server.Listener.Addr().String(),
		Handler: handler,
	}

	manager := NewManager(
		httpServer,
		cleanupService,
		nil,
		db,
		redisClient,
		tracker,
		5*time.Second,
		logger,
	)

	ctx := context.Background()
	if err := manager.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	if !cleanupService.stopped {
		t.Error("Expected cleanup service to be stopped")
	}
}

func TestManager_ShutdownWithInFlightRequests(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	tracker := NewRequestTracker()
	cleanupService := &mockCleanupService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	httpServer := &http.Server{
		Addr:    server.Listener.Addr().String(),
		Handler: handler,
	}

	manager := NewManager(
		httpServer,
		cleanupService,
		nil,
		db,
		redisClient,
		tracker,
		1*time.Second,
		logger,
	)

	tracker.Start()
	tracker.Start()

	go func() {
		time.Sleep(100 * time.Millisecond)
		tracker.End()
		tracker.End()
	}()

	ctx := context.Background()
	if err := manager.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("Expected active count to be 0 after shutdown, got %d", tracker.ActiveCount())
	}
}

func TestManager_ShutdownTimeout(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	tracker := NewRequestTracker()
	cleanupService := &mockCleanupService{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	httpServer := &http.Server{
		Addr:    server.Listener.Addr().String(),
		Handler: handler,
	}

	manager := NewManager(
		httpServer,
		cleanupService,
		nil,
		db,
		redisClient,
		tracker,
		100*time.Millisecond,
		logger,
	)

	tracker.Start()

	ctx := context.Background()
	if err := manager.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	if tracker.ActiveCount() != 1 {
		t.Errorf("Expected active count to be 1 after timeout, got %d", tracker.ActiveCount())
	}

	tracker.End()
}
