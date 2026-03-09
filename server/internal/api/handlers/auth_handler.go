package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Preventing timing-based username enumeration.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)

type AuthHandler struct {
	userRepo *repository.UserRepository
	authSvc  *auth.Service
}

func NewAuthHandler(userRepo *repository.UserRepository, authSvc *auth.Service) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, authSvc: authSvc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.CreateUser(r.Context(), body.Username, string(hash))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.authSvc.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.authSvc.SetCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]string{
		"userId":   user.ID,
		"username": user.Username,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByUsername(r.Context(), body.Username)
	notFound := err == repository.ErrNotFound

	hashToCompare := []byte(user.PasswordHash)
	if notFound {
		// Use dummy hash to consume constant time and prevent timing attacks.
		hashToCompare = dummyHash
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if bcrypt.CompareHashAndPassword(hashToCompare, []byte(body.Password)) != nil || notFound {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.authSvc.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.authSvc.SetCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{
		"userId":   user.ID,
		"username": user.Username,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authSvc.ClearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"userId":   authUser.ID,
		"username": authUser.Username,
	})
}
