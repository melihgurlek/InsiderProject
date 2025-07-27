package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal tracks total number of HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status_code"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	// HTTPRequestsInFlight tracks current number of HTTP requests being processed
	HTTPRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"method", "route"},
	)

	// DatabaseOperations tracks database operation metrics
	DatabaseOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)

	// DatabaseOperationDuration tracks database operation duration
	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// CacheOperations tracks cache operation metrics
	CacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"},
	)

	// CacheOperationDuration tracks cache operation duration
	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// TransactionQueueSize tracks the current size of the transaction processing queue
	TransactionQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "transaction_queue_size",
			Help: "Current number of tasks in the transaction processing queue",
		},
	)

	// TransactionProcessingDuration tracks transaction processing duration
	TransactionProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_processing_duration_seconds",
			Help:    "Transaction processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"transaction_type"},
	)

	// TransactionProcessingSuccess tracks successful transaction processing
	TransactionProcessingSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_processing_success_total",
			Help: "Total number of successfully processed transactions",
		},
		[]string{"transaction_type"},
	)

	// TransactionProcessingFailure tracks failed transaction processing
	TransactionProcessingFailure = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_processing_failure_total",
			Help: "Total number of failed transaction processing attempts",
		},
		[]string{"transaction_type"},
	)

	// ===== BUSINESS METRICS =====

	// UserRegistrationTotal tracks total user registrations
	UserRegistrationTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registration_total",
			Help: "Total number of user registrations",
		},
	)

	// ActiveUsers tracks currently active users
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of currently active users",
		},
	)

	// UserLoginTotal tracks total user logins
	UserLoginTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_total",
			Help: "Total number of user logins",
		},
		[]string{"status"}, // success, failure
	)

	// TransactionVolume tracks total transaction volume in currency units
	TransactionVolume = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_volume_total",
			Help: "Total transaction volume in currency units",
		},
		[]string{"transaction_type", "status"}, // credit, debit, transfer, success, failed
	)

	// TransactionCount tracks total number of transactions
	TransactionCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_count_total",
			Help: "Total number of transactions",
		},
		[]string{"transaction_type", "status"}, // credit, debit, transfer, success, failed
	)

	// AverageTransactionAmount tracks average transaction amount
	AverageTransactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_amount",
			Help:    "Transaction amounts for calculating averages",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"transaction_type"},
	)

	// BalanceTotal tracks total balance across all users
	BalanceTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "balance_total",
			Help: "Total balance across all users",
		},
	)

	// BalanceDistribution tracks balance distribution across users
	BalanceDistribution = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "user_balance_distribution",
			Help:    "Distribution of user balances",
			Buckets: []float64{0, 100, 500, 1000, 5000, 10000, 50000, 100000},
		},
	)

	// TransactionSuccessRate tracks transaction success rate
	TransactionSuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transaction_success_rate",
			Help: "Transaction success rate as a percentage",
		},
		[]string{"transaction_type"},
	)

	// DailyActiveUsers tracks daily active users
	DailyActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "daily_active_users",
			Help: "Number of daily active users",
		},
	)

	// MonthlyActiveUsers tracks monthly active users
	MonthlyActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "monthly_active_users",
			Help: "Number of monthly active users",
		},
	)

	// RevenueMetrics tracks revenue-related metrics
	RevenueMetrics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "revenue_total",
			Help: "Total revenue generated",
		},
		[]string{"revenue_type"}, // fees, commissions, etc.
	)

	// ErrorRate tracks error rates by type
	ErrorRate = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"error_type", "severity"}, // validation, database, auth, critical, warning, info
	)

	// SystemHealth tracks system health indicators
	SystemHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_health",
			Help: "System health indicators (1 = healthy, 0 = unhealthy)",
		},
		[]string{"component"}, // database, redis, api
	)

	// CacheHitRatio tracks cache hit ratio
	CacheHitRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_hit_ratio",
			Help: "Cache hit ratio as a percentage",
		},
	)

	// DatabaseConnectionPool tracks database connection pool metrics
	DatabaseConnectionPool = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connection_pool",
			Help: "Database connection pool metrics",
		},
		[]string{"state"}, // active, idle, total
	)

	// APIResponseTimePercentiles tracks API response time percentiles
	APIResponseTimePercentiles = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_response_time_seconds",
			Help:    "API response time for percentile calculations",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"endpoint", "method"},
	)
)
