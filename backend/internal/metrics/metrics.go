package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request metrics
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_requests_total",
			Help: "Total number of requests",
		},
		[]string{"api_key_id", "account_id", "model", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "relay_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to 10s
		},
		[]string{"api_key_id", "account_id", "model"},
	)

	// Token metrics
	TokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_tokens_total",
			Help: "Total number of tokens processed",
		},
		[]string{"api_key_id", "account_id", "model", "type"},
	)

	// Cost metrics
	CostTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_cost_total",
			Help: "Total cost in USD",
		},
		[]string{"api_key_id", "account_id", "model"},
	)

	// Scheduling metrics
	SchedulerSelectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_selections_total",
			Help: "Total number of account selections",
		},
		[]string{"strategy", "reason"},
	)

	SchedulerSelectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scheduler_selection_duration_seconds",
			Help:    "Account selection duration",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 8), // 1ms to 128ms
		},
		[]string{"strategy"},
	)

	// Account health metrics
	AccountHealthScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "account_health_score",
			Help: "Account health score (0-1)",
		},
		[]string{"account_id"},
	)

	AccountQuarantineTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "account_quarantine_total",
			Help: "Total number of account quarantines",
		},
		[]string{"account_id", "reason"},
	)

	// Concurrency metrics
	ConcurrencyCurrent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "concurrency_current",
			Help: "Current concurrency count",
		},
		[]string{"type", "id"},
	)

	ConcurrencyLimit = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "concurrency_limit",
			Help: "Concurrency limit",
		},
		[]string{"type", "id"},
	)

	SemaphoreAcquireTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "semaphore_acquire_total",
			Help: "Total semaphore acquire attempts",
		},
		[]string{"type", "status"},
	)

	SemaphoreAcquireDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "semaphore_acquire_duration_seconds",
			Help:    "Semaphore acquire duration",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 8),
		},
		[]string{"type"},
	)

	// Rate limit metrics
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total rate limit hits",
		},
		[]string{"api_key_id", "window"},
	)

	RateLimitCurrent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rate_limit_current",
			Help: "Current rate limit count",
		},
		[]string{"api_key_id", "window"},
	)

	// Circuit breaker metrics
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"name"},
	)

	CircuitBreakerTransitionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_transitions_total",
			Help: "Total circuit breaker state transitions",
		},
		[]string{"name", "from", "to"},
	)

	CircuitBreakerRejectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_rejected_total",
			Help: "Total requests rejected by circuit breaker",
		},
		[]string{"name"},
	)

	// Proxy metrics
	ProxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total proxy requests",
		},
		[]string{"proxy_id", "status"},
	)

	ProxyLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_latency_seconds",
			Help:    "Proxy request latency",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"proxy_id"},
	)

	ProxyFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_failures_total",
			Help: "Total proxy failures",
		},
		[]string{"proxy_id", "reason"},
	)
)
