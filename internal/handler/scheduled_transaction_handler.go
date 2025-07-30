package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/internal/middleware"
)

// ScheduledTransactionHandler handles HTTP requests for scheduled transactions
type ScheduledTransactionHandler struct {
	scheduledService domain.ScheduledTransactionService
}

// NewScheduledTransactionHandler creates a new ScheduledTransactionHandler
func NewScheduledTransactionHandler(scheduledService domain.ScheduledTransactionService) *ScheduledTransactionHandler {
	return &ScheduledTransactionHandler{
		scheduledService: scheduledService,
	}
}

// RegisterRoutes registers the scheduled transaction routes
func (h *ScheduledTransactionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateScheduledTransaction)
	r.Get("/", h.ListUserScheduledTransactions)
	r.Get("/stats", h.GetScheduledTransactionStats)
	r.Get("/{id}", h.GetScheduledTransaction)
	r.Put("/{id}", h.UpdateScheduledTransaction)
	r.Delete("/{id}", h.CancelScheduledTransaction)
	r.Post("/execute", h.ExecuteScheduledTransactions)
}

// CreateScheduledTransactionRequest represents a request to create a scheduled transaction
type CreateScheduledTransactionRequest struct {
	UserID      int       `json:"user_id"`
	ToUserID    *int      `json:"to_user_id,omitempty"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	ScheduleAt  time.Time `json:"schedule_at"`
	Recurring   bool      `json:"recurring"`
	Recurrence  string    `json:"recurrence,omitempty"`
	MaxRuns     *int      `json:"max_runs,omitempty"`
	Description string    `json:"description,omitempty"`
}

// CreateScheduledTransaction handles creation of a new scheduled transaction
func (h *ScheduledTransactionHandler) CreateScheduledTransaction(w http.ResponseWriter, r *http.Request) {
	// The middleware has already validated the request body.
	req, ok := middleware.GetValidatedBody[*CreateScheduledTransactionRequest](r.Context())
	if !ok {
		// This panic will be caught by the error middleware, indicating a server setup issue.
		panic("could not retrieve validated body")
	}

	st := &domain.ScheduledTransaction{
		UserID:      req.UserID,
		ToUserID:    req.ToUserID,
		Amount:      req.Amount,
		Type:        req.Type,
		ScheduleAt:  req.ScheduleAt,
		Recurring:   req.Recurring,
		Recurrence:  req.Recurrence,
		MaxRuns:     req.MaxRuns,
		Description: req.Description,
	}

	// The service layer will perform the final, deeper business logic validation
	if err := h.scheduledService.CreateScheduledTransaction(st); err != nil {
		// Check if it's a validation error from the service layer
		var valErr *domain.ValidationError
		if errors.As(err, &valErr) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Otherwise, it's an internal server error
		log.Error().Err(err).Msg("Failed to create scheduled transaction")
		h.respondError(w, http.StatusInternalServerError, "failed to create scheduled transaction")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(st)
}

// GetScheduledTransaction handles retrieval of a scheduled transaction by ID
func (h *ScheduledTransactionHandler) GetScheduledTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid scheduled transaction ID")
		return
	}

	st, err := h.scheduledService.GetScheduledTransaction(id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to get scheduled transaction")
		h.respondError(w, http.StatusInternalServerError, "failed to get scheduled transaction: "+err.Error())
		return
	}

	if st == nil {
		h.respondError(w, http.StatusNotFound, "scheduled transaction not found")
		return
	}

	json.NewEncoder(w).Encode(st)
}

// ListUserScheduledTransactions handles listing scheduled transactions for a user
func (h *ScheduledTransactionHandler) ListUserScheduledTransactions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	transactions, err := h.scheduledService.ListUserScheduledTransactions(userID)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to list user scheduled transactions")
		h.respondError(w, http.StatusInternalServerError, "failed to list scheduled transactions: "+err.Error())
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

// UpdateScheduledTransactionRequest represents a request to update a scheduled transaction
type UpdateScheduledTransactionRequest struct {
	Amount      *float64   `json:"amount,omitempty" validate:"omitempty,gt=0"`
	ScheduleAt  *time.Time `json:"schedule_at,omitempty"`
	Recurring   *bool      `json:"recurring,omitempty"`
	Recurrence  *string    `json:"recurrence,omitempty" validate:"omitempty,oneof=daily weekly monthly yearly"`
	MaxRuns     *int       `json:"max_runs,omitempty" validate:"omitempty,min=1"`
	Description *string    `json:"description,omitempty"`
}

// Validate checks the request data. This method is called by the new middleware.
func (req *CreateScheduledTransactionRequest) Validate() error {
	if req.Type == "transfer" && req.ToUserID == nil {
		return errors.New("transfer requires to_user_id")
	}
	if req.Type == "transfer" && req.UserID == *req.ToUserID {
		return errors.New("cannot transfer to self")
	}
	// The domain object will handle deeper validation like time checks
	return nil
}

// UpdateScheduledTransaction handles updating a scheduled transaction
func (h *ScheduledTransactionHandler) UpdateScheduledTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid scheduled transaction ID")
		return
	}

	var req UpdateScheduledTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get existing scheduled transaction
	existing, err := h.scheduledService.GetScheduledTransaction(id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to get existing scheduled transaction")
		h.respondError(w, http.StatusInternalServerError, "failed to get scheduled transaction: "+err.Error())
		return
	}

	if existing == nil {
		h.respondError(w, http.StatusNotFound, "scheduled transaction not found")
		return
	}

	// Update fields if provided
	if req.Amount != nil {
		existing.Amount = *req.Amount
	}
	if req.ScheduleAt != nil {
		existing.ScheduleAt = *req.ScheduleAt
	}
	if req.Recurring != nil {
		existing.Recurring = *req.Recurring
	}
	if req.Recurrence != nil {
		existing.Recurrence = *req.Recurrence
	}
	if req.MaxRuns != nil {
		existing.MaxRuns = req.MaxRuns
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}

	// Recalculate next run for recurring transactions
	if existing.Recurring {
		existing.NextRunAt = existing.CalculateNextRun()
	}

	if err := h.scheduledService.UpdateScheduledTransaction(existing); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to update scheduled transaction")
		h.respondError(w, http.StatusInternalServerError, "failed to update scheduled transaction: "+err.Error())
		return
	}

	json.NewEncoder(w).Encode(existing)
}

// CancelScheduledTransaction handles cancellation of a scheduled transaction
func (h *ScheduledTransactionHandler) CancelScheduledTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid scheduled transaction ID")
		return
	}

	if err := h.scheduledService.CancelScheduledTransaction(id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to cancel scheduled transaction")
		h.respondError(w, http.StatusInternalServerError, "failed to cancel scheduled transaction: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetScheduledTransactionStats handles retrieval of scheduled transaction statistics
func (h *ScheduledTransactionHandler) GetScheduledTransactionStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.scheduledService.GetScheduledTransactionStats()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get scheduled transaction stats")
		h.respondError(w, http.StatusInternalServerError, "failed to get scheduled transaction stats: "+err.Error())
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// ExecuteScheduledTransactions handles manual execution of pending scheduled transactions
func (h *ScheduledTransactionHandler) ExecuteScheduledTransactions(w http.ResponseWriter, r *http.Request) {
	if err := h.scheduledService.ExecuteScheduledTransactions(); err != nil {
		log.Error().Err(err).Msg("Failed to execute scheduled transactions")
		h.respondError(w, http.StatusInternalServerError, "failed to execute scheduled transactions: "+err.Error())
		return
	}

	response := map[string]string{
		"message": "Scheduled transactions execution completed",
		"status":  "success",
	}

	json.NewEncoder(w).Encode(response)
}

// respondError is a helper method to respond with error
func (h *ScheduledTransactionHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
