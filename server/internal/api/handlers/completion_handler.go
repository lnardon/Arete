package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lnardon/arete/internal/repository"
)

type CompletionHandler struct {
	repo *repository.HabitRepository
}

func NewCompletionHandler(repo *repository.HabitRepository) *CompletionHandler {
	return &CompletionHandler{repo: repo}
}

func (h *CompletionHandler) GetCompletions(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date query param required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	
	completions, err := h.repo.GetCompletionsForDate(r.Context(), date)
	if err != nil {
		http.Error(w, "failed to get completions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, completions)
}

func (h *CompletionHandler) GetCompletionsRange(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" || end == "" {
		http.Error(w, "start and end query params required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	completions, err := h.repo.GetCompletionsForRange(r.Context(), start, end)
	if err != nil {
		http.Error(w, "failed to get completions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, completions)
}

func (h *CompletionHandler) ToggleCompletion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		HabitID string `json:"habitId"`
		Date    string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.HabitID == "" || body.Date == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.ToggleCompletion(r.Context(), body.HabitID, body.Date); err != nil {
		http.Error(w, "failed to toggle completion", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
