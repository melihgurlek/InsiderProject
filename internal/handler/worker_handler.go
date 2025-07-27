package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/internal/worker"
)

// WorkerHandler handles worker-related HTTP requests
type WorkerHandler struct {
	transactionProcessor domain.TransactionProcessor
	batchProcessor       *worker.BatchProcessor
}

// NewWorkerHandler creates a new WorkerHandler
func NewWorkerHandler(transactionProcessor domain.TransactionProcessor, bp *worker.BatchProcessor) *WorkerHandler {
	return &WorkerHandler{
		transactionProcessor: transactionProcessor,
		batchProcessor:       bp,
	}
}

// RegisterRoutes registers the worker routes
func (h *WorkerHandler) RegisterRoutes(r chi.Router) {
	r.Post("/tasks", h.SubmitTask)
	r.Post("/batch", h.SubmitBatch)
	r.Get("/stats", h.GetStats)
	r.Get("/health", h.GetHealth)
}

// SubmitTaskRequest represents a request to submit a single task
type SubmitTaskRequest struct {
	Type     string  `json:"type" validate:"required,oneof=credit debit transfer"`
	UserID   int     `json:"user_id" validate:"required,min=1"`
	ToUserID *int    `json:"to_user_id,omitempty"` // for transfers
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Priority int     `json:"priority,omitempty" validate:"min=0,max=10"`
}

// SubmitTaskResponse represents the response for task submission
type SubmitTaskResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// SubmitTask handles submission of a single transaction task
func (h *WorkerHandler) SubmitTask(w http.ResponseWriter, r *http.Request) {
	var req SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := h.validateSubmitTaskRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create task
	task := &domain.TransactionTask{
		ID:       uuid.New().String(),
		Type:     req.Type,
		UserID:   req.UserID,
		ToUserID: req.ToUserID,
		Amount:   req.Amount,
		Priority: req.Priority,
	}

	// Submit task
	err := h.transactionProcessor.SubmitTask(r.Context(), task)
	if err != nil {
		log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to submit task")
		h.respondError(w, http.StatusInternalServerError, "failed to submit task: "+err.Error())
		return
	}

	response := SubmitTaskResponse{
		TaskID:    task.ID,
		Status:    "submitted",
		Message:   "Task submitted successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// SubmitBatchRequest represents a request to submit multiple tasks
type SubmitBatchRequest struct {
	Tasks []SubmitTaskRequest `json:"tasks" validate:"required,min=1,max=100"`
}

// SubmitBatchResponse represents the response for batch submission
type SubmitBatchResponse struct {
	BatchID   string   `json:"batch_id"`
	TaskIDs   []string `json:"task_ids"`
	Status    string   `json:"status"`
	Message   string   `json:"message"`
	Timestamp int64    `json:"timestamp"`
}

// SubmitBatch handles submission of multiple transaction tasks
func (h *WorkerHandler) SubmitBatch(w http.ResponseWriter, r *http.Request) {
	var req SubmitBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate the batch request itself
	if len(req.Tasks) == 0 {
		h.respondError(w, http.StatusBadRequest, "at least one task is required")
		return
	}

	if len(req.Tasks) > 100 {
		h.respondError(w, http.StatusBadRequest, "maximum 100 tasks allowed per batch")
		return
	}

	// Convert request tasks to domain tasks and validate each one
	tasks := make([]*domain.TransactionTask, len(req.Tasks))
	for i, taskReq := range req.Tasks {
		if err := h.validateSubmitTaskRequest(&taskReq); err != nil {
			msg := fmt.Sprintf("invalid task at index %d: %s", i, err.Error())
			h.respondError(w, http.StatusBadRequest, msg)
			return
		}

		tasks[i] = &domain.TransactionTask{
			ID:       uuid.New().String(),
			Type:     taskReq.Type,
			UserID:   taskReq.UserID,
			ToUserID: taskReq.ToUserID,
			Amount:   taskReq.Amount,
			Priority: taskReq.Priority,
		}
	}

	// Run the batch processing in a background goroutine so the API can respond immediately.
	go func() {
		// Create a new background context because the original request's context
		// will be canceled as soon as this HTTP handler returns.
		bgCtx := context.Background()

		log.Info().Int("task_count", len(tasks)).Msg("Starting asynchronous batch processing")
		result, err := h.batchProcessor.ProcessBatch(bgCtx, tasks)
		if err != nil {
			// This log captures errors from the batch execution itself
			log.Error().Err(err).Msg("Asynchronous batch processing failed")
			return
		}
		// This log confirms the final outcome of the async job
		log.Info().
			Str("batch_id", result.BatchID).
			Int("successful", result.SuccessfulTasks).
			Int("failed", result.FailedTasks).
			Msg("Asynchronous batch processing finished")
	}()

	// Immediately send a response to the client acknowledging the submission.
	response := SubmitBatchResponse{
		BatchID:   uuid.New().String(), // This is an acknowledgment ID for the submission
		Status:    "submitted",
		Message:   "Batch submitted for asynchronous processing.",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// GetStatsResponse represents the response for processing statistics
type GetStatsResponse struct {
	TotalProcessed     int64   `json:"total_processed"`
	SuccessfulTasks    int64   `json:"successful_tasks"`
	FailedTasks        int64   `json:"failed_tasks"`
	QueueSize          int     `json:"queue_size"`
	ActiveWorkers      int     `json:"active_workers"`
	AverageProcessTime float64 `json:"average_process_time_seconds"`
	Timestamp          int64   `json:"timestamp"`
}

// GetStats returns current processing statistics
func (h *WorkerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.transactionProcessor.GetStats()

	response := GetStatsResponse{
		TotalProcessed:     stats.TotalProcessed,
		SuccessfulTasks:    stats.SuccessfulTasks,
		FailedTasks:        stats.FailedTasks,
		QueueSize:          stats.QueueSize,
		ActiveWorkers:      stats.ActiveWorkers,
		AverageProcessTime: stats.AverageProcessTime,
		Timestamp:          time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetHealthResponse represents the health check response
type GetHealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// GetHealth returns the health status of the worker system
func (h *WorkerHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	stats := h.transactionProcessor.GetStats()

	response := GetHealthResponse{
		Status:    "healthy",
		Message:   "Worker system is operational",
		Timestamp: time.Now().Unix(),
	}

	// Check if queue is getting too full
	if stats.QueueSize > 1000 {
		response.Status = "warning"
		response.Message = "Queue size is high"
	}

	// Check if there are too many failed tasks
	if stats.FailedTasks > 0 && float64(stats.FailedTasks)/float64(stats.TotalProcessed) > 0.1 {
		response.Status = "warning"
		response.Message = "High failure rate detected"
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// validateSubmitTaskRequest validates a task submission request
func (h *WorkerHandler) validateSubmitTaskRequest(req *SubmitTaskRequest) error {
	if req.Type == "" {
		return errors.New("type is required")
	}

	if req.Type != "credit" && req.Type != "debit" && req.Type != "transfer" {
		return errors.New("type must be credit, debit, or transfer")
	}

	if req.UserID <= 0 {
		return errors.New("user_id must be positive")
	}

	if req.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if req.Type == "transfer" && req.ToUserID == nil {
		return errors.New("to_user_id is required for transfer type")
	}

	if req.Type == "transfer" && *req.ToUserID <= 0 {
		return errors.New("to_user_id must be positive")
	}

	if req.Priority < 0 || req.Priority > 10 {
		return errors.New("priority must be between 0 and 10")
	}

	return nil
}

// respondError sends an error response
func (h *WorkerHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
