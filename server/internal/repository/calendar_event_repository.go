package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

const calendarEventColumns = `id, user_id, title, description, location, start_at, end_at, all_day, timezone, recurrence_rule, color, google_event_id, google_calendar_id, source, created_at, updated_at`

type CalendarEventRepository struct {
	db *database.DB
}

func NewCalendarEventRepository(db *database.DB) *CalendarEventRepository {
	return &CalendarEventRepository{db: db}
}

func scanCalendarEvent(row interface {
	Scan(dest ...any) error
}, e *models.CalendarEvent) error {
	return row.Scan(
		&e.ID, &e.UserID, &e.Title, &e.Description, &e.Location,
		&e.StartAt, &e.EndAt, &e.AllDay, &e.Timezone, &e.RecurrenceRule, &e.Color,
		&e.GoogleEventID, &e.GoogleCalendarID, &e.Source, &e.CreatedAt, &e.UpdatedAt,
	)
}

func (r *CalendarEventRepository) ListEventsInRange(ctx context.Context, userID string, rangeStart, rangeEnd time.Time) ([]models.CalendarEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1 AND start_at < $3 AND end_at > $2
		 ORDER BY start_at ASC`,
		userID, rangeStart, rangeEnd,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		var e models.CalendarEvent
		if err := scanCalendarEvent(rows, &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if events == nil {
		events = []models.CalendarEvent{}
	}
	return events, rows.Err()
}

func (r *CalendarEventRepository) GetEvent(ctx context.Context, userID, id string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := r.db.QueryRowContext(ctx,
		`SELECT `+calendarEventColumns+` FROM calendar_events WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	err := scanCalendarEvent(row, &e)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

func (r *CalendarEventRepository) CreateEvent(ctx context.Context, userID, title string, description, location *string, startAt, endAt time.Time, allDay bool, timezone string, recurrenceRule *string, color string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, recurrence_rule, color)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING `+calendarEventColumns,
		userID, title, description, location, startAt, endAt, allDay, timezone, recurrenceRule, color,
	)
	err := scanCalendarEvent(row, &e)
	return e, err
}

func (r *CalendarEventRepository) UpdateEvent(ctx context.Context, userID, id, title string, description, location *string, startAt, endAt time.Time, allDay bool, timezone string, recurrenceRule *string, color string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := r.db.QueryRowContext(ctx,
		`UPDATE calendar_events
		 SET title = $1, description = $2, location = $3, start_at = $4, end_at = $5,
		     all_day = $6, timezone = $7, recurrence_rule = $8, color = $9, updated_at = NOW()
		 WHERE id = $10 AND user_id = $11
		 RETURNING `+calendarEventColumns,
		title, description, location, startAt, endAt, allDay, timezone, recurrenceRule, color, id, userID,
	)
	err := scanCalendarEvent(row, &e)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

// UpsertFromGoogle is the pull side of sync: it writes Google's version of an
// event into calendar_events, keyed by (user_id, google_event_id). Marking
// google_synced_at = NOW() alongside the write is what keeps this from being
// picked up as "dirty" by ListDirtyForPush on the very next tick. color is
// only applied on insert — it's Arete-only (never pushed to Google), so an
// existing row keeps whatever color the user picked.
func (r *CalendarEventRepository) UpsertFromGoogle(ctx context.Context, userID, googleEventID, googleCalendarID, title string, description, location *string, startAt, endAt time.Time, allDay bool, timezone, color string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	// The WHERE must repeat idx_calendar_events_user_google_event's predicate,
	// or Postgres can't match the partial index to the ON CONFLICT target.
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, color, google_event_id, google_calendar_id, source, google_synced_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'google', NOW())
		 ON CONFLICT (user_id, google_event_id) WHERE google_event_id IS NOT NULL DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   location = EXCLUDED.location,
		   start_at = EXCLUDED.start_at,
		   end_at = EXCLUDED.end_at,
		   all_day = EXCLUDED.all_day,
		   timezone = EXCLUDED.timezone,
		   google_calendar_id = EXCLUDED.google_calendar_id,
		   updated_at = NOW(),
		   google_synced_at = NOW()
		 RETURNING `+calendarEventColumns,
		userID, title, description, location, startAt, endAt, allDay, timezone, color, googleEventID, googleCalendarID,
	)
	err := scanCalendarEvent(row, &e)
	return e, err
}

// DeleteByGoogleEventID handles a pull-side cancellation (an event deleted on
// Google's end).
func (r *CalendarEventRepository) DeleteByGoogleEventID(ctx context.Context, userID, googleEventID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM calendar_events WHERE user_id = $1 AND google_event_id = $2`,
		userID, googleEventID,
	)
	return err
}

// ListDirtyForPush returns events that either have never been pushed to
// Google, or were edited locally since their last push/pull — including
// Google-origin events, since the UI lets users edit those too.
func (r *CalendarEventRepository) ListDirtyForPush(ctx context.Context, userID string) ([]models.CalendarEvent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1
		   AND (google_event_id IS NULL OR updated_at > google_synced_at)
		 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		var e models.CalendarEvent
		if err := scanCalendarEvent(rows, &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if events == nil {
		events = []models.CalendarEvent{}
	}
	return events, rows.Err()
}

// MarkSynced records that a local, app-origin event was just pushed to
// Google, attaching the Google-assigned event ID (on first push) and
// resetting the dirty marker.
func (r *CalendarEventRepository) MarkSynced(ctx context.Context, id, googleEventID, googleCalendarID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE calendar_events SET google_event_id = $1, google_calendar_id = $2, google_synced_at = NOW() WHERE id = $3`,
		googleEventID, googleCalendarID, id,
	)
	return err
}

func (r *CalendarEventRepository) DeleteEvent(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM calendar_events WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
