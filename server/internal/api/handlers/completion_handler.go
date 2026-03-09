package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
)

type CompletionHandler struct {
	repo *repository.HabitRepository
}

func NewCompletionHandler(repo *repository.HabitRepository) *CompletionHandler {
	return &CompletionHandler{repo: repo}
}

func (h *CompletionHandler) GetCompletions(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date query param required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	completions, err := h.repo.GetCompletionsForDate(r.Context(), authUser.ID, date)
	if err != nil {
		http.Error(w, "failed to get completions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, completions)
}

func (h *CompletionHandler) GetCompletionsRange(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" || end == "" {
		http.Error(w, "start and end query params required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	completions, err := h.repo.GetCompletionsForRange(r.Context(), authUser.ID, start, end)
	if err != nil {
		http.Error(w, "failed to get completions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, completions)
}

func (h *CompletionHandler) ToggleCompletion(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		HabitID string `json:"habitId"`
		Date    string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.HabitID == "" || body.Date == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.ToggleCompletion(r.Context(), authUser.ID, body.HabitID, body.Date); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "habit not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to toggle completion", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
