package recurrence

import (
	"log/slog"
	"sort"
	"time"

	"github.com/lnardon/arete/internal/models"
)

func occurrence(series models.CalendarEvent, loc *time.Location, start time.Time) models.CalendarEvent {
	o := series
	seriesID, original := series.ID, start
	o.ID = instanceID(series.ID, start, series.AllDay, loc)
	o.StartAt = start
	o.EndAt = occurrenceEnd(series, loc, start)
	o.RecurringEventID = &seriesID
	o.OriginalStartAt = &original
	o.RecurrenceEndAt = nil
	return o
}

// Occurrence presents the plain (unedited) occurrence of series at start.
func Occurrence(series models.CalendarEvent, start time.Time) models.CalendarEvent {
	loc, err := time.LoadLocation(series.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return occurrence(series, loc, start)
}

func exceptionOccurrence(series models.CalendarEvent, loc *time.Location, exception models.CalendarEvent) models.CalendarEvent {
	o := exception
	o.ID = instanceID(series.ID, *exception.OriginalStartAt, series.AllDay, loc)
	o.Recurrence = series.Recurrence
	return o
}

// ExceptionOccurrence presents an exception row as the occurrence it
// replaces: addressed by the instance ID of its original slot, and carrying
// the series' rule so clients can show "repeats weekly".
func ExceptionOccurrence(series models.CalendarEvent, exception models.CalendarEvent) models.CalendarEvent {
	loc, err := time.LoadLocation(series.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return exceptionOccurrence(series, loc, exception)
}

// Occurrences merges the three row kinds into what a calendar shows for
// [from, to): single events, every series occurrence that no exception
// replaces, and confirmed exceptions whose own times overlap the range.
// series must include every series an exception points at. A series whose
// rule can't be expanded is logged and skipped rather than failing the view.
func Occurrences(singles, series, exceptions []models.CalendarEvent, from, to time.Time) []models.CalendarEvent {
	type seriesInfo struct {
		event models.CalendarEvent
		loc   *time.Location
	}
	byID := map[string]seriesInfo{}
	for _, s := range series {
		if len(s.Recurrence) == 0 {
			continue
		}
		loc, err := time.LoadLocation(s.Timezone)
		if err != nil {
			slog.Warn("calendar: skipping series with unknown timezone", "eventId", s.ID, "timezone", s.Timezone)
			continue
		}
		byID[s.ID] = seriesInfo{event: s, loc: loc}
	}

	replaced := map[string]map[int64]bool{}
	for _, x := range exceptions {
		if x.RecurringEventID == nil || x.OriginalStartAt == nil {
			continue
		}
		if replaced[*x.RecurringEventID] == nil {
			replaced[*x.RecurringEventID] = map[int64]bool{}
		}
		replaced[*x.RecurringEventID][x.OriginalStartAt.Unix()] = true
	}

	out := make([]models.CalendarEvent, 0, len(singles))
	out = append(out, singles...)

	for _, s := range byID {
		starts, err := expand(s.event, s.loc, from, to)
		if err != nil {
			slog.Warn("calendar: skipping series that failed to expand", "eventId", s.event.ID, "error", err)
			continue
		}
		for _, start := range starts {
			if !replaced[s.event.ID][start.Unix()] {
				out = append(out, occurrence(s.event, s.loc, start))
			}
		}
	}

	for _, x := range exceptions {
		if x.Status != models.CalendarEventConfirmed || x.RecurringEventID == nil || x.OriginalStartAt == nil {
			continue
		}
		if !x.StartAt.Before(to) || !x.EndAt.After(from) {
			continue
		}
		s, ok := byID[*x.RecurringEventID]
		if !ok {
			continue
		}
		out = append(out, exceptionOccurrence(s.event, s.loc, x))
	}

	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].StartAt.Equal(out[j].StartAt) {
			return out[i].StartAt.Before(out[j].StartAt)
		}
		return out[i].ID < out[j].ID
	})
	return out
}
