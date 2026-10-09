package recurrence

import (
	"testing"
	"time"

	"github.com/lnardon/arete/internal/models"
)

func exception(s models.CalendarEvent, original, start, end time.Time, status string) models.CalendarEvent {
	seriesID := s.ID
	return models.CalendarEvent{
		ID:               "0b5f0c1e-2d3a-4b5c-8d9e-1f2a3b4c5d6e",
		Title:            "Gym class (moved)",
		StartAt:          start,
		EndAt:            end,
		Timezone:         s.Timezone,
		RecurringEventID: &seriesID,
		OriginalStartAt:  &original,
		Status:           status,
		Color:            s.Color,
	}
}

func TestOccurrences(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	at := func(month time.Month, day, hour int) time.Time { return time.Date(2026, month, day, hour, 0, 0, 0, sp) }
	gym := series(at(10, 9, 18), at(10, 9, 19), "America/Sao_Paulo", false, "RRULE:FREQ=WEEKLY;BYDAY=FR")

	// The week of Monday Oct 19.
	from, to := at(10, 19, 0), at(10, 26, 0)

	tests := []struct {
		name       string
		exceptions []models.CalendarEvent
		wantIDs    []string
		wantStarts []time.Time
	}{
		{
			name:       "plain occurrence",
			wantIDs:    []string{gym.ID + "_20261023T210000Z"},
			wantStarts: []time.Time{at(10, 23, 18)},
		},
		{
			name:       "moved within the week keeps its instance ID",
			exceptions: []models.CalendarEvent{exception(gym, at(10, 23, 18), at(10, 24, 19), at(10, 24, 20), models.CalendarEventConfirmed)},
			wantIDs:    []string{gym.ID + "_20261023T210000Z"},
			wantStarts: []time.Time{at(10, 24, 19)},
		},
		{
			name:       "cancelled",
			exceptions: []models.CalendarEvent{exception(gym, at(10, 23, 18), at(10, 23, 18), at(10, 23, 19), models.CalendarEventCancelled)},
		},
		{
			name:       "moved out of the week",
			exceptions: []models.CalendarEvent{exception(gym, at(10, 23, 18), at(10, 27, 18), at(10, 27, 19), models.CalendarEventConfirmed)},
		},
		{
			name: "moved into the week from the next one",
			exceptions: []models.CalendarEvent{
				exception(gym, at(10, 30, 18), at(10, 22, 18), at(10, 22, 19), models.CalendarEventConfirmed),
			},
			wantIDs:    []string{gym.ID + "_20261030T210000Z", gym.ID + "_20261023T210000Z"},
			wantStarts: []time.Time{at(10, 22, 18), at(10, 23, 18)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Occurrences(nil, []models.CalendarEvent{gym}, tt.exceptions, from, to)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("got %d occurrences %+v, want %d", len(got), got, len(tt.wantIDs))
			}
			for i, o := range got {
				if o.ID != tt.wantIDs[i] || !o.StartAt.Equal(tt.wantStarts[i]) {
					t.Errorf("occurrence %d = %s at %v, want %s at %v", i, o.ID, o.StartAt, tt.wantIDs[i], tt.wantStarts[i])
				}
				if o.RecurringEventID == nil || *o.RecurringEventID != gym.ID || len(o.Recurrence) == 0 {
					t.Errorf("occurrence %d isn't tied to its series: %+v", i, o)
				}
			}
		})
	}
}

func TestOccurrencesMergesSinglesInOrder(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	at := func(day, hour int) time.Time { return time.Date(2026, 10, day, hour, 0, 0, 0, sp) }
	breakfast := series(at(7, 20), at(7, 21), "America/Sao_Paulo", false, "RRULE:FREQ=DAILY")
	breakfast.ID = "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
	dentist := models.CalendarEvent{ID: "9e8d7c6b-5a4f-4e3d-2c1b-0a9f8e7d6c5b", Title: "Dentist", StartAt: at(9, 10), EndAt: at(9, 11), Timezone: "America/Sao_Paulo"}

	got := Occurrences([]models.CalendarEvent{dentist}, []models.CalendarEvent{breakfast}, nil, at(9, 0), at(10, 0))
	if len(got) != 2 || got[0].ID != dentist.ID || !got[1].StartAt.Equal(at(9, 20)) {
		t.Errorf("got %+v, want the dentist then breakfast at 8 p.m.", got)
	}
}
