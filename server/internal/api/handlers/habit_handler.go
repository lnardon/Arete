package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
)

type HabitHandler struct {
	repo *repository.HabitRepository
}

func NewHabitHandler(repo *repository.HabitRepository) *HabitHandler {
	return &HabitHandler{repo: repo}
}

func (h *HabitHandler) ListHabits(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	habits, err := h.repo.ListHabits(r.Context(), authUser.ID)
	if err != nil {
		http.Error(w, "failed to list habits", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, habits)
}

func (h *HabitHandler) CreateHabit(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	habit, err := h.repo.CreateHabit(r.Context(), authUser.ID, body.Name)
	if err != nil {
		http.Error(w, "failed to create habit", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, habit)
}

func (h *HabitHandler) UpdateHabit(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	habit, err := h.repo.UpdateHabit(r.Context(), authUser.ID, id, body.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "habit not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update habit", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, habit)
}

func (h *HabitHandler) DeleteHabit(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	err := h.repo.DeleteHabit(r.Context(), authUser.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "habit not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to delete habit", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
