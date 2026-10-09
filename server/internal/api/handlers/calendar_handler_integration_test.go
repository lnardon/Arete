package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/config"
	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
	"github.com/lnardon/arete/internal/repository"
)

// These tests run the calendar handler against a real Postgres. They're
// skipped unless TEST_DB_NAME is set, e.g.:
//
//	docker run -d --rm -p 127.0.0.1:55499:5432 -e POSTGRES_USER=arete \
//	  -e POSTGRES_PASSWORD=arete -e POSTGRES_DB=arete_test postgres:17
//	TEST_DB_PORT=55499 TEST_DB_NAME=arete_test go test ./internal/api/handlers/
//
// Use a throwaway database: the migrations run against it.

var testDB *database.DB

func TestMain(m *testing.M) {
	if os.Getenv("TEST_DB_NAME") != "" {
		// Migrations are found relative to the working directory.
		if err := os.Chdir("../../.."); err != nil {
			fmt.Fprintln(os.Stderr, "chdir to server root:", err)
			os.Exit(1)
		}
		db, err := database.New(config.DatabaseConfig{
			Host:     envOr("TEST_DB_HOST", "localhost"),
			Port:     envOr("TEST_DB_PORT", "5432"),
			User:     envOr("TEST_DB_USER", "arete"),
			Password: envOr("TEST_DB_PASSWORD", "arete"),
			DBName:   os.Getenv("TEST_DB_NAME"),
			SSLMode:  "disable",
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "connect test database:", err)
			os.Exit(1)
		}
		testDB = db
	}
	os.Exit(m.Run())
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type fakeGoogle struct {
	pushes  int
	deleted []string
}

func (f *fakeGoogle) DeleteRemoteEvent(_ context.Context, _, googleEventID string) error {
	f.deleted = append(f.deleted, googleEventID)
	return nil
}

func (f *fakeGoogle) PushUser(context.Context, string) error {
	f.pushes++
	return nil
}

type calendarEnv struct {
	t      *testing.T
	userID string
	router *mux.Router
	google *fakeGoogle
	repo   *repository.CalendarEventRepository
}

func newCalendarEnv(t *testing.T) *calendarEnv {
	t.Helper()
	if testDB == nil {
		t.Skip("TEST_DB_NAME not set")
	}

	var userID string
	err := testDB.QueryRow(`INSERT INTO users (username, password_hash) VALUES ($1, 'x') RETURNING id`,
		fmt.Sprintf("calendar-test-%d", time.Now().UnixNano())).Scan(&userID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { testDB.Exec(`DELETE FROM users WHERE id = $1`, userID) })

	env := &calendarEnv{t: t, userID: userID, google: &fakeGoogle{}, repo: repository.NewCalendarEventRepository(testDB)}
	h := NewCalendarEventHandler(env.repo, env.google)
	r := mux.NewRouter()
	r.HandleFunc("/calendar/events", h.ListEvents).Methods("GET")
	r.HandleFunc("/calendar/events", h.CreateEvent).Methods("POST")
	r.HandleFunc("/calendar/events/{id}", h.GetEvent).Methods("GET")
	r.HandleFunc("/calendar/events/{id}", h.UpdateEvent).Methods("PUT")
	r.HandleFunc("/calendar/events/{id}", h.DeleteEvent).Methods("DELETE")
	env.router = r
	return env
}

func (e *calendarEnv) do(method, path string, body any) *httptest.ResponseRecorder {
	e.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			e.t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req = req.WithContext(auth.WithUser(req.Context(), auth.AuthUser{ID: e.userID}))
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func (e *calendarEnv) mustDo(method, path string, body any, wantStatus int) *httptest.ResponseRecorder {
	e.t.Helper()
	rec := e.do(method, path, body)
	if rec.Code != wantStatus {
		e.t.Fatalf("%s %s = %d (%s), want %d", method, path, rec.Code, strings.TrimSpace(rec.Body.String()), wantStatus)
	}
	return rec
}

func (e *calendarEnv) create(body map[string]any) models.CalendarEvent {
	e.t.Helper()
	var ev models.CalendarEvent
	if err := json.Unmarshal(e.mustDo("POST", "/calendar/events", body, http.StatusCreated).Body.Bytes(), &ev); err != nil {
		e.t.Fatalf("decode created event: %v", err)
	}
	return ev
}

func (e *calendarEnv) list(from, to time.Time) []models.CalendarEvent {
	e.t.Helper()
	path := fmt.Sprintf("/calendar/events?start=%s&end=%s", from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))
	var events []models.CalendarEvent
	if err := json.Unmarshal(e.mustDo("GET", path, nil, http.StatusOK).Body.Bytes(), &events); err != nil {
		e.t.Fatalf("decode list: %v", err)
	}
	return events
}

func (e *calendarEnv) count(query string, args ...any) int {
	e.t.Helper()
	var n int
	if err := testDB.QueryRow(query, args...).Scan(&n); err != nil {
		e.t.Fatalf("count: %v", err)
	}
	return n
}

var saoPaulo = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func sp(month time.Month, day, hour int) time.Time {
	return time.Date(2026, month, day, hour, 0, 0, 0, saoPaulo)
}

func eventBody(title string, start, end time.Time, recurrence []string) map[string]any {
	return map[string]any{
		"title":      title,
		"startAt":    start.Format(time.RFC3339),
		"endAt":      end.Format(time.RFC3339),
		"allDay":     false,
		"timezone":   "America/Sao_Paulo",
		"recurrence": recurrence,
		"color":      "#22c55e",
	}
}

type wantOccurrence struct {
	start time.Time
	title string
}

func assertOccurrences(t *testing.T, got []models.CalendarEvent, want ...wantOccurrence) {
	t.Helper()
	if len(got) != len(want) {
		var summary []string
		for _, e := range got {
			summary = append(summary, fmt.Sprintf("%s@%s", e.Title, e.StartAt.In(saoPaulo).Format("Jan 2 15:04")))
		}
		t.Fatalf("got %d occurrences %v, want %d", len(got), summary, len(want))
	}
	for i, w := range want {
		if !got[i].StartAt.Equal(w.start) || got[i].Title != w.title {
			t.Errorf("occurrence %d = %q at %s, want %q at %s", i, got[i].Title, got[i].StartAt.In(saoPaulo).Format("Jan 2 15:04"), w.title, w.start.Format("Jan 2 15:04"))
		}
	}
}

func instancePath(seriesID string, slot time.Time) string {
	return "/calendar/events/" + seriesID + "_" + slot.UTC().Format("20060102T150405Z")
}

func TestRecurringEventLifecycle(t *testing.T) {
	env := newCalendarEnv(t)
	gym := env.create(eventBody("Gym class", sp(10, 9, 18), sp(10, 9, 19), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}))
	month := func() []models.CalendarEvent { return env.list(sp(10, 1, 0), sp(11, 1, 0)) }

	assertOccurrences(t, month(),
		wantOccurrence{sp(10, 9, 18), "Gym class"},
		wantOccurrence{sp(10, 16, 18), "Gym class"},
		wantOccurrence{sp(10, 23, 18), "Gym class"},
		wantOccurrence{sp(10, 30, 18), "Gym class"},
	)
	week := env.list(sp(10, 19, 0), sp(10, 26, 0))
	if len(week) != 1 || week[0].ID != gym.ID+"_20261023T210000Z" {
		t.Fatalf("week view = %+v, want the Oct 23 occurrence by instance ID", week)
	}

	// This event: move Oct 23 to 7 p.m.
	env.mustDo("PUT", instancePath(gym.ID, sp(10, 23, 18))+"?scope=this",
		eventBody("Gym class (late)", sp(10, 23, 19), sp(10, 23, 20), nil), http.StatusOK)
	// This event: delete Oct 30.
	env.mustDo("DELETE", instancePath(gym.ID, sp(10, 30, 18))+"?scope=this", nil, http.StatusNoContent)
	env.mustDo("GET", instancePath(gym.ID, sp(10, 30, 18)), nil, http.StatusNotFound)

	assertOccurrences(t, month(),
		wantOccurrence{sp(10, 9, 18), "Gym class"},
		wantOccurrence{sp(10, 16, 18), "Gym class"},
		wantOccurrence{sp(10, 23, 19), "Gym class (late)"},
	)
	// The edited occurrence keeps its instance ID.
	var moved models.CalendarEvent
	json.Unmarshal(env.mustDo("GET", instancePath(gym.ID, sp(10, 23, 18)), nil, http.StatusOK).Body.Bytes(), &moved)
	if !moved.StartAt.Equal(sp(10, 23, 19)) {
		t.Errorf("GET edited occurrence starts %v, want Oct 23 19:00", moved.StartAt)
	}

	// All events, from Oct 16: rename and move 6 p.m. -> 7 p.m. The rename
	// reaches the edited occurrence, which keeps its own (already 7 p.m.)
	// time; the deleted Oct 30 stays deleted.
	env.mustDo("PUT", instancePath(gym.ID, sp(10, 16, 18))+"?scope=all",
		eventBody("Gym", sp(10, 16, 19), sp(10, 16, 20), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}), http.StatusOK)
	assertOccurrences(t, month(),
		wantOccurrence{sp(10, 9, 19), "Gym"},
		wantOccurrence{sp(10, 16, 19), "Gym"},
		wantOccurrence{sp(10, 23, 19), "Gym"},
	)

	// This and following, from Oct 16: move to 8 p.m. The old series ends
	// before Oct 16; the later exceptions move to the new series.
	var next models.CalendarEvent
	json.Unmarshal(env.mustDo("PUT", instancePath(gym.ID, sp(10, 16, 19))+"?scope=following",
		eventBody("Gym", sp(10, 16, 20), sp(10, 16, 21), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}), http.StatusOK).Body.Bytes(), &next)
	if next.ID == "" || next.ID == gym.ID {
		t.Fatalf("following edit returned %+v, want a new series", next)
	}
	assertOccurrences(t, month(),
		wantOccurrence{sp(10, 9, 19), "Gym"},
		wantOccurrence{sp(10, 16, 20), "Gym"},
		wantOccurrence{sp(10, 23, 20), "Gym"},
	)
	if n := env.count(`SELECT count(*) FROM calendar_events WHERE recurring_event_id = $1 AND status = 'cancelled'`, next.ID); n != 1 {
		t.Errorf("new series has %d cancelled exceptions, want the carried Oct 30", n)
	}

	// Delete this and following on the new series, from Oct 23.
	env.mustDo("DELETE", instancePath(next.ID, sp(10, 23, 20))+"?scope=following", nil, http.StatusNoContent)
	assertOccurrences(t, month(),
		wantOccurrence{sp(10, 9, 19), "Gym"},
		wantOccurrence{sp(10, 16, 20), "Gym"},
	)

	// Delete all events of the original series.
	env.mustDo("DELETE", instancePath(gym.ID, sp(10, 9, 19))+"?scope=all", nil, http.StatusNoContent)
	assertOccurrences(t, month(), wantOccurrence{sp(10, 16, 20), "Gym"})

	if env.google.pushes != 0 || len(env.google.deleted) != 0 {
		t.Errorf("unsynced series touched Google: %+v", env.google)
	}
}

