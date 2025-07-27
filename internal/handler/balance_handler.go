package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// BalanceHandler handles balance-related HTTP requests.
type BalanceHandler struct {
	service domain.BalanceService
}

// NewBalanceHandler creates a new BalanceHandler.
func NewBalanceHandler(service domain.BalanceService) *BalanceHandler {
	return &BalanceHandler{service: service}
}

// RegisterRoutes registers balance endpoints to the router.
func (h *BalanceHandler) RegisterRoutes(r chi.Router) {
	r.Get("/balances/current/{user_id}", h.GetCurrentBalance)
	r.Get("/balances/historical/{user_id}", h.GetHistoricalBalance)
	r.Get("/balances/at-time/{user_id}", h.GetBalanceAtTime)
}

func (h *BalanceHandler) GetCurrentBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userIDInt, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	balance, err := h.service.GetCurrentBalance(userIDInt)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) GetHistoricalBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userIDInt, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	balance, err := h.service.GetHistoricalBalance(userIDInt)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) GetBalanceAtTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	timeStr := r.URL.Query().Get("time")
	if timeStr == "" {
		h.respondError(w, http.StatusBadRequest, "missing time parameter")
		return
	}
	queryTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid time format")
		return
	}

	balance, err := h.service.GetBalanceAtTime(userID, queryTime)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) respondError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
