package retry

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestRetryManager_Do_Success(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	err := rm.Do(context.Background(), func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

func TestRetryManager_Do_RetryAndSuccess(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	err := rm.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return &RetryableError{StatusCode: 503}
		}
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestRetryManager_Do_MaxAttemptsExceeded(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	err := rm.Do(context.Background(), func() error {
		attempts++
		return &RetryableError{StatusCode: 503}
	})

	if err == nil {
		t.Error("Do() error = nil, want error")
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestRetryManager_Do_NonRetryableError(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	expectedErr := errors.New("non-retryable error")
	err := rm.Do(context.Background(), func() error {
		attempts++
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("Do() error = %v, want %v", err, expectedErr)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

func TestRetryManager_Do_ContextCanceled(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0
	err := rm.Do(ctx, func() error {
		attempts++
		return &RetryableError{StatusCode: 503}
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do() error = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

func TestRetryManager_DoWithFailover_Success(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	accountIDs := []int64{}
	err := rm.DoWithFailover(
		context.Background(),
		func(accountID int64) error {
			attempts++
			accountIDs = append(accountIDs, accountID)
			return nil
		},
		func() (int64, error) {
			return int64(attempts + 1), nil
		},
	)

	if err != nil {
		t.Errorf("DoWithFailover() error = %v, want nil", err)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
	if len(accountIDs) != 1 || accountIDs[0] != 1 {
		t.Errorf("accountIDs = %v, want [1]", accountIDs)
	}
}

func TestRetryManager_DoWithFailover_AccountFailover(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	accountIDs := []int64{}
	err := rm.DoWithFailover(
		context.Background(),
		func(accountID int64) error {
			attempts++
			accountIDs = append(accountIDs, accountID)
			if attempts < 3 {
				return &RetryableError{StatusCode: 503}
			}
			return nil
		},
		func() (int64, error) {
			return int64(attempts + 1), nil
		},
	)

	if err != nil {
		t.Errorf("DoWithFailover() error = %v, want nil", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	// Should use different accounts: 1, 2, 3
	if len(accountIDs) != 3 || accountIDs[0] != 1 || accountIDs[1] != 2 || accountIDs[2] != 3 {
		t.Errorf("accountIDs = %v, want [1 2 3]", accountIDs)
	}
}

func TestRetryManager_DoWithFailover_GetNextAccountError(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0.0,
	}
	rm := NewRetryManager(policy)

	attempts := 0
	expectedErr := errors.New("no accounts available")
	err := rm.DoWithFailover(
		context.Background(),
		func(accountID int64) error {
			attempts++
			return &RetryableError{StatusCode: 503}
		},
		func() (int64, error) {
			if attempts > 0 {
				return 0, expectedErr
			}
			return 1, nil
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Errorf("DoWithFailover() error = %v, want %v", err, expectedErr)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

// RetryableError is a test error type that implements error interface
type RetryableError struct {
	StatusCode int
}

func (e *RetryableError) Error() string {
	return "retryable error"
}

func (e *RetryableError) HTTPResponse() *http.Response {
	return &http.Response{StatusCode: e.StatusCode}
}
