package retry

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// RetryManager handles retry logic with backoff and failover
type RetryManager interface {
	Do(ctx context.Context, fn func() error) error
	DoWithFailover(ctx context.Context, fn func(accountID int64) error, getNextAccount func() (int64, error)) error
}

type retryManager struct {
	policy  RetryPolicy
	backoff BackoffStrategy
}

// NewRetryManager creates a new retry manager with given policy
func NewRetryManager(policy RetryPolicy) RetryManager {
	return &retryManager{
		policy:  policy,
		backoff: NewExponentialBackoff(policy),
	}
}

// Do executes function with retry logic
func (rm *retryManager) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= rm.policy.MaxAttempts; attempt++ {
		// Execute function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check context after first attempt
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		// Check if error is retryable
		if !rm.isRetryable(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt < rm.policy.MaxAttempts {
			backoffDuration := rm.backoff.Next(attempt)
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

// DoWithFailover executes function with retry logic and account failover
func (rm *retryManager) DoWithFailover(ctx context.Context, fn func(accountID int64) error, getNextAccount func() (int64, error)) error {
	var lastErr error

	for attempt := 1; attempt <= rm.policy.MaxAttempts; attempt++ {
		// Get next account
		accountID, err := getNextAccount()
		if err != nil {
			return err
		}

		// Execute function with account
		err = fn(accountID)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check context after first attempt
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		// Check if error is retryable
		if !rm.isRetryable(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt < rm.policy.MaxAttempts {
			backoffDuration := rm.backoff.Next(attempt)
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

// isRetryable checks if error should be retried
func (rm *retryManager) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check if error provides HTTP response
	type httpResponder interface {
		HTTPResponse() *http.Response
	}

	var responder httpResponder
	if errors.As(err, &responder) {
		return IsRetryable(err, responder.HTTPResponse())
	}

	return IsRetryable(err, nil)
}
