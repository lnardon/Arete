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
