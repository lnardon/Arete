package google

import (
	"slices"
	"testing"
	"time"

	calendar "google.golang.org/api/calendar/v3"

	"github.com/lnardon/arete/internal/models"
)

func TestApplyLocalFieldsRecurrence(t *testing.T) {
	sp, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	seriesID := "6f1c2a9e-3b7d-4c1e-9a2f-0d8e7b6c5a41"
	original := time.Date(2026, 10, 23, 18, 0, 0, 0, sp)

	timed := models.CalendarEvent{
		Title: "Gym class", StartAt: time.Date(2026, 10, 9, 18, 0, 0, 0, sp), EndAt: time.Date(2026, 10, 9, 19, 0, 0, 0, sp),
		Timezone: "America/Sao_Paulo", Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"},
	}
	allDay := models.CalendarEvent{
		Title: "Review", StartAt: time.Date(2026, 10, 9, 0, 0, 0, 0, sp), EndAt: time.Date(2026, 10, 10, 0, 0, 0, 0, sp),
		AllDay: true, Timezone: "America/Sao_Paulo", Recurrence: []string{"RRULE:FREQ=WEEKLY"},
	}
	single := timed
	single.Recurrence = nil
	exception := single
	exception.RecurringEventID, exception.OriginalStartAt = &seriesID, &original

	t.Run("series sends its rule and zone", func(t *testing.T) {
		ev := &calendar.Event{}
		applyLocalFields(ev, timed, sp)
		if !slices.Equal(ev.Recurrence, timed.Recurrence) || ev.Start.TimeZone != "America/Sao_Paulo" {
			t.Errorf("got recurrence %q, zone %q", ev.Recurrence, ev.Start.TimeZone)
		}
	})

	t.Run("all-day series still names its zone", func(t *testing.T) {
		ev := &calendar.Event{}
		applyLocalFields(ev, allDay, sp)
		if ev.Start.Date != "2026-10-09" || ev.Start.TimeZone != "America/Sao_Paulo" || ev.End.TimeZone != "America/Sao_Paulo" {
			t.Errorf("got start %+v, end %+v", ev.Start, ev.End)
		}
	})

	t.Run("series turned single clears the rule explicitly", func(t *testing.T) {
		ev := &calendar.Event{Recurrence: []string{"RRULE:FREQ=WEEKLY"}}
		applyLocalFields(ev, single, sp)
		if ev.Recurrence != nil || !slices.Contains(ev.NullFields, "Recurrence") {
			t.Errorf("got recurrence %q, null fields %q", ev.Recurrence, ev.NullFields)
		}
	})

	t.Run("instance edit never carries a rule", func(t *testing.T) {
		ev := &calendar.Event{RecurringEventId: "series123"}
		applyLocalFields(ev, exception, sp)
		if ev.Recurrence != nil || len(ev.NullFields) != 0 {
			t.Errorf("got recurrence %q, null fields %q", ev.Recurrence, ev.NullFields)
		}
	})
}

func TestGoogleOriginalStart(t *testing.T) {
	sp, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}

	got, err := googleOriginalStart(&calendar.EventDateTime{DateTime: "2026-10-23T18:00:00-03:00", TimeZone: "America/Sao_Paulo"}, sp)
	if err != nil || !got.Equal(time.Date(2026, 10, 23, 21, 0, 0, 0, time.UTC)) {
		t.Errorf("timed original start = %v, %v", got, err)
	}

	got, err = googleOriginalStart(&calendar.EventDateTime{Date: "2026-10-23"}, sp)
	if err != nil || !got.Equal(time.Date(2026, 10, 23, 0, 0, 0, 0, sp)) {
		t.Errorf("all-day original start = %v, %v; want local midnight", got, err)
	}

	if _, err := googleOriginalStart(nil, sp); err == nil {
		t.Error("missing originalStartTime should fail")
	}
}

func TestGoogleEventToLocalReadsRecurringZone(t *testing.T) {
	ev := &calendar.Event{
		Summary:    "Gym class",
		Start:      &calendar.EventDateTime{DateTime: "2026-10-09T18:00:00-03:00", TimeZone: "America/Sao_Paulo"},
		End:        &calendar.EventDateTime{DateTime: "2026-10-09T19:00:00-03:00", TimeZone: "America/Sao_Paulo"},
		Recurrence: []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"},
	}
	f, err := googleEventToLocal(ev, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if f.Timezone != "America/Sao_Paulo" || f.Title != "Gym class" {
		t.Errorf("got %+v", f)
	}
}
