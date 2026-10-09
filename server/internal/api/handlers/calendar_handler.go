package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/models"
	"github.com/lnardon/arete/internal/recurrence"
	"github.com/lnardon/arete/internal/repository"
)

// Edit and delete scopes for an occurrence of a recurring event, matching
// Google Calendar's "This event / This and following events / All events".
const (
	scopeThis      = "this"
	scopeFollowing = "following"
	scopeAll       = "all"
)

// maxListRange bounds GET /calendar/events, so expanding series that never
// end stays cheap. The day and week views ask for 1 or 7 days.
const maxListRange = 92 * 24 * time.Hour

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// googleSync is the slice of *google.SyncEngine this handler needs. Declared
// here (rather than importing internal/google directly) so this package
// doesn't need to depend on the Google integration to handle plain, unsynced
// events.
type googleSync interface {
	DeleteRemoteEvent(ctx context.Context, userID, googleEventID string) error
	PushUser(ctx context.Context, userID string) error
}

type CalendarEventHandler struct {
	repo   *repository.CalendarEventRepository
	google googleSync
}

func NewCalendarEventHandler(repo *repository.CalendarEventRepository, google googleSync) *CalendarEventHandler {
	return &CalendarEventHandler{repo: repo, google: google}
}

type calendarEventBody struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Location    *string  `json:"location"`
	StartAt     string   `json:"startAt"`
	EndAt       string   `json:"endAt"`
	AllDay      bool     `json:"allDay"`
	Timezone    string   `json:"timezone"`
	Recurrence  []string `json:"recurrence"`
	Color       string   `json:"color"`
}

// eventInput is a validated create/update body.
type eventInput struct {
	calendarEventBody
	startAt, endAt time.Time
	loc            *time.Location
}

func (in eventInput) fields() repository.CalendarEventFields {
	return repository.CalendarEventFields{
		Title:       in.Title,
		Description: in.Description,
		Location:    in.Location,
		StartAt:     in.startAt,
		EndAt:       in.endAt,
		AllDay:      in.AllDay,
		Timezone:    in.Timezone,
		Recurrence:  in.Recurrence,
		Color:       in.Color,
	}
}

// withRecurrenceEnd stores when a series' last occurrence ends, which the
// range query needs to skip finished series.
func withRecurrenceEnd(f repository.CalendarEventFields) (repository.CalendarEventFields, error) {
	f.RecurrenceEndAt = nil
	if len(f.Recurrence) == 0 {
		f.Recurrence = nil
		return f, nil
	}
	end, err := recurrence.End(f.Event())
	if err != nil {
		return f, err
	}
	f.RecurrenceEndAt = end
	return f, nil
}

// eventRef is what an {id} path parameter points at: a single event or a
// series, or — for an instance ID — one occurrence of a series.
type eventRef struct {
	event     models.CalendarEvent  // the single event or series
	slot      *time.Time            // the occurrence's original start
	exception *models.CalendarEvent // the exception replacing it, if any
}

// occurrence presents the referenced occurrence as the list endpoint does.
func (ref eventRef) occurrence() models.CalendarEvent {
	if ref.exception != nil {
		return recurrence.ExceptionOccurrence(ref.event, *ref.exception)
	}
	return recurrence.Occurrence(ref.event, *ref.slot)
}

func (h *CalendarEventHandler) resolveEventRef(ctx context.Context, userID, id string) (eventRef, error) {
	seriesID, stamp, isInstance := recurrence.ParseInstanceID(id)
	if !isInstance {
		if !uuidRegex.MatchString(id) {
			return eventRef{}, repository.ErrNotFound
		}
		event, err := h.repo.GetEvent(ctx, userID, id)
		return eventRef{event: event}, err
	}

	series, err := h.repo.GetEvent(ctx, userID, seriesID)
	if err != nil {
		return eventRef{}, err
	}
	if len(series.Recurrence) == 0 {
		return eventRef{}, repository.ErrNotFound
	}
	slot, err := recurrence.ParseStamp(series, stamp)
	if err != nil {
		return eventRef{}, repository.ErrNotFound
	}

	ref := eventRef{event: series, slot: &slot}
	exception, err := h.repo.GetException(ctx, userID, seriesID, slot)
	switch {
	case err == nil:
		ref.exception = &exception
	case errors.Is(err, repository.ErrNotFound):
		if !recurrence.IsOccurrence(series, slot) {
			return eventRef{}, repository.ErrNotFound
		}
	default:
		return eventRef{}, err
	}
	return ref, nil
}