func TestFollowingSplitKeepsCount(t *testing.T) {
	env := newCalendarEnv(t)
	daily := env.create(eventBody("Breakfast", sp(10, 9, 20), sp(10, 9, 21), []string{"RRULE:FREQ=DAILY;COUNT=5"}))

	env.mustDo("PUT", instancePath(daily.ID, sp(10, 11, 20))+"?scope=following",
		eventBody("Dinner", sp(10, 11, 20), sp(10, 11, 21), []string{"RRULE:FREQ=DAILY;COUNT=5"}), http.StatusOK)

	assertOccurrences(t, env.list(sp(10, 1, 0), sp(11, 1, 0)),
		wantOccurrence{sp(10, 9, 20), "Breakfast"},
		wantOccurrence{sp(10, 10, 20), "Breakfast"},
		wantOccurrence{sp(10, 11, 20), "Dinner"},
		wantOccurrence{sp(10, 12, 20), "Dinner"},
		wantOccurrence{sp(10, 13, 20), "Dinner"},
	)
}

func TestAllEventsTextEditPatchesEditedOccurrences(t *testing.T) {
	env := newCalendarEnv(t)
	gym := env.create(eventBody("Gym class", sp(10, 9, 18), sp(10, 9, 19), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}))
	env.mustDo("PUT", instancePath(gym.ID, sp(10, 16, 18))+"?scope=this",
		eventBody("Gym class", sp(10, 16, 7), sp(10, 16, 8), nil), http.StatusOK)

	// Only the title changes, so exceptions stay put and get the new title.
	env.mustDo("PUT", instancePath(gym.ID, sp(10, 23, 18))+"?scope=all",
		eventBody("Crossfit", sp(10, 23, 18), sp(10, 23, 19), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}), http.StatusOK)

	assertOccurrences(t, env.list(sp(10, 9, 0), sp(10, 24, 0)),
		wantOccurrence{sp(10, 9, 18), "Crossfit"},
		wantOccurrence{sp(10, 16, 7), "Crossfit"},
		wantOccurrence{sp(10, 23, 18), "Crossfit"},
	)
}

