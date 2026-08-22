package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type PomodoroProjectHandler struct {
	repo *repository.PomodoroRepository
}

func NewPomodoroProjectHandler(repo *repository.PomodoroRepository) *PomodoroProjectHandler {
	return &PomodoroProjectHandler{repo: repo}
}

func (h *PomodoroProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	projects, err := h.repo.ListProjects(r.Context(), authUser.ID)
	if err != nil {
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *PomodoroProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(body.Name) > 255 {
		http.Error(w, "project name must be 255 characters or fewer", http.StatusBadRequest)
		return
	}
	if body.Color == "" {
		body.Color = "#6366f1"
	}
	if !hexColorRegex.MatchString(body.Color) {
		http.Error(w, "color must be a hex value like #6366f1", http.StatusBadRequest)
		return
	}

	project, err := h.repo.CreateProject(r.Context(), authUser.ID, body.Name, body.Color)
	if err != nil {
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *PomodoroProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(body.Name) > 255 {
		http.Error(w, "project name must be 255 characters or fewer", http.StatusBadRequest)
		return
	}
	if body.Color == "" {
		body.Color = "#6366f1"
	}
	if !hexColorRegex.MatchString(body.Color) {
		http.Error(w, "color must be a hex value like #6366f1", http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	project, err := h.repo.UpdateProject(r.Context(), authUser.ID, id, body.Name, body.Color)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update project", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *PomodoroProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if err := h.repo.DeleteProject(r.Context(), authUser.ID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type PomodoroTimerHandler struct {
	repo *repository.PomodoroRepository
}

func NewPomodoroTimerHandler(repo *repository.PomodoroRepository) *PomodoroTimerHandler {
	return &PomodoroTimerHandler{repo: repo}
}

func (h *PomodoroTimerHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	entry, err := h.repo.GetActiveEntry(r.Context(), authUser.ID)
	if err != nil {
		http.Error(w, "failed to get active timer", http.StatusInternalServerError)
		return
	}
	if entry == nil {
		writeJSON(w, http.StatusOK, map[string]bool{"active": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"active": true, "entry": entry})
}

func (h *PomodoroTimerHandler) Start(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var body struct {
		ProjectID      *string `json:"projectId"`
		PlannedMinutes int     `json:"plannedMinutes"`
		LocalDate      string  `json:"localDate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.PlannedMinutes <= 0 {
		http.Error(w, "plannedMinutes must be greater than 0", http.StatusBadRequest)
		return
	}
	if body.LocalDate == "" || !dateRegex.MatchString(body.LocalDate) {
		http.Error(w, "localDate is required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	entry, err := h.repo.StartEntry(r.Context(), authUser.ID, body.ProjectID, body.PlannedMinutes, body.LocalDate)
	if err != nil {
		if errors.Is(err, repository.ErrActiveEntryExists) {
			http.Error(w, "a timer is already running", http.StatusConflict)
			return
		}
		http.Error(w, "failed to start timer", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (h *PomodoroTimerHandler) Stop(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	entry, err := h.repo.StopActiveEntry(r.Context(), authUser.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "no active timer", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to stop timer", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *PomodoroTimerHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" || !dateRegex.MatchString(start) || end == "" || !dateRegex.MatchString(end) {
		http.Error(w, "start and end query params required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	var projectID *string
	if v := r.URL.Query().Get("projectId"); v != "" {
		projectID = &v
	}

	entries, err := h.repo.ListEntries(r.Context(), authUser.ID, start, end, projectID)
	if err != nil {
		http.Error(w, "failed to list entries", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *PomodoroTimerHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if err := h.repo.DeleteEntry(r.Context(), authUser.ID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete entry", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
