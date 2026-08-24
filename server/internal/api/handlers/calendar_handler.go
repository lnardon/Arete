package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
)

type CalendarEventHandler struct {
	repo *repository.CalendarEventRepository
}

func NewCalendarEventHandler(repo *repository.CalendarEventRepository) *CalendarEventHandler {
	return &CalendarEventHandler{repo: repo}
}

type calendarEventBody struct {
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	Location       *string `json:"location"`
	StartAt        string  `json:"startAt"`
	EndAt          string  `json:"endAt"`
	AllDay         bool    `json:"allDay"`
	Timezone       string  `json:"timezone"`
	RecurrenceRule *string `json:"recurrenceRule"`
	Color          string  `json:"color"`
}

func (h *CalendarEventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	startAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	if err != nil {
		http.Error(w, "start query param required (RFC3339)", http.StatusBadRequest)
		return
	}
	endAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	if err != nil {
		http.Error(w, "end query param required (RFC3339)", http.StatusBadRequest)
		return
	}

	events, err := h.repo.ListEventsInRange(r.Context(), authUser.ID, startAt, endAt)
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *CalendarEventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	event, err := h.repo.GetEvent(r.Context(), authUser.ID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get event", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func (h *CalendarEventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	body, startAt, endAt, ok := decodeCalendarEventBody(w, r)
	if !ok {
		return
	}

	event, err := h.repo.CreateEvent(r.Context(), authUser.ID, body.Title, body.Description, body.Location, startAt, endAt, body.AllDay, body.Timezone, body.RecurrenceRule, body.Color)
	if err != nil {
		http.Error(w, "failed to create event", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (h *CalendarEventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	body, startAt, endAt, ok := decodeCalendarEventBody(w, r)
	if !ok {
		return
	}

	id := mux.Vars(r)["id"]
	event, err := h.repo.UpdateEvent(r.Context(), authUser.ID, id, body.Title, body.Description, body.Location, startAt, endAt, body.AllDay, body.Timezone, body.RecurrenceRule, body.Color)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update event", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func (h *CalendarEventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if err := h.repo.DeleteEvent(r.Context(), authUser.ID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeCalendarEventBody parses and validates the shared create/update
// request shape, writing an error response and returning ok=false on
// failure so callers can just return.
func decodeCalendarEventBody(w http.ResponseWriter, r *http.Request) (calendarEventBody, time.Time, time.Time, bool) {
	var body calendarEventBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}
	if len(body.Title) > 255 {
		http.Error(w, "title must be 255 characters or fewer", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}

	startAt, err := time.Parse(time.RFC3339, body.StartAt)
	if err != nil {
		http.Error(w, "startAt must be an RFC3339 timestamp", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}
	endAt, err := time.Parse(time.RFC3339, body.EndAt)
	if err != nil {
		http.Error(w, "endAt must be an RFC3339 timestamp", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}
	if !endAt.After(startAt) {
		http.Error(w, "endAt must be after startAt", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}

	if body.Timezone == "" {
		http.Error(w, "timezone is required (IANA name)", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}
	if _, err := time.LoadLocation(body.Timezone); err != nil {
		http.Error(w, "timezone must be a valid IANA timezone name", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}

	if body.Color == "" {
		body.Color = "#6366f1"
	}
	if !hexColorRegex.MatchString(body.Color) {
		http.Error(w, "color must be a hex value like #6366f1", http.StatusBadRequest)
		return body, time.Time{}, time.Time{}, false
	}

	return body, startAt, endAt, true
}