func TestSeriesAndSingleConversions(t *testing.T) {
	env := newCalendarEnv(t)
	single := env.create(eventBody("Dentist", sp(10, 9, 10), sp(10, 9, 11), nil))

	// A single event becomes a series by giving it a rule.
	env.mustDo("PUT", "/calendar/events/"+single.ID, eventBody("Dentist", sp(10, 9, 10), sp(10, 9, 11), []string{"RRULE:FREQ=WEEKLY"}), http.StatusOK)
	if got := env.list(sp(10, 1, 0), sp(11, 1, 0)); len(got) != 4 {
		t.Fatalf("converted series lists %d occurrences, want 4", len(got))
	}

	// "Does not repeat" on all events collapses it into the edited occurrence.
	env.mustDo("PUT", instancePath(single.ID, sp(10, 23, 10))+"?scope=all", eventBody("Dentist", sp(10, 23, 10), sp(10, 23, 11), nil), http.StatusOK)
	assertOccurrences(t, env.list(sp(10, 1, 0), sp(11, 1, 0)), wantOccurrence{sp(10, 23, 10), "Dentist"})
}

func TestRecurringEventValidation(t *testing.T) {
	env := newCalendarEnv(t)
	gym := env.create(eventBody("Gym class", sp(10, 9, 18), sp(10, 9, 19), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}))
	occurrence := instancePath(gym.ID, sp(10, 16, 18))

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{"invalid rule", "POST", "/calendar/events", eventBody("x", sp(10, 9, 18), sp(10, 9, 19), []string{"RRULE:FREQ=HOURLY"})},
		{"DTSTART line", "POST", "/calendar/events", eventBody("x", sp(10, 9, 18), sp(10, 9, 19), []string{"DTSTART:20261009T210000Z", "RRULE:FREQ=DAILY"})},
		{"occurrence edit without scope", "PUT", occurrence, eventBody("x", sp(10, 16, 18), sp(10, 16, 19), nil)},
		{"rule on a single-occurrence edit", "PUT", occurrence + "?scope=this", eventBody("x", sp(10, 16, 18), sp(10, 16, 19), []string{"RRULE:FREQ=DAILY"})},
		{"scope on a whole series", "DELETE", "/calendar/events/" + gym.ID + "?scope=this", nil},
		{"range over 92 days", "GET", "/calendar/events?start=2026-01-01T00:00:00Z&end=2026-06-01T00:00:00Z", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env.mustDo(tt.method, tt.path, tt.body, http.StatusBadRequest)
		})
	}

	env.mustDo("GET", instancePath(gym.ID, sp(10, 17, 18)), nil, http.StatusNotFound) // a Saturday
	env.mustDo("GET", "/calendar/events/not-a-uuid", nil, http.StatusNotFound)
}