// lookup resolves the {id} path parameter, writing a 404 or 500 and returning
// ok=false on failure so callers can just return.
func (h *CalendarEventHandler) lookup(w http.ResponseWriter, r *http.Request, userID string) (eventRef, bool) {
	ref, err := h.resolveEventRef(r.Context(), userID, mux.Vars(r)["id"])
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return ref, false
		}
		http.Error(w, "failed to get event", http.StatusInternalServerError)
		return ref, false
	}
	return ref, true
}

// ListEvents returns everything the calendar shows in [start, end): single
// events, the occurrences of recurring series, and edited occurrences.
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
	if !endAt.After(startAt) {
		http.Error(w, "end must be after start", http.StatusBadRequest)
		return
	}
	if endAt.Sub(startAt) > maxListRange {
		http.Error(w, "range must be 92 days or less", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	singles, err := h.repo.ListSinglesInRange(ctx, authUser.ID, startAt, endAt)
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	series, err := h.repo.ListSeriesInRange(ctx, authUser.ID, startAt, endAt)
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	exceptions, err := h.repo.ListExceptionsInRange(ctx, authUser.ID, startAt, endAt)
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}

	// An occurrence moved into the range can belong to a series that doesn't
	// otherwise reach it, so load every series an exception points at.
	known := map[string]bool{}
	for _, s := range series {
		known[s.ID] = true
	}
	for _, x := range exceptions {
		id := *x.RecurringEventID
		if known[id] {
			continue
		}
		known[id] = true
		s, err := h.repo.GetEvent(ctx, authUser.ID, id)
		if err != nil {
			http.Error(w, "failed to list events", http.StatusInternalServerError)
			return
		}
		series = append(series, s)
	}

	writeJSON(w, http.StatusOK, recurrence.Occurrences(singles, series, exceptions, startAt, endAt))
}

func (h *CalendarEventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	ref, ok := h.lookup(w, r, authUser.ID)
	if !ok {
		return
	}
	if ref.slot == nil {
		writeJSON(w, http.StatusOK, ref.event)
		return
	}
	if ref.exception != nil && ref.exception.Status == models.CalendarEventCancelled {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, ref.occurrence())
}

