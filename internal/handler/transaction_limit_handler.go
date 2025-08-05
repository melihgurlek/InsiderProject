package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/melihgurlek/backend-path/internal/domain"

	"github.com/go-chi/chi/v5"
)

type TransactionLimitHandler struct {
	Service domain.TransactionLimitService
}

func NewTransactionLimitHandler(service domain.TransactionLimitService) *TransactionLimitHandler {
	return &TransactionLimitHandler{Service: service}
}

func (h *TransactionLimitHandler) RegisterRoutes(r chi.Router) {
	r.Route("/users/{userID}/limits", func(r chi.Router) {
		r.Get("/", h.ListRules)
		r.Post("/", h.AddRule)
		r.Delete("/{ruleID}", h.RemoveRule)
	})
}

func (h *TransactionLimitHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}
	rules, err := h.Service.ListRules(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if rules == nil {
		rules = []domain.TransactionLimitRule{}
	}
	json.NewEncoder(w).Encode(rules)
}

type addRuleRequest struct {
	RuleType    string        `json:"rule_type"`
	LimitAmount float64       `json:"limit_amount"`
	Currency    string        `json:"currency"`
	Window      time.Duration `json:"window"`
	Active      bool          `json:"active"`
}

func (h *TransactionLimitHandler) AddRule(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}
	var req addRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.RuleType == "" || req.LimitAmount <= 0 {
		http.Error(w, "missing or invalid rule_type or limit_amount", http.StatusBadRequest)
		return
	}
	rule := domain.TransactionLimitRule{
		ID:          "",
		UserID:      userID,
		RuleType:    domain.RuleType(req.RuleType),
		LimitAmount: req.LimitAmount,
		Currency:    req.Currency,
		Window:      req.Window,
		Active:      req.Active,
	}
	rule, err = h.Service.AddRule(r.Context(), rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rule)
	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionLimitHandler) RemoveRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleID")
	if err := h.Service.RemoveRule(r.Context(), ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
