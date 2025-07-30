package service

import (
	"context"
	"sync"
	"time"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/pkg/metrics"
	"github.com/rs/zerolog/log"
)

// BusinessMetricsService handles business metrics collection and updates
type BusinessMetricsService struct {
	userRepo        domain.UserRepository
	transactionRepo domain.TransactionRepository
	balanceRepo     domain.BalanceRepository
	mu              sync.RWMutex
	lastUpdate      time.Time
	updateInterval  time.Duration
	stopChan        chan struct{}
}

// NewBusinessMetricsService creates a new business metrics service
func NewBusinessMetricsService(
	userRepo domain.UserRepository,
	transactionRepo domain.TransactionRepository,
	balanceRepo domain.BalanceRepository,
) *BusinessMetricsService {
	return &BusinessMetricsService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		balanceRepo:     balanceRepo,
		updateInterval:  5 * time.Minute, // Update metrics every 5 minutes
		stopChan:        make(chan struct{}),
	}
}

// Start begins the background metrics collection
func (s *BusinessMetricsService) Start(ctx context.Context) {
	log.Info().Msg("Starting business metrics service")

	go s.metricsCollector(ctx)
}

// Stop stops the background metrics collection
func (s *BusinessMetricsService) Stop() {
	log.Info().Msg("Stopping business metrics service")
	close(s.stopChan)
}

// metricsCollector runs in the background to collect and update business metrics
func (s *BusinessMetricsService) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	// Initial collection
	s.collectMetrics(ctx, 1000, 0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.collectMetrics(ctx, 1000, 0)
		}
	}
}

// collectMetrics collects all business metrics from the database
func (s *BusinessMetricsService) collectMetrics(ctx context.Context, limit int, offset int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Debug().Msg("Collecting business metrics")

	// Collect user metrics
	s.collectUserMetrics(ctx)

	// Collect transaction metrics
	s.collectTransactionMetrics(ctx, limit, offset)

	// Collect balance metrics
	s.collectBalanceMetrics(ctx)

	// Collect system health metrics
	s.collectSystemHealthMetrics(ctx)

	s.lastUpdate = time.Now()
}

// collectUserMetrics collects user-related metrics
func (s *BusinessMetricsService) collectUserMetrics(ctx context.Context) {
	// Get total user count
	users, err := s.userRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users for metrics")
		metrics.ErrorRate.WithLabelValues("database", "warning").Inc()
		return
	}

	// Calculate active users (users with recent activity)
	activeUsers := 0
	dailyActiveUsers := 0
	monthlyActiveUsers := 0
	now := time.Now()

	for _, user := range users {
		// Simple logic: consider users active if they have recent transactions
		// In a real system, you'd track user sessions or activity timestamps
		if user.UpdatedAt.After(now.Add(-24 * time.Hour)) {
			dailyActiveUsers++
		}
		if user.UpdatedAt.After(now.Add(-30 * 24 * time.Hour)) {
			monthlyActiveUsers++
		}
		if user.UpdatedAt.After(now.Add(-1 * time.Hour)) {
			activeUsers++
		}
	}

	metrics.ActiveUsers.Set(float64(activeUsers))
	metrics.DailyActiveUsers.Set(float64(dailyActiveUsers))
	metrics.MonthlyActiveUsers.Set(float64(monthlyActiveUsers))
}

// collectTransactionMetrics collects transaction-related metrics
func (s *BusinessMetricsService) collectTransactionMetrics(ctx context.Context, limit int, offset int) {
	// Get all transactions
	transactions, err := s.transactionRepo.ListAll(ctx, limit, offset) // Default limit of 1000, offset 0
	if err != nil {
		log.Error().Err(err).Msg("Failed to get transactions for metrics")
		metrics.ErrorRate.WithLabelValues("database", "warning").Inc()
		return
	}

	// Initialize counters
	transactionCounts := make(map[string]map[string]int)
	transactionVolumes := make(map[string]map[string]float64)
	successCounts := make(map[string]int)
	totalCounts := make(map[string]int)

	// Initialize maps
	for _, txnType := range []string{"credit", "debit", "transfer"} {
		transactionCounts[txnType] = make(map[string]int)
		transactionVolumes[txnType] = make(map[string]float64)
	}

	// Process transactions
	for _, txn := range transactions {
		txnType := string(txn.Type)
		status := string(txn.Status)

		// Count transactions
		transactionCounts[txnType][status]++
		totalCounts[txnType]++

		// Track volumes
		transactionVolumes[txnType][status] += float64(txn.Amount)

		// Track success rates
		if txn.Status == "completed" {
			successCounts[txnType]++
		}

		// Record transaction amount for histogram
		metrics.AverageTransactionAmount.WithLabelValues(txnType).Observe(float64(txn.Amount))
	}

	// Update Prometheus metrics
	for txnType, statusCounts := range transactionCounts {
		for status, count := range statusCounts {
			metrics.TransactionCount.WithLabelValues(txnType, status).Add(float64(count))
		}
	}

	for txnType, statusVolumes := range transactionVolumes {
		for status, volume := range statusVolumes {
			metrics.TransactionVolume.WithLabelValues(txnType, status).Add(volume)
		}
	}

	// Calculate and update success rates
	for txnType, total := range totalCounts {
		if total > 0 {
			successRate := float64(successCounts[txnType]) / float64(total) * 100
			metrics.TransactionSuccessRate.WithLabelValues(txnType).Set(successRate)
		}
	}
}

