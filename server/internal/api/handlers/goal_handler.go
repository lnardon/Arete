package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
)

var validPeriodTypes = map[string]bool{
	"month":    true,
	"quarter":  true,
	"semester": true,
	"year":     true,
}

type GoalHandler struct {
	repo *repository.GoalRepository
}

func NewGoalHandler(repo *repository.GoalRepository) *GoalHandler {
	return &GoalHandler{repo: repo}
}

func (h *GoalHandler) ListGoals(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	periodType := r.URL.Query().Get("period_type")
	periodKey := r.URL.Query().Get("period_key")
	if !validPeriodTypes[periodType] || periodKey == "" {
		http.Error(w, "period_type and period_key are required", http.StatusBadRequest)
		return
	}

	goals, err := h.repo.ListGoals(r.Context(), authUser.ID, periodType, periodKey)
	if err != nil {
		http.Error(w, "failed to list goals", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, goals)
}

func (h *GoalHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		Title      string `json:"title"`
		PeriodType string `json:"periodType"`
		PeriodKey  string `json:"periodKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Title == "" || len(body.Title) > 255 {
		http.Error(w, "title must be between 1 and 255 characters", http.StatusBadRequest)
		return
	}
	if !validPeriodTypes[body.PeriodType] || body.PeriodKey == "" {
		http.Error(w, "invalid periodType or periodKey", http.StatusBadRequest)
		return
	}

	goal, err := h.repo.CreateGoal(r.Context(), authUser.ID, body.Title, body.PeriodType, body.PeriodKey)
	if err != nil {
		http.Error(w, "failed to create goal", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, goal)
}

func (h *GoalHandler) ToggleGoal(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	goal, err := h.repo.ToggleGoal(r.Context(), authUser.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "goal not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to toggle goal", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, goal)
}

func (h *GoalHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Title == "" || len(body.Title) > 255 {
		http.Error(w, "title must be between 1 and 255 characters", http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	goal, err := h.repo.UpdateGoal(r.Context(), authUser.ID, id, body.Title)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "goal not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update goal", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, goal)
}

func (h *GoalHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	err := h.repo.DeleteGoal(r.Context(), authUser.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "goal not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete goal", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
