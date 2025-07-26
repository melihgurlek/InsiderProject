package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "credit successful"})
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
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
	var req struct {
		FromUserID int     `json:"from_user_id"`
		ToUserID   int     `json:"to_user_id"`
		Amount     float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err := h.service.Transfer(req.FromUserID, req.ToUserID, float64(req.Amount))
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "transfer successful"})
}

func (h *TransactionHandler) ListAllTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := h.service.ListAllTransactions()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid transaction id")
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
	userID := chi.URLParam(r, "user_id")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	transactions, err := h.service.ListUserTransactions(userIDInt)
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
