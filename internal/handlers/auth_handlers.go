package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ticket-system/internal/auth"
	"ticket-system/internal/store"
)

// Handlers holds shared dependencies (right now, just the store) so every
// handler method has access to storage without global variables.
type Handlers struct {
	Store *store.Store
}

// New wires up a Handlers with the given store.
func New(s *store.Store) *Handlers {
	return &Handlers{Store: s}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// Register handles POST /auth/register.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user, err := h.Store.CreateUser(req.Email, hash)
	if err != nil {
		if err == store.ErrDuplicateEmail {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{ID: user.ID, Email: user.Email})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login handles POST /auth/login.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := h.Store.GetUserByEmail(req.Email)
	if err != nil {
		// Same message whether the email doesn't exist or the password is
		// wrong — don't reveal which one it was.
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}
