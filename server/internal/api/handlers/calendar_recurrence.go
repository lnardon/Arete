package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/lnardon/arete/internal/models"
	"github.com/lnardon/arete/internal/recurrence"
	"github.com/lnardon/arete/internal/repository"
)

// This file holds the scoped edits and deletes of a recurring series'
// occurrences. The model is Google Calendar's: a series row with RFC 5545
// lines, plus exception rows that edit or cancel single occurrences.

// edits records which text fields an occurrence edit changed, compared with
// the occurrence as the user saw it. Only those reach the rest of the series:
// an "all events" title change renames occurrences edited one by one, while
// fields the user left alone keep each occurrence's own value.
type edits struct {
	title, description, location, color bool
}

func diffEdits(base models.CalendarEvent, in eventInput) edits {
	return edits{
		title:       base.Title != in.Title,
		description: !sameText(base.Description, in.Description),
		location:    !sameText(base.Location, in.Location),
		color:       base.Color != in.Color,
	}
}

func sameText(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (e edits) apply(f *repository.CalendarEventFields, in eventInput) {
	if e.title {
		f.Title = in.Title
	}
	if e.description {
		f.Description = in.Description
	}
	if e.location {
		f.Location = in.Location
	}
	if e.color {
		f.Color = in.Color
	}
}

func (e edits) patch(in eventInput) repository.ExceptionPatch {
	return repository.ExceptionPatch{
		SetTitle: e.title, Title: in.Title,
		SetDescription: e.description, Description: in.Description,
		SetLocation: e.location, Location: in.Location,
		SetColor: e.color, Color: in.Color,
	}
}

func loadLocation(name string) *time.Location {
	if loc, err := time.LoadLocation(name); err == nil {
		return loc
	}
	return time.UTC
}

// reshape works out the series an occurrence edit produces when the edit
// applies from that occurrence onward. base is the occurrence as the user saw
// it, in is what they saved, and anchor is the slot the reshaped series
// starts from (the series' start for "all", the edited slot for
// "following"). Every slot moves the way the edited occurrence moved, and the
// occurrence length changes by as much as the edit changed it.
func reshape(series, base models.CalendarEvent, in eventInput, anchor time.Time) (repository.CalendarEventFields, recurrence.Shift, edits) {
	oldLoc := loadLocation(series.Timezone)
	shift := recurrence.NewShift(base.StartAt, oldLoc, in.startAt, in.loc)
	changed := diffEdits(base, in)

	f := repository.FieldsOf(series)
	changed.apply(&f, in)
	f.AllDay, f.Timezone, f.Recurrence = in.AllDay, in.Timezone, in.Recurrence
	f.StartAt = shift.Apply(anchor)

	switch {
	case in.AllDay != series.AllDay:
		f.EndAt = recurrence.EndFrom(f.StartAt, in.startAt, in.endAt, in.AllDay, in.loc)
	case in.AllDay:
		days := recurrence.DayCount(series.StartAt, series.EndAt) +
			recurrence.DayCount(in.startAt, in.endAt) - recurrence.DayCount(base.StartAt, base.EndAt)
		if days < 1 {
			days = recurrence.DayCount(in.startAt, in.endAt)
		}
		f.EndAt = f.StartAt.In(in.loc).AddDate(0, 0, days)
	default:
		length := series.EndAt.Sub(series.StartAt) + in.endAt.Sub(in.startAt) - base.EndAt.Sub(base.StartAt)
		if length <= 0 {
			length = in.endAt.Sub(in.startAt)
		}
		f.EndAt = f.StartAt.Add(length)
	}
	return f, shift, changed
}

// carryExceptions moves a series' edited and cancelled occurrences onto its
// reshaped version, so a "this and following" or time-changing "all events"
// edit doesn't bring back deleted occurrences or drop edited ones. Each slot
// moves with the shift; occurrences the user had moved keep their own times,
// the rest follow the new series' times. Exceptions whose slot isn't an
// occurrence of the new series are dropped. editedSlot is skipped: that
// occurrence is rebuilt from the edit itself (see editedOccurrence). from, when
// set, limits the carry to exceptions at or after it.
func carryExceptions(oldSeries, newSeries models.CalendarEvent, exceptions []models.CalendarEvent, shift recurrence.Shift, changed edits, in eventInput, editedSlot time.Time, from *time.Time) []repository.ExceptionWrite {
	var out []repository.ExceptionWrite
	taken := map[int64]bool{}
	for _, x := range exceptions {
		original := *x.OriginalStartAt
		if original.Equal(editedSlot) || (from != nil && original.Before(*from)) {
			continue
		}
		slot := shift.Apply(original)
		if taken[slot.Unix()] || !recurrence.IsOccurrence(newSeries, slot) {
			continue
		}
		taken[slot.Unix()] = true

		f := repository.FieldsOf(x)
		f.Recurrence, f.RecurrenceEndAt = nil, nil
		moved := !x.StartAt.Equal(original) ||
			!x.EndAt.Equal(recurrence.OccurrenceEnd(oldSeries, original)) ||
			x.AllDay != oldSeries.AllDay
		if x.Status == models.CalendarEventCancelled || !moved {
			f.StartAt, f.EndAt = slot, recurrence.OccurrenceEnd(newSeries, slot)
			f.AllDay, f.Timezone = newSeries.AllDay, newSeries.Timezone
		}
		if x.Status == models.CalendarEventConfirmed {
			changed.apply(&f, in)
		}
		out = append(out, repository.ExceptionWrite{OriginalStartAt: slot, Fields: f, Status: x.Status, Source: x.Source})
	}
	return out
}

// editedOccurrence returns the exception the edited occurrence needs on the
// reshaped series, if the series alone doesn't already show it as saved — for
// example an occurrence that had been moved and keeps its own time.
func editedOccurrence(newSeries models.CalendarEvent, shift recurrence.Shift, base models.CalendarEvent, in eventInput) []repository.ExceptionWrite {
	slot := shift.Apply(*base.OriginalStartAt)
	matchesSeries := in.startAt.Equal(slot) &&
		in.endAt.Equal(recurrence.OccurrenceEnd(newSeries, slot)) &&
		in.AllDay == newSeries.AllDay &&
		in.Title == newSeries.Title &&
		sameText(in.Description, newSeries.Description) &&
		sameText(in.Location, newSeries.Location) &&
		in.Color == newSeries.Color
	if matchesSeries || !recurrence.IsOccurrence(newSeries, slot) {
		return nil
	}

	f := in.fields()
	f.Recurrence = nil
	return []repository.ExceptionWrite{{OriginalStartAt: slot, Fields: f, Status: models.CalendarEventConfirmed, Source: "app"}}
}

// timingChanged reports whether an edit moves the series' occurrence slots,
// which is what invalidates the original start of every exception.
func timingChanged(series models.CalendarEvent, next repository.CalendarEventFields) bool {
	return !next.StartAt.Equal(series.StartAt) ||
		!next.EndAt.Equal(series.EndAt) ||
		next.AllDay != series.AllDay ||
		next.Timezone != series.Timezone ||
		!recurrence.SameLines(next.Recurrence, series.Recurrence)
}

// editThis edits one occurrence: an exception replaces it.
func (h *CalendarEventHandler) editThis(w http.ResponseWriter, r *http.Request, userID string, ref eventRef, in eventInput) {
	if len(in.Recurrence) > 0 {
		http.Error(w, "recurrence must be null when editing a single occurrence", http.StatusBadRequest)
		return
	}

	exception, err := h.repo.UpsertException(r.Context(), userID, ref.event.ID, *ref.slot, in.fields(), models.CalendarEventConfirmed)
	if err != nil {
		http.Error(w, "failed to update event", http.StatusInternalServerError)
		return
	}
	h.syncSeries(r.Context(), userID, ref.event, nil)
	writeJSON(w, http.StatusOK, recurrence.ExceptionOccurrence(ref.event, exception))
}

// editAll applies an occurrence edit to the whole series.
func (h *CalendarEventHandler) editAll(w http.ResponseWriter, r *http.Request, userID string, series, base models.CalendarEvent, in eventInput) {
	next, shift, changed := reshape(series, base, in, series.StartAt)
	if len(next.Recurrence) == 0 {
		// "Does not repeat": the series collapses into one event at the
		// occurrence the user edited.
		next.StartAt, next.EndAt = in.startAt, in.endAt
	}
	next, err := withRecurrenceEnd(next)
	if err != nil {
		http.Error(w, "invalid recurrence", http.StatusBadRequest)
		return
	}

	rw := repository.SeriesRewrite{SeriesID: series.ID, Series: next}
	if !timingChanged(series, next) {
		rw.Patch = changed.patch(in)
	} else {
		exceptions, err := h.repo.ListExceptions(r.Context(), userID, series.ID)
		if err != nil {
			http.Error(w, "failed to update event", http.StatusInternalServerError)
			return
		}
		rw.ReplaceExceptions = true
		if len(next.Recurrence) > 0 {
			nextSeries := next.Event()
			rw.Exceptions = append(
				carryExceptions(series, nextSeries, exceptions, shift, changed, in, *base.OriginalStartAt, nil),
				editedOccurrence(nextSeries, shift, base, in)...,
			)
		}
	}

	updated, _, removed, err := h.repo.RewriteSeries(r.Context(), userID, rw)
	if err != nil {
		writeRewriteError(w, err)
		return
	}
	h.syncSeries(r.Context(), userID, updated, removed)
	writeJSON(w, http.StatusOK, updated)
}

// editFollowing splits the series at the edited occurrence: the original
// series ends just before it, and a new series carries the edit from there.
func (h *CalendarEventHandler) editFollowing(w http.ResponseWriter, r *http.Request, userID string, ref eventRef, in eventInput) {
	series, slot, base := ref.event, *ref.slot, ref.occurrence()
	if slot.Equal(series.StartAt) {
		h.editAll(w, r, userID, series, base, in)
		return
	}

	old, kept, ok := trimmedSeries(w, series, slot)
	if !ok {
		return
	}

	next, shift, changed := reshape(series, base, in, slot)
	if recurrence.SameLines(in.Recurrence, series.Recurrence) {
		// The rule carried over unchanged, so occurrences already spent on
		// the original series no longer count toward the new one's COUNT.
		next.Recurrence = recurrence.ReduceCount(next.Recurrence, kept)
	}
	if len(next.Recurrence) == 0 {
		next.StartAt, next.EndAt = in.startAt, in.endAt
	}
	next, err := withRecurrenceEnd(next)
	if err != nil {
		http.Error(w, "invalid recurrence", http.StatusBadRequest)
		return
	}

	rw := repository.SeriesRewrite{
		SeriesID:          series.ID,
		Series:            old,
		Create:            &next,
		ReplaceExceptions: true,
		ReplaceFrom:       &slot,
	}
	if len(next.Recurrence) > 0 {
		exceptions, err := h.repo.ListExceptions(r.Context(), userID, series.ID)
		if err != nil {
			http.Error(w, "failed to update event", http.StatusInternalServerError)
			return
		}
		nextSeries := next.Event()
		rw.Exceptions = append(
			carryExceptions(series, nextSeries, exceptions, shift, changed, in, slot, &slot),
			editedOccurrence(nextSeries, shift, base, in)...,
		)
	}

	updated, created, removed, err := h.repo.RewriteSeries(r.Context(), userID, rw)
	if err != nil {
		writeRewriteError(w, err)
		return
	}
	h.syncSeries(r.Context(), userID, updated, removed)
	writeJSON(w, http.StatusOK, created)
}

// cancelOccurrence deletes one occurrence: a cancelled exception hides it,
// the same way Google records a deleted instance.
func (h *CalendarEventHandler) cancelOccurrence(w http.ResponseWriter, r *http.Request, userID string, ref eventRef) {
	fields := repository.FieldsOf(recurrence.Occurrence(ref.event, *ref.slot))
	fields.Recurrence, fields.RecurrenceEndAt = nil, nil

	if _, err := h.repo.UpsertException(r.Context(), userID, ref.event.ID, *ref.slot, fields, models.CalendarEventCancelled); err != nil {
		http.Error(w, "failed to delete event", http.StatusInternalServerError)
		return
	}
	h.syncSeries(r.Context(), userID, ref.event, nil)
	w.WriteHeader(http.StatusNoContent)
}

// deleteFollowing ends the series just before the occurrence.
func (h *CalendarEventHandler) deleteFollowing(w http.ResponseWriter, r *http.Request, userID string, ref eventRef) {
	old, _, ok := trimmedSeries(w, ref.event, *ref.slot)
	if !ok {
		return
	}

	updated, _, removed, err := h.repo.RewriteSeries(r.Context(), userID, repository.SeriesRewrite{
		SeriesID:          ref.event.ID,
		Series:            old,
		ReplaceExceptions: true,
		ReplaceFrom:       ref.slot,
	})
	if err != nil {
		writeRewriteError(w, err)
		return
	}
	h.syncSeries(r.Context(), userID, updated, removed)
	w.WriteHeader(http.StatusNoContent)
}

// trimmedSeries returns the series' fields ended just before slot, and how
// many occurrences it keeps, writing a 500 and returning ok=false on failure.
func trimmedSeries(w http.ResponseWriter, series models.CalendarEvent, slot time.Time) (repository.CalendarEventFields, int, bool) {
	lines, kept, err := recurrence.TrimBefore(series, slot)
	if err != nil {
		http.Error(w, "failed to split recurring event", http.StatusInternalServerError)
		return repository.CalendarEventFields{}, 0, false
	}
	old := repository.FieldsOf(series)
	old.Recurrence = lines
	if old, err = withRecurrenceEnd(old); err != nil {
		http.Error(w, "failed to split recurring event", http.StatusInternalServerError)
		return old, 0, false
	}
	return old, kept, true
}

func writeRewriteError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	slog.Error("calendar: rewrite recurring event failed", "error", err)
	http.Error(w, "failed to update event", http.StatusInternalServerError)
}

// syncSeries pushes an edit to a Google-linked series right away, then
// cancels the Google occurrences of exceptions the edit removed. A regular
// tick pulls before it pushes, so waiting for it could let Google's stale copy
// of the series land in between. Failures are logged and the rows stay dirty
// for the next tick, as with DeleteEvent's synchronous Google delete.
func (h *CalendarEventHandler) syncSeries(ctx context.Context, userID string, series models.CalendarEvent, removedGoogleIDs []string) {
	if series.GoogleEventID == nil && len(removedGoogleIDs) == 0 {
		return
	}
	if err := h.google.PushUser(ctx, userID); err != nil {
		slog.Error("calendar: immediate google push failed, leaving it for the next sync", "eventId", series.ID, "error", err)
	}
	for _, googleID := range removedGoogleIDs {
		if err := h.google.DeleteRemoteEvent(ctx, userID, googleID); err != nil {
			slog.Error("calendar: cancel removed google occurrence failed", "eventId", series.ID, "googleEventId", googleID, "error", err)
		}
	}
}
