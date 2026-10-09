package recurrence

import (
	"errors"
	"testing"
	"time"

	"github.com/lnardon/arete/internal/models"
)

func mustLoad(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return loc
}

func series(start, end time.Time, tz string, allDay bool, lines ...string) models.CalendarEvent {
	return models.CalendarEvent{
		ID:         "6f1c2a9e-3b7d-4c1e-9a2f-0d8e7b6c5a41",
		Title:      "Gym class",
		StartAt:    start,
		EndAt:      end,
		Timezone:   tz,
		AllDay:     allDay,
		Recurrence: lines,
		Status:     models.CalendarEventConfirmed,
		Color:      "#22c55e",
	}
}

func expandOrFail(t *testing.T, s models.CalendarEvent, from, to time.Time) []time.Time {
	t.Helper()
	got, err := Expand(s, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	return got
}

func assertTimes(t *testing.T, got []time.Time, want ...time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d occurrences %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("occurrence %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestExpandKeepsLocalTimeAcrossDST(t *testing.T) {
	ny := mustLoad(t, "America/New_York")
	// Breakfast every day at 8 p.m.; DST ends in New York on 2026-11-01.
	s := series(time.Date(2026, 10, 28, 20, 0, 0, 0, ny), time.Date(2026, 10, 28, 20, 30, 0, 0, ny), "America/New_York", false, "RRULE:FREQ=DAILY")

	got := expandOrFail(t, s, time.Date(2026, 10, 31, 0, 0, 0, 0, ny), time.Date(2026, 11, 3, 0, 0, 0, 0, ny))
	assertTimes(t, got,
		time.Date(2026, 10, 31, 20, 0, 0, 0, ny),
		time.Date(2026, 11, 1, 20, 0, 0, 0, ny),
		time.Date(2026, 11, 2, 20, 0, 0, 0, ny),
	)
	// Same wall-clock time, different UTC time: the offset changed.
	if got[0].UTC().Hour() != 0 || got[1].UTC().Hour() != 1 {
		t.Errorf("UTC hours = %d, %d; want 0 then 1", got[0].UTC().Hour(), got[1].UTC().Hour())
	}
}

func TestExpandWeekly(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	s := series(time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp), "America/Sao_Paulo", false, "RRULE:FREQ=WEEKLY;BYDAY=FR")

	got := expandOrFail(t, s, time.Date(2026, 10, 5, 0, 0, 0, 0, sp), time.Date(2026, 11, 2, 0, 0, 0, 0, sp))
	assertTimes(t, got,
		time.Date(2026, 10, 9, 18, 0, 0, 0, sp),
		time.Date(2026, 10, 16, 18, 0, 0, 0, sp),
		time.Date(2026, 10, 23, 18, 0, 0, 0, sp),
		time.Date(2026, 10, 30, 18, 0, 0, 0, sp),
	)
}

func TestExpandIncludesOccurrenceStartedBeforeRange(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	s := series(time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 20, 0, 0, 0, sp), "America/Sao_Paulo", false, "RRULE:FREQ=DAILY")

	got := expandOrFail(t, s, time.Date(2026, 10, 10, 19, 0, 0, 0, sp), time.Date(2026, 10, 10, 23, 0, 0, 0, sp))
	assertTimes(t, got, time.Date(2026, 10, 10, 18, 0, 0, 0, sp))
}

func TestExpandAllDay(t *testing.T) {
	ny := mustLoad(t, "America/New_York")
	// A one-day all-day event spanning the DST change still lasts one day.
	s := series(time.Date(2026, 10, 25, 0, 0, 0, 0, ny), time.Date(2026, 10, 26, 0, 0, 0, 0, ny), "America/New_York", true, "RRULE:FREQ=WEEKLY")

	got := expandOrFail(t, s, time.Date(2026, 10, 30, 0, 0, 0, 0, ny), time.Date(2026, 11, 9, 0, 0, 0, 0, ny))
	assertTimes(t, got, time.Date(2026, 11, 1, 0, 0, 0, 0, ny), time.Date(2026, 11, 8, 0, 0, 0, 0, ny))
	if end := OccurrenceEnd(s, got[0]); !end.Equal(time.Date(2026, 11, 2, 0, 0, 0, 0, ny)) {
		t.Errorf("all-day occurrence ends %v, want local midnight Nov 2", end)
	}
}