func (h *CalendarEventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	in, ok := decodeCalendarEventBody(w, r)
	if !ok {
		return
	}
	fields, err := withRecurrenceEnd(in.fields())
	if err != nil {
		http.Error(w, "invalid recurrence", http.StatusBadRequest)
		return
	}

	event, err := h.repo.CreateEvent(r.Context(), authUser.ID, fields)
	if err != nil {
		http.Error(w, "failed to create event", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

// UpdateEvent edits a single event or series by its UUID, or one occurrence
// of a series by its instance ID together with ?scope=this|following|all.
func (h *CalendarEventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	in, ok := decodeCalendarEventBody(w, r)
	if !ok {
		return
	}
	ref, ok := h.lookup(w, r, authUser.ID)
	if !ok {
		return
	}
	if ref.event.RecurringEventID != nil {
		http.Error(w, "address an edited occurrence by its instance id", http.StatusBadRequest)
		return
	}

	scope := r.URL.Query().Get("scope")
	if ref.slot == nil {
		if scope != "" && scope != scopeAll {
			http.Error(w, "scope only applies to an occurrence of a recurring event", http.StatusBadRequest)
			return
		}
		if len(ref.event.Recurrence) > 0 {
			// The series row itself: an "all events" edit made from its
			// first occurrence.
			h.editAll(w, r, authUser.ID, ref.event, recurrence.Occurrence(ref.event, ref.event.StartAt), in)
			return
		}
		h.updateSingle(w, r, authUser.ID, ref.event.ID, in)
		return
	}

	switch scope {
	case scopeThis:
		h.editThis(w, r, authUser.ID, ref, in)
	case scopeFollowing:
		h.editFollowing(w, r, authUser.ID, ref, in)
	case scopeAll:
		h.editAll(w, r, authUser.ID, ref.event, ref.occurrence(), in)
	default:
		http.Error(w, "scope must be this, following or all", http.StatusBadRequest)
	}
}

func (h *CalendarEventHandler) updateSingle(w http.ResponseWriter, r *http.Request, userID, id string, in eventInput) {
	fields, err := withRecurrenceEnd(in.fields())
	if err != nil {
		http.Error(w, "invalid recurrence", http.StatusBadRequest)
		return
	}

	event, err := h.repo.UpdateEvent(r.Context(), userID, id, fields)
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

// DeleteEvent deletes a single event or series by its UUID, or occurrences of
// a series by an instance ID together with ?scope=this|following|all.
func (h *CalendarEventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	ref, ok := h.lookup(w, r, authUser.ID)
	if !ok {
		return
	}
	if ref.event.RecurringEventID != nil {
		http.Error(w, "address an edited occurrence by its instance id", http.StatusBadRequest)
		return
	}

	scope := r.URL.Query().Get("scope")
	if ref.slot == nil {
		if scope != "" && scope != scopeAll {
			http.Error(w, "scope only applies to an occurrence of a recurring event", http.StatusBadRequest)
			return
		}
		h.deleteWhole(w, r, authUser.ID, ref.event)
		return
	}

	switch scope {
	case scopeThis:
		h.cancelOccurrence(w, r, authUser.ID, ref)
	case scopeFollowing:
		if ref.slot.Equal(ref.event.StartAt) {
			h.deleteWhole(w, r, authUser.ID, ref.event)
			return
		}
		h.deleteFollowing(w, r, authUser.ID, ref)
	case scopeAll:
		h.deleteWhole(w, r, authUser.ID, ref.event)
	default:
		http.Error(w, "scope must be this, following or all", http.StatusBadRequest)
	}
}

// deleteWhole deletes a single event or a whole series; a series' exceptions
// go with it via ON DELETE CASCADE, locally and on Google.
func (h *CalendarEventHandler) deleteWhole(w http.ResponseWriter, r *http.Request, userID string, event models.CalendarEvent) {
	// A Google-linked event must be deleted on Google's side synchronously,
	// not on the next sync tick — otherwise the next pull would upsert it
	// straight back into existence locally, since Google would still have it.
	if event.GoogleEventID != nil {
		if err := h.google.DeleteRemoteEvent(r.Context(), userID, *event.GoogleEventID); err != nil {
			slog.Error("calendar: remote google delete failed, deleting local copy anyway", "eventId", event.ID, "error", err)
		}
	}

	if err := h.repo.DeleteEvent(r.Context(), userID, event.ID); err != nil {
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
func decodeCalendarEventBody(w http.ResponseWriter, r *http.Request) (eventInput, bool) {
	var in eventInput
	if err := json.NewDecoder(r.Body).Decode(&in.calendarEventBody); err != nil || in.Title == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return in, false
	}
	if len(in.Title) > 255 {
		http.Error(w, "title must be 255 characters or fewer", http.StatusBadRequest)
		return in, false
	}

	var err error
	if in.startAt, err = time.Parse(time.RFC3339, in.StartAt); err != nil {
		http.Error(w, "startAt must be an RFC3339 timestamp", http.StatusBadRequest)
		return in, false
	}
	if in.endAt, err = time.Parse(time.RFC3339, in.EndAt); err != nil {
		http.Error(w, "endAt must be an RFC3339 timestamp", http.StatusBadRequest)
		return in, false
	}
	if !in.endAt.After(in.startAt) {
		http.Error(w, "endAt must be after startAt", http.StatusBadRequest)
		return in, false
	}

	if in.Timezone == "" {
		http.Error(w, "timezone is required (IANA name)", http.StatusBadRequest)
		return in, false
	}
	if in.loc, err = time.LoadLocation(in.Timezone); err != nil {
		http.Error(w, "timezone must be a valid IANA timezone name", http.StatusBadRequest)
		return in, false
	}

	if len(in.Recurrence) == 0 {
		in.Recurrence = nil
	}
	if err := recurrence.Validate(in.Recurrence, in.AllDay); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return in, false
	}

	if in.Color == "" {
		in.Color = "#6366f1"
	}
	if !hexColorRegex.MatchString(in.Color) {
		http.Error(w, "color must be a hex value like #6366f1", http.StatusBadRequest)
		return in, false
	}

	return in, true
}
