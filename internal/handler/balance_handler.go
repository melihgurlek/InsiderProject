package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/internal/middleware"
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
	r.Get("/balances/current", h.GetCurrentBalance)
	r.Get("/balances/historical", h.GetHistoricalBalance)
	r.Get("/balances/at-time", h.GetBalanceAtTime)
}

func (h *BalanceHandler) GetCurrentBalance(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: GetCurrentBalance called\n")

	targetID, err := authorizeAndGetTargetID(r)
	if err != nil {
		fmt.Printf("DEBUG: authorizeAndGetTargetID error: %v\n", err)
		if he, ok := err.(*handlerError); ok {
			h.respondError(w, he.statusCode, he.message)
		} else {
			h.respondError(w, http.StatusInternalServerError, "an internal server error occurred")
		}
		return
	}

	fmt.Printf("DEBUG: targetID: %d\n", targetID)

	balance, err := h.service.GetCurrentBalance(targetID)
	if err != nil {
		fmt.Printf("DEBUG: GetCurrentBalance service error: %v\n", err)
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Printf("DEBUG: balance: %+v\n", balance)

	// If no balance record exists, return a default balance with 0 amount
	if balance == nil {
		fmt.Printf("DEBUG: balance is nil, creating default\n")
		balance = &domain.Balance{
			UserID:        targetID,
			Amount:        0,
			LastUpdatedAt: time.Now(),
		}
	}

	fmt.Printf("DEBUG: about to encode balance: %+v\n", balance)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		fmt.Printf("DEBUG: JSON encode error: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: GetCurrentBalance completed successfully\n")
}

func (h *BalanceHandler) GetHistoricalBalance(w http.ResponseWriter, r *http.Request) {
	targetID, err := authorizeAndGetTargetID(r)
	if err != nil {
		if he, ok := err.(*handlerError); ok {
			h.respondError(w, he.statusCode, he.message)
		} else {
			h.respondError(w, http.StatusInternalServerError, "an internal server error occurred")
		}
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	balances, err := h.service.GetHistoricalBalance(targetID, limit)
	if err != nil {
		if he, ok := err.(*handlerError); ok {
			h.respondError(w, he.statusCode, he.message)
		} else {
			h.respondError(w, http.StatusInternalServerError, "an internal server error occurred")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balances)
}

func (h *BalanceHandler) GetBalanceAtTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	targetID, err := authorizeAndGetTargetID(r)
	if err != nil {
		if he, ok := err.(*handlerError); ok {
			h.respondError(w, he.statusCode, he.message)
		} else {
			h.respondError(w, http.StatusInternalServerError, "an internal server error occurred")
		}
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

	balance, err := h.service.GetBalanceAtTime(targetID, queryTime)
	if err != nil {
		if he, ok := err.(*handlerError); ok {
			h.respondError(w, he.statusCode, he.message)
		} else {
			h.respondError(w, http.StatusInternalServerError, "an internal server error occurred")
		}
		return
	}

	// If no balance record exists for the given time, return a default balance
	if balance == nil {
		balance = &domain.Balance{
			UserID:        targetID,
			Amount:        0,
			LastUpdatedAt: queryTime,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) respondError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func authorizeAndGetTargetID(r *http.Request) (int, error) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		return 0, &handlerError{statusCode: http.StatusUnauthorized, message: "invalid token claims"}
	}

	targetUserIDStr := r.URL.Query().Get("user_id")
	if targetUserIDStr != "" {
		if claims.Role != "admin" {
			return 0, &handlerError{statusCode: http.StatusForbidden, message: "you do not have permission to view other users' balances"}
		}
		targetID, err := strconv.Atoi(targetUserIDStr)
		if err != nil {
			return 0, &handlerError{statusCode: http.StatusBadRequest, message: "invalid user_id in query parameter"}
		}
		return targetID, nil
	}

	targetID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return 0, &handlerError{statusCode: http.StatusInternalServerError, message: "invalid user_id in token"}
	}
	return targetID, nil
}

type handlerError struct {
	statusCode int
	message    string
}

func (e *handlerError) Error() string {
	return e.message
}
