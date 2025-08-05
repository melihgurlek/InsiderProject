package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/internal/middleware"
)

// TransactionHandler handles transaction-related HTTP requests.
type TransactionHandler struct {
	service domain.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(service domain.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/transactions/credit", h.Credit)
	r.Post("/transactions/debit", h.Debit)
	r.Post("/transactions/transfer", h.Transfer)
	r.Get("/transactions/history", h.ListAllTransactions)
	r.Get("/transactions/{id}", h.GetTransactionByID)
	r.Get("/transactions/user/{user_id}", h.ListUserTransactions)
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	// Only admins can credit an account.
	if claims.Role != "admin" {
		h.respondError(w, http.StatusForbidden, "you do not have permission to perform this action")
		return
	}

	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err := h.service.Credit(req.UserID, float64(req.Amount))
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "credit successful"})
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// A user can only transfer from their own account, unless they are an admin.
	if claims.Role != "admin" && claims.UserID != strconv.Itoa(req.UserID) {
		h.respondError(w, http.StatusForbidden, "you can only debit your own account")
		return
	}

	err := h.service.Debit(req.UserID, float64(req.Amount))
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "debit successful"})
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	var req struct {
		FromUserID int     `json:"from_user_id"`
		ToUserID   int     `json:"to_user_id"`
		Amount     float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// A user can only transfer from their own account, unless they are an admin.
	if claims.Role != "admin" && claims.UserID != strconv.Itoa(req.FromUserID) {
		h.respondError(w, http.StatusForbidden, "you can only transfer from your own account")
		return
	}

	err := h.service.Transfer(req.FromUserID, req.ToUserID, float64(req.Amount))
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "transfer successful"})
}

func (h *TransactionHandler) ListAllTransactions(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	// Get limit and offset from query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100 // Default limit
	offset := 0  // Default offset

	if limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	if offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	if claims.Role != "admin" {
		h.respondError(w, http.StatusForbidden, "you do not have permission to list transactions")
		return
	}

	transactions, err := h.service.ListAllTransactions(r.Context(), limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {

	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	if claims.Role != "admin" && claims.UserID != strconv.Itoa(idInt) {
		h.respondError(w, http.StatusForbidden, "you do not have permission to view this transaction")
		return
	}

	transaction, err := h.service.GetTransaction(idInt)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) ListUserTransactions(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	targetIDStr := chi.URLParam(r, "user_id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// A user can only list their own transactions, unless they are an admin.
	if claims.Role != "admin" && claims.UserID != strconv.Itoa(targetID) {
		h.respondError(w, http.StatusForbidden, "you do not have permission to view these transactions")
		return
	}

	transactions, err := h.service.ListUserTransactions(targetID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}
func (h *TransactionHandler) respondError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