func TestGoogleLinkedSeriesPushesRightAway(t *testing.T) {
	env := newCalendarEnv(t)
	gym := env.create(eventBody("Gym class", sp(10, 9, 18), sp(10, 9, 19), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}))
	if _, err := testDB.Exec(`UPDATE calendar_events SET google_event_id = 'series123', google_synced_at = NOW() WHERE id = $1`, gym.ID); err != nil {
		t.Fatal(err)
	}

	env.mustDo("PUT", instancePath(gym.ID, sp(10, 23, 18))+"?scope=this", eventBody("Gym class", sp(10, 23, 19), sp(10, 23, 20), nil), http.StatusOK)
	if env.google.pushes != 1 {
		t.Errorf("pushes after an occurrence edit = %d, want 1", env.google.pushes)
	}

	// The exception is dirty, and series come before exceptions in the push.
	dirty, err := env.repo.ListDirtyForPush(context.Background(), env.userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirty) != 1 || dirty[0].RecurringEventID == nil {
		t.Fatalf("dirty rows = %+v, want just the exception", dirty)
	}
	if _, err := testDB.Exec(`UPDATE calendar_events SET google_event_id = 'series123_20261023T210000Z', google_synced_at = NOW() WHERE id = $1`, dirty[0].ID); err != nil {
		t.Fatal(err)
	}

	// Splitting from Oct 16 moves the Oct 23 exception to the new series; its
	// old Google instance is cancelled after the push.
	env.mustDo("PUT", instancePath(gym.ID, sp(10, 16, 18))+"?scope=following",
		eventBody("Gym", sp(10, 16, 18), sp(10, 16, 19), []string{"RRULE:FREQ=WEEKLY;BYDAY=FR"}), http.StatusOK)
	if env.google.pushes != 2 || len(env.google.deleted) != 1 || env.google.deleted[0] != "series123_20261023T210000Z" {
		t.Errorf("google calls = %+v, want a second push and the old exception cancelled", env.google)
	}

	dirty, _ = env.repo.ListDirtyForPush(context.Background(), env.userID)
	if len(dirty) != 3 || dirty[0].RecurringEventID != nil || dirty[1].RecurringEventID != nil || dirty[2].RecurringEventID == nil {
		t.Errorf("dirty rows = %d, want the trimmed series, the new series, then the carried exception", len(dirty))
	}
}
