# Retry Package

Smart retry logic with exponential backoff, jitter, and error classification for improved system resilience.

## Features

- **Exponential Backoff with Jitter**: Prevents thundering herd problem
- **Intelligent Error Classification**: Distinguishes retryable from non-retryable errors
- **Account Failover**: Automatically selects different accounts on retry
- **Context Awareness**: Respects cancellation and deadlines
- **Minimal Overhead**: < 1ms for successful first attempts

## Configuration

Add to `config.yaml`:

```yaml
retry:
  max_attempts: 3
  initial_backoff: "100ms"
  max_backoff: "5s"
  multiplier: 2.0
  jitter: 0.1
```

## Usage

### Basic Retry

```go
policy := retry.RetryPolicy{
    MaxAttempts:    3,
    InitialBackoff: 100 * time.Millisecond,
    MaxBackoff:     5 * time.Second,
    Multiplier:     2.0,
    Jitter:         0.1,
}
rm := retry.NewRetryManager(policy)

err := rm.Do(ctx, func() error {
    return doSomething()
})
```

### Retry with Account Failover

```go
err := rm.DoWithFailover(
    ctx,
    func(accountID int64) error {
        return processWithAccount(accountID)
    },
    func() (int64, error) {
        return scheduler.SelectNextAccount()
    },
)
```

## Error Classification

### Retryable Errors
- Network timeouts and connection errors
- HTTP 500, 502, 503, 504
- HTTP 429 (rate limit)
- Context deadline exceeded

### Non-Retryable Errors
- HTTP 400, 401, 403, 404
- Invalid requests
- Authentication failures
- Context canceled

## Architecture

```
RetryManager
    |-- BackoffStrategy (exponential with jitter)
    |-- Error Matcher (retryable classification)
    |-- Context Handler (cancellation support)
```

## Testing

Run tests:
```bash
go test -v ./internal/service/retry/...
```

Coverage: 87.3%
