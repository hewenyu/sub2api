package retry

import (
	"context"
	"errors"
	"net"
	"net/http"
)

// IsRetryable determines if an error or HTTP response should be retried
func IsRetryable(err error, resp *http.Response) bool {
	// Context canceled is not retryable
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Network errors are retryable
	if IsNetworkError(err) {
		return true
	}

	// Check HTTP status codes
	if resp != nil {
		return IsServerError(resp) || resp.StatusCode == http.StatusTooManyRequests
	}

	return false
}

// IsNetworkError checks if error is a network-related error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Context deadline exceeded is retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for net.Error with timeout
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Check for network operation errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

// IsServerError checks if HTTP response is a server error (5xx)
func IsServerError(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	return resp.StatusCode >= 500 && resp.StatusCode < 600
}