func TestExpandEndConditions(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	start, end := time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp)
	from, to := time.Date(2026, 10, 1, 0, 0, 0, 0, sp), time.Date(2026, 11, 1, 0, 0, 0, 0, sp)

	tests := []struct {
		name string
		line string
		want int
	}{
		{"count", "RRULE:FREQ=DAILY;COUNT=3", 3},
		// 18:00 in São Paulo is 21:00Z, so Oct 12 is the last day kept.
		{"until UTC", "RRULE:FREQ=DAILY;UNTIL=20261012T235959Z", 4},
		{"until date on a timed series keeps that whole day", "RRULE:FREQ=DAILY;UNTIL=20261012", 4},
		{"interval", "RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandOrFail(t, series(start, end, "America/Sao_Paulo", false, tt.line), from, to)
			if len(got) != tt.want {
				t.Errorf("got %d occurrences %v, want %d", len(got), got, tt.want)
			}
		})
	}
}

func TestExpandHonorsGoogleExdates(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	from, to := time.Date(2026, 10, 9, 0, 0, 0, 0, sp), time.Date(2026, 10, 13, 0, 0, 0, 0, sp)

	timed := series(time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp), "America/Sao_Paulo", false,
		"RRULE:FREQ=DAILY", "EXDATE;TZID=America/Sao_Paulo:20261011T180000")
	assertTimes(t, expandOrFail(t, timed, from, to),
		time.Date(2026, 10, 9, 18, 0, 0, 0, sp),
		time.Date(2026, 10, 10, 18, 0, 0, 0, sp),
		time.Date(2026, 10, 12, 18, 0, 0, 0, sp),
	)

	allDay := series(time.Date(2026, 10, 9, 0, 0, 0, 0, sp), time.Date(2026, 10, 10, 0, 0, 0, 0, sp), "America/Sao_Paulo", true,
		"RRULE:FREQ=DAILY", "EXDATE;VALUE=DATE:20261011")
	assertTimes(t, expandOrFail(t, allDay, from, to),
		time.Date(2026, 10, 9, 0, 0, 0, 0, sp),
		time.Date(2026, 10, 10, 0, 0, 0, 0, sp),
		time.Date(2026, 10, 12, 0, 0, 0, 0, sp),
	)
}

func TestExpandAlwaysIncludesStart(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	// RFC 5545: DTSTART is the first instance even when the rule (Fridays)
	// wouldn't produce it (a Wednesday).
	s := series(time.Date(2026, 10, 7, 18, 0, 0, 0, sp), time.Date(2026, 10, 7, 19, 0, 0, 0, sp), "America/Sao_Paulo", false, "RRULE:FREQ=WEEKLY;BYDAY=FR")
	got := expandOrFail(t, s, time.Date(2026, 10, 1, 0, 0, 0, 0, sp), time.Date(2026, 10, 12, 0, 0, 0, 0, sp))
	assertTimes(t, got, time.Date(2026, 10, 7, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 18, 0, 0, 0, sp))
}

func TestExpandUnionsSeveralRules(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	s := series(time.Date(2026, 10, 5, 18, 0, 0, 0, sp), time.Date(2026, 10, 5, 19, 0, 0, 0, sp), "America/Sao_Paulo", false,
		"RRULE:FREQ=WEEKLY;BYDAY=MO", "RRULE:FREQ=WEEKLY;BYDAY=TH")
	got := expandOrFail(t, s, time.Date(2026, 10, 5, 0, 0, 0, 0, sp), time.Date(2026, 10, 12, 0, 0, 0, 0, sp))
	assertTimes(t, got, time.Date(2026, 10, 5, 18, 0, 0, 0, sp), time.Date(2026, 10, 8, 18, 0, 0, 0, sp))
}

func TestEnd(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	start, end := time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp)

	got, err := End(series(start, end, "America/Sao_Paulo", false, "RRULE:FREQ=DAILY"))
	if err != nil || got != nil {
		t.Errorf("End of an endless series = %v, %v; want nil, nil", got, err)
	}

	got, err = End(series(start, end, "America/Sao_Paulo", false, "RRULE:FREQ=DAILY;COUNT=3"))
	if err != nil {
		t.Fatalf("End: %v", err)
	}
	if want := time.Date(2026, 10, 11, 19, 0, 0, 0, sp); got == nil || !got.Equal(want) {
		t.Errorf("End = %v, want %v", got, want)
	}
}

