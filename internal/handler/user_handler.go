package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/internal/middleware"
	"github.com/melihgurlek/backend-path/pkg"
)

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateRequest represents the request body for user updates.
type UpdateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	service   domain.UserService
	jwtSecret string
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service domain.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{service: service, jwtSecret: jwtSecret}
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
	req, ok := middleware.GetValidatedBody[*RegisterRequest](r.Context())
	if !ok {
		panic("could not retrieve validated body")
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
	req, ok := middleware.GetValidatedBody[*LoginRequest](r.Context())
	if !ok {
		panic("could not retrieve validated body")
	}

	user, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Generate JWT token
	token, err := pkg.GenerateToken(h.jwtSecret, strconv.Itoa(user.ID), user.Role)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"token":    token,
	})
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	if claims.Role != "admin" {
		h.respondError(w, http.StatusForbidden, "you do not have permission to list users")
		return
	}

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
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	targetIDStr := chi.URLParam(r, "id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Use IsAdminOrSelf for authorization
	if !middleware.IsAdminOrSelf(claims, targetID) {
		h.respondError(w, http.StatusForbidden, "you do not have permission to view this user")
		return
	}

	user, err := h.service.GetUser(targetID) // Use targetID
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
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}
	targetIDStr := chi.URLParam(r, "id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Use IsAdminOrSelf for authorization
	if !middleware.IsAdminOrSelf(claims, targetID) {
		h.respondError(w, http.StatusForbidden, "you do not have permission to update this user")
		return
	}

	// --- 4. Original Logic (with Privilege Escalation fix) ---
	req, ok := middleware.GetValidatedBody[*UpdateRequest](r.Context())
	if !ok {
		panic("could not retrieve validated body")
	}

	user, err := h.service.GetUser(targetID)
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

	// **SECURITY FIX**: Prevents a regular user from making themselves an admin.
	// Only an existing admin can change a user's role.
	if claims.Role == "admin" && req.Role != "" {
		user.Role = req.Role
	}

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
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}
	targetIDStr := chi.URLParam(r, "id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Use IsAdminOrSelf for authorization
	if !middleware.IsAdminOrSelf(claims, targetID) {
		h.respondError(w, http.StatusForbidden, "you do not have permission to delete this user")
		return
	}
	// --- Original Logic ---
	if err := h.service.DeleteUser(targetID); err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) respondError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
