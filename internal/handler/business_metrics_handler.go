package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/melihgurlek/backend-path/internal/service"
)

// BusinessMetricsHandler handles business metrics API endpoints
type BusinessMetricsHandler struct {
	businessMetricsService *service.BusinessMetricsService
}

// NewBusinessMetricsHandler creates a new BusinessMetricsHandler
func NewBusinessMetricsHandler(businessMetricsService *service.BusinessMetricsService) *BusinessMetricsHandler {
	return &BusinessMetricsHandler{
		businessMetricsService: businessMetricsService,
	}
}

// RegisterRoutes registers the business metrics routes
func (h *BusinessMetricsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/summary", h.GetMetricsSummary)
	r.Get("/kpis", h.GetKeyPerformanceIndicators)
}

// GetMetricsSummary returns a summary of current business metrics
func (h *BusinessMetricsHandler) GetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary := h.businessMetricsService.GetMetricsSummary(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Error().Err(err).Msg("Failed to encode metrics summary")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetKeyPerformanceIndicators returns key performance indicators
func (h *BusinessMetricsHandler) GetKeyPerformanceIndicators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get metrics summary
	summary := h.businessMetricsService.GetMetricsSummary(ctx)

	// Calculate KPIs
	kpis := map[string]interface{}{
		"user_metrics": map[string]interface{}{
			"active_users":         summary["active_users"],
			"daily_active_users":   summary["daily_active_users"],
			"monthly_active_users": summary["monthly_active_users"],
		},
		"financial_metrics": map[string]interface{}{
			"total_balance":   summary["balance_total"],
			"cache_hit_ratio": summary["cache_hit_ratio"],
		},
		"system_health": map[string]interface{}{
			"last_update": summary["last_update"],
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(kpis); err != nil {
		log.Error().Err(err).Msg("Failed to encode KPIs")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