// collectBalanceMetrics collects balance-related metrics
func (s *BusinessMetricsService) collectBalanceMetrics(ctx context.Context) {
	// Get all balances - we'll need to get them from users
	users, err := s.userRepo.List()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users for balance metrics")
		metrics.ErrorRate.WithLabelValues("database", "warning").Inc()
		return
	}

	// Calculate total balance
	totalBalance := float64(0)
	for _, user := range users {
		balance, err := s.balanceRepo.GetByUserID(user.ID)
		if err != nil {
			log.Error().Err(err).Int("user_id", user.ID).Msg("Failed to get balance for user")
			continue
		}
		if balance != nil {
			totalBalance += float64(balance.Amount)
			// Record balance distribution
			metrics.BalanceDistribution.Observe(float64(balance.Amount))
		}
	}

	metrics.BalanceTotal.Set(totalBalance)
}

// collectSystemHealthMetrics collects system health indicators
func (s *BusinessMetricsService) collectSystemHealthMetrics(ctx context.Context) {
	//Use the Ping method for a real health check.
	if err := s.userRepo.Ping(ctx); err != nil {
		log.Error().Err(err).Msg("Database health check failed")
		metrics.SystemHealth.WithLabelValues("database").Set(0.0) // 0 for unhealthy
	} else {
		metrics.SystemHealth.WithLabelValues("database").Set(1.0) // 1 for healthy
	}

	// Cache hit ratio calculation (this would be updated by cache middleware)
	// For now, we'll set a default value
	metrics.CacheHitRatio.Set(85.0) // 85% cache hit ratio

	// API health (assuming healthy if we can reach this point)
	metrics.SystemHealth.WithLabelValues("api").Set(1.0)
}

// RecordUserRegistration records a new user registration
func (s *BusinessMetricsService) RecordUserRegistration() {
	metrics.UserRegistrationTotal.Inc()
}

// RecordUserLogin records a user login attempt
func (s *BusinessMetricsService) RecordUserLogin(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	metrics.UserLoginTotal.WithLabelValues(status).Inc()
}

// RecordTransaction records a transaction for metrics
func (s *BusinessMetricsService) RecordTransaction(txnType string, amount int64, success bool) {
	status := "failed"
	if success {
		status = "success"
	}

	metrics.TransactionCount.WithLabelValues(txnType, status).Inc()
	metrics.TransactionVolume.WithLabelValues(txnType, status).Add(float64(amount))
	metrics.AverageTransactionAmount.WithLabelValues(txnType).Observe(float64(amount))
}

// RecordError records an error for metrics
func (s *BusinessMetricsService) RecordError(errorType, severity string) {
	metrics.ErrorRate.WithLabelValues(errorType, severity).Inc()
}

// UpdateCacheHitRatio updates the cache hit ratio
func (s *BusinessMetricsService) UpdateCacheHitRatio(hitRatio float64) {
	metrics.CacheHitRatio.Set(hitRatio)
}

// UpdateDatabaseConnectionPool updates database connection pool metrics
func (s *BusinessMetricsService) UpdateDatabaseConnectionPool(active, idle, total int) {
	metrics.DatabaseConnectionPool.WithLabelValues("active").Set(float64(active))
	metrics.DatabaseConnectionPool.WithLabelValues("idle").Set(float64(idle))
	metrics.DatabaseConnectionPool.WithLabelValues("total").Set(float64(total))
}

// GetMetricsSummary returns a summary of current metrics
func (s *BusinessMetricsService) GetMetricsSummary(ctx context.Context) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get current values from Prometheus metrics
	summary := map[string]interface{}{
		"last_update":          s.lastUpdate,
		"active_users":         metrics.ActiveUsers,
		"daily_active_users":   metrics.DailyActiveUsers,
		"monthly_active_users": metrics.MonthlyActiveUsers,
		"balance_total":        metrics.BalanceTotal,
		"cache_hit_ratio":      metrics.CacheHitRatio,
	}

	return summary
}