func TestIsOccurrence(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	s := series(time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp), "America/Sao_Paulo", false, "RRULE:FREQ=WEEKLY;BYDAY=FR")
	if !IsOccurrence(s, time.Date(2026, 10, 23, 18, 0, 0, 0, sp)) {
		t.Error("Oct 23 18:00 should be an occurrence")
	}
	if IsOccurrence(s, time.Date(2026, 10, 23, 19, 0, 0, 0, sp)) {
		t.Error("Oct 23 19:00 should not be an occurrence")
	}
	if IsOccurrence(s, time.Date(2026, 10, 24, 18, 0, 0, 0, sp)) {
		t.Error("Saturday Oct 24 should not be an occurrence")
	}
}

func TestTrimBeforeAndReduceCount(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	s := series(time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp), "America/Sao_Paulo", false,
		"RRULE:FREQ=DAILY;COUNT=10", "EXDATE;TZID=America/Sao_Paulo:20261020T180000")
	split := time.Date(2026, 10, 12, 18, 0, 0, 0, sp)

	lines, kept, err := TrimBefore(s, split)
	if err != nil {
		t.Fatalf("TrimBefore: %v", err)
	}
	if kept != 3 {
		t.Errorf("kept = %d, want 3", kept)
	}
	want := []string{"RRULE:FREQ=DAILY;UNTIL=20261012T205959Z", "EXDATE;TZID=America/Sao_Paulo:20261020T180000"}
	if !SameLines(lines, want) {
		t.Errorf("lines = %q, want %q", lines, want)
	}

	trimmed := s
	trimmed.Recurrence = lines
	got := expandOrFail(t, trimmed, time.Date(2026, 10, 1, 0, 0, 0, 0, sp), time.Date(2026, 11, 1, 0, 0, 0, 0, sp))
	if len(got) != 3 || !got[2].Equal(time.Date(2026, 10, 11, 18, 0, 0, 0, sp)) {
		t.Errorf("trimmed series expands to %v, want Oct 9-11", got)
	}

	if reduced := ReduceCount(s.Recurrence, kept); reduced[0] != "RRULE:FREQ=DAILY;COUNT=7" {
		t.Errorf("ReduceCount = %q, want COUNT=7", reduced[0])
	}
}

func TestTrimBeforeAllDayUsesDateUntil(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	s := series(time.Date(2026, 10, 9, 0, 0, 0, 0, sp), time.Date(2026, 10, 10, 0, 0, 0, 0, sp), "America/Sao_Paulo", true, "RRULE:FREQ=DAILY")

	lines, kept, err := TrimBefore(s, time.Date(2026, 10, 12, 0, 0, 0, 0, sp))
	if err != nil {
		t.Fatalf("TrimBefore: %v", err)
	}
	if kept != 3 || lines[0] != "RRULE:FREQ=DAILY;UNTIL=20261011" {
		t.Errorf("TrimBefore = %q, %d; want UNTIL=20261011, 3", lines, kept)
	}
	if err := Validate(lines, true); err != nil {
		t.Errorf("trimmed all-day rule fails validation: %v", err)
	}
}

func TestInstanceIDRoundTrip(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	timed := series(time.Date(2026, 10, 9, 18, 0, 0, 0, sp), time.Date(2026, 10, 9, 19, 0, 0, 0, sp), "America/Sao_Paulo", false, "RRULE:FREQ=WEEKLY")
	allDay := series(time.Date(2026, 10, 9, 0, 0, 0, 0, sp), time.Date(2026, 10, 10, 0, 0, 0, 0, sp), "America/Sao_Paulo", true, "RRULE:FREQ=WEEKLY")

	tests := []struct {
		name   string
		series models.CalendarEvent
		slot   time.Time
		want   string
	}{
		{"timed, stamped in UTC", timed, time.Date(2026, 10, 23, 18, 0, 0, 0, sp), timed.ID + "_20261023T210000Z"},
		{"all-day, stamped as a date", allDay, time.Date(2026, 10, 23, 0, 0, 0, 0, sp), allDay.ID + "_20261023"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := InstanceID(tt.series, tt.slot)
			if id != tt.want {
				t.Fatalf("InstanceID = %s, want %s", id, tt.want)
			}
			seriesID, stamp, ok := ParseInstanceID(id)
			if !ok || seriesID != tt.series.ID {
				t.Fatalf("ParseInstanceID(%s) = %s, %s, %v", id, seriesID, stamp, ok)
			}
			slot, err := ParseStamp(tt.series, stamp)
			if err != nil || !slot.Equal(tt.slot) {
				t.Errorf("ParseStamp = %v, %v; want %v", slot, err, tt.slot)
			}
		})
	}

	if _, _, ok := ParseInstanceID(timed.ID); ok {
		t.Error("a plain UUID parsed as an instance ID")
	}
}

