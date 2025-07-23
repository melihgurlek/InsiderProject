package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	service domain.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterRoutes registers user auth routes to the router.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)

	// User CRUD
	r.Get("/users", h.ListUsers)
	r.Get("/users/{id}", h.GetUserByID)
	r.Put("/users/{id}", h.UpdateUser)
	r.Delete("/users/{id}", h.DeleteUser)
}

// Register handles user registration.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// Login handles user login.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	var resp []map[string]interface{}
	for _, u := range users {
		resp = append(resp, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"role":     u.Role,
		})
	}
	json.NewEncoder(w).Encode(resp)
}

// GetUserByID handles GET /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.service.GetUser(id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		h.respondError(w, http.StatusNotFound, "user not found")
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.service.GetUser(id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		h.respondError(w, http.StatusNotFound, "user not found")
		return
	}
	user.Username = req.Username
	user.Email = req.Email
	user.Role = req.Role
	if err := h.service.UpdateUser(user); err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.service.DeleteUser(id); err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) respondError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
