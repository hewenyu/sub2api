package circuitbreaker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCircuitBreaker_Closed(t *testing.T) {
	cb := New()
	client := &http.Client{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := cb.Do(client, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := New()
	cb.failureThreshold = 3
	client := &http.Client{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)

	for range 3 {
		cb.Do(client, req)
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.GetState())
	}

	_, err := cb.Do(client, req)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := New()
	cb.failureThreshold = 2
	cb.timeout = 100 * time.Millisecond
	client := &http.Client{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)

	cb.Do(client, req)
	cb.Do(client, req)

	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.GetState())
	}

	time.Sleep(150 * time.Millisecond)

	_, err := cb.Do(client, req)
	if err != nil && errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected circuit to be half-open, got error: %v", err)
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open after failure in half-open, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := New()
	cb.failureThreshold = 2
	cb.successThreshold = 2
	cb.timeout = 100 * time.Millisecond
	client := &http.Client{}

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer successServer.Close()

	failReq, _ := http.NewRequest("GET", failServer.URL, nil)
	cb.Do(client, failReq)
	cb.Do(client, failReq)

	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.GetState())
	}

	time.Sleep(150 * time.Millisecond)

	successReq, _ := http.NewRequest("GET", successServer.URL, nil)
	cb.Do(client, successReq)
	cb.Do(client, successReq)

	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after recovery, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := New()
	cb.failureThreshold = 1
	client := &http.Client{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	cb.Do(client, req)

	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.GetState())
	}

	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after reset, got %v", cb.GetState())
	}
}
