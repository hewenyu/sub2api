package relay

import "errors"

// retryableError marks an upstream error as safe to retry within the same request.
type retryableError struct {
	msg string
}

func (e *retryableError) Error() string {
	return e.msg
}

func newRetryableError(msg string) error {
	return &retryableError{msg: msg}
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var re *retryableError
	return errors.As(err, &re)
}
