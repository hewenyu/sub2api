package retry

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		// Retryable status codes
		{"500 internal server error", 500, nil, true},
		{"502 bad gateway", 502, nil, true},
		{"503 service unavailable", 503, nil, true},
		{"504 gateway timeout", 504, nil, true},
		{"429 rate limit", 429, nil, true},

		// Non-retryable status codes
		{"200 ok", 200, nil, false},
		{"400 bad request", 400, nil, false},
		{"401 unauthorized", 401, nil, false},
		{"403 forbidden", 403, nil, false},
		{"404 not found", 404, nil, false},
		{"422 unprocessable entity", 422, nil, false},

		// Network errors (retryable)
		{"timeout error", 0, &net.OpError{Op: "dial", Err: &timeoutError{}}, true},
		{"connection refused", 0, &net.OpError{Op: "dial", Err: errors.New("connection refused")}, true},
		{"context deadline exceeded", 0, context.DeadlineExceeded, true},

		// Non-retryable errors
		{"context canceled", 0, context.Canceled, false},
		{"generic error", 0, errors.New("generic error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.statusCode > 0 {
				resp = &http.Response{StatusCode: tt.statusCode}
			}

			got := IsRetryable(tt.err, resp)
			if got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"timeout error", &net.OpError{Op: "dial", Err: &timeoutError{}}, true},
		{"connection refused", &net.OpError{Op: "dial", Err: errors.New("connection refused")}, true},
		{"dns error", &net.DNSError{}, true},
		{"context deadline", context.DeadlineExceeded, true},
		{"generic error", errors.New("generic"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNetworkError(tt.err)
			if got != tt.want {
				t.Errorf("IsNetworkError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"500", 500, true},
		{"502", 502, true},
		{"503", 503, true},
		{"504", 504, true},
		{"599", 599, true},
		{"400", 400, false},
		{"404", 404, false},
		{"200", 200, false},
		{"0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			got := IsServerError(resp)
			if got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// timeoutError implements net.Error for testing
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
