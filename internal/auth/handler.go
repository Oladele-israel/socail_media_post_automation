// internal/auth/handler.go
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Oladele-israel/socialmedia-post-automation/internal/middleware"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Router(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/logout", h.Logout)
		r.Get("/me", h.Me)
	})

	return r
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput // ← from dto.go, already has validate tags

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Replaces ALL manual if checks — validator reads the struct tags
	if errors := validator.Validate(input); errors != nil {
		respondValidationError(w, errors)
		return
	}

	result, err := h.service.Register(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput // ← from dto.go

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errors := validator.Validate(input); errors != nil {
		respondValidationError(w, errors)
		return
	}

	result, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var input RefreshInput // ← from dto.go

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errors := validator.Validate(input); errors != nil {
		respondValidationError(w, errors)
		return
	}

	result, err := h.service.RefreshTokens(r.Context(), input.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var input LogoutInput // ← from dto.go

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errors := validator.Validate(input); errors != nil {
		respondValidationError(w, errors)
		return
	}

	tokenID := middleware.GetTokenID(r)
	tokenExpiry := middleware.GetTokenExpiry(r)
	remaining := time.Until(tokenExpiry)

	if err := h.service.Logout(r.Context(), input.RefreshToken, tokenID, remaining); err != nil {
		respondError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	respond(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	user, err := h.service.repo.GetUserById(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respond(w, http.StatusOK, user)
}

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

func respondValidationError(w http.ResponseWriter, errors interface{}) {
	respond(w, http.StatusUnprocessableEntity, map[string]interface{}{
		"error":  "validation failed",
		"fields": errors,
	})
}