func TestValidate(t *testing.T) {
	valid := []struct {
		lines  []string
		allDay bool
	}{
		{[]string{"RRULE:FREQ=DAILY"}, false},
		{[]string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}, false},
		{[]string{"RRULE:FREQ=MONTHLY;BYDAY=2FR;UNTIL=20261231T235959Z"}, false},
		{[]string{"RRULE:FREQ=YEARLY;COUNT=5"}, true},
		{[]string{"RRULE:FREQ=WEEKLY;UNTIL=20261231"}, true},
		{[]string{"RRULE:FREQ=DAILY", "EXDATE;TZID=America/Sao_Paulo:20261016T180000"}, false},
	}
	for _, tt := range valid {
		if err := Validate(tt.lines, tt.allDay); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", tt.lines, err)
		}
	}

	invalidCases := []struct {
		name   string
		lines  []string
		allDay bool
	}{
		{"DTSTART line", []string{"DTSTART:20261009T210000Z", "RRULE:FREQ=DAILY"}, false},
		{"COUNT and UNTIL", []string{"RRULE:FREQ=DAILY;COUNT=3;UNTIL=20261231T235959Z"}, false},
		{"sub-daily frequency", []string{"RRULE:FREQ=HOURLY"}, false},
		{"local UNTIL on a timed series", []string{"RRULE:FREQ=DAILY;UNTIL=20261231T235959"}, false},
		{"date-time UNTIL on an all-day series", []string{"RRULE:FREQ=DAILY;UNTIL=20261231T235959Z"}, true},
		{"two RRULEs", []string{"RRULE:FREQ=DAILY", "RRULE:FREQ=WEEKLY"}, false},
		{"no RRULE", []string{"EXDATE:20261016T210000Z"}, false},
		{"COUNT too large", []string{"RRULE:FREQ=DAILY;COUNT=5000"}, false},
		{"unparsable", []string{"RRULE:FREQ=WEEKLY;BYDAY=XX"}, false},
	}
	for _, tt := range invalidCases {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.lines, tt.allDay); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate(%q) = %v, want ErrInvalid", tt.lines, err)
			}
		})
	}
}

func TestShift(t *testing.T) {
	sp := mustLoad(t, "America/Sao_Paulo")
	ny := mustLoad(t, "America/New_York")

	// 6 p.m. -> 7 p.m. on the same day: every slot keeps its date.
	s := NewShift(time.Date(2026, 10, 23, 18, 0, 0, 0, sp), sp, time.Date(2026, 10, 23, 19, 0, 0, 0, sp), sp)
	if got, want := s.Apply(time.Date(2026, 10, 30, 18, 0, 0, 0, sp)), time.Date(2026, 10, 30, 19, 0, 0, 0, sp); !got.Equal(want) {
		t.Errorf("same-day shift = %v, want %v", got, want)
	}

	// Friday -> Saturday: every slot moves a day.
	s = NewShift(time.Date(2026, 10, 23, 18, 0, 0, 0, sp), sp, time.Date(2026, 10, 24, 18, 0, 0, 0, sp), sp)
	if got, want := s.Apply(time.Date(2026, 10, 30, 18, 0, 0, 0, sp)), time.Date(2026, 10, 31, 18, 0, 0, 0, sp); !got.Equal(want) {
		t.Errorf("next-day shift = %v, want %v", got, want)
	}

	// The wall-clock time survives a DST change between the two slots.
	s = NewShift(time.Date(2026, 10, 28, 20, 0, 0, 0, ny), ny, time.Date(2026, 10, 28, 21, 0, 0, 0, ny), ny)
	if got, want := s.Apply(time.Date(2026, 11, 4, 20, 0, 0, 0, ny)), time.Date(2026, 11, 4, 21, 0, 0, 0, ny); !got.Equal(want) {
		t.Errorf("shift across DST = %v, want %v", got, want)
	}

	// No change is the identity.
	slot := time.Date(2026, 10, 30, 18, 0, 0, 0, sp)
	s = NewShift(slot, sp, slot, sp)
	if got := s.Apply(time.Date(2026, 11, 6, 18, 0, 0, 0, sp)); !got.Equal(time.Date(2026, 11, 6, 18, 0, 0, 0, sp)) {
		t.Errorf("identity shift = %v", got)
	}
}
