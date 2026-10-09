package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

const calendarEventColumns = `id, user_id, title, description, location, start_at, end_at, all_day, timezone, recurrence, recurrence_end_at, recurring_event_id, original_start_at, status, color, google_event_id, google_calendar_id, source, created_at, updated_at`

// calendarEventColumnsOf qualifies every column with a table alias, for
// queries that join calendar_events to itself.
func calendarEventColumnsOf(alias string) string {
	cols := strings.Split(calendarEventColumns, ", ")
	for i, c := range cols {
		cols[i] = alias + "." + c
	}
	return strings.Join(cols, ", ")
}

// CalendarEventFields is the user-editable part of an event, shared by every
// write. Recurrence is nil for single events and exceptions.
type CalendarEventFields struct {
	Title           string
	Description     *string
	Location        *string
	StartAt         time.Time
	EndAt           time.Time
	AllDay          bool
	Timezone        string
	Recurrence      []string
	RecurrenceEndAt *time.Time
	Color           string
}

// FieldsOf copies the editable fields out of an existing row.
func FieldsOf(e models.CalendarEvent) CalendarEventFields {
	return CalendarEventFields{
		Title:           e.Title,
		Description:     e.Description,
		Location:        e.Location,
		StartAt:         e.StartAt,
		EndAt:           e.EndAt,
		AllDay:          e.AllDay,
		Timezone:        e.Timezone,
		Recurrence:      e.Recurrence,
		RecurrenceEndAt: e.RecurrenceEndAt,
		Color:           e.Color,
	}
}

// Event presents the fields as an unsaved event, for the recurrence math.
func (f CalendarEventFields) Event() models.CalendarEvent {
	return models.CalendarEvent{
		Title:           f.Title,
		Description:     f.Description,
		Location:        f.Location,
		StartAt:         f.StartAt,
		EndAt:           f.EndAt,
		AllDay:          f.AllDay,
		Timezone:        f.Timezone,
		Recurrence:      f.Recurrence,
		RecurrenceEndAt: f.RecurrenceEndAt,
		Color:           f.Color,
		Status:          models.CalendarEventConfirmed,
	}
}

// recurrenceArg keeps "recurrence IS NOT NULL" meaning "series": an empty
// rule list is stored as NULL, never as an empty array.
func recurrenceArg(lines []string) any {
	if len(lines) == 0 {
		return nil
	}
	return pq.Array(lines)
}

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
		&e.StartAt, &e.EndAt, &e.AllDay, &e.Timezone,
		pq.Array(&e.Recurrence), &e.RecurrenceEndAt, &e.RecurringEventID, &e.OriginalStartAt, &e.Status,
		&e.Color, &e.GoogleEventID, &e.GoogleCalendarID, &e.Source, &e.CreatedAt, &e.UpdatedAt,
	)
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func listCalendarEvents(ctx context.Context, q queryer, query string, args ...any) ([]models.CalendarEvent, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []models.CalendarEvent{}
	for rows.Next() {
		var e models.CalendarEvent
		if err := scanCalendarEvent(rows, &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// ListSinglesInRange returns non-recurring events overlapping [rangeStart, rangeEnd).
func (r *CalendarEventRepository) ListSinglesInRange(ctx context.Context, userID string, rangeStart, rangeEnd time.Time) ([]models.CalendarEvent, error) {
	return listCalendarEvents(ctx, r.db,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1 AND recurrence IS NULL AND recurring_event_id IS NULL
		   AND start_at < $3 AND end_at > $2
		 ORDER BY start_at ASC`,
		userID, rangeStart, rangeEnd,
	)
}

// ListSeriesInRange returns recurring series that may have an occurrence in
// [rangeStart, rangeEnd).
func (r *CalendarEventRepository) ListSeriesInRange(ctx context.Context, userID string, rangeStart, rangeEnd time.Time) ([]models.CalendarEvent, error) {
	return listCalendarEvents(ctx, r.db,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1 AND recurrence IS NOT NULL
		   AND start_at < $3 AND (recurrence_end_at IS NULL OR recurrence_end_at > $2)`,
		userID, rangeStart, rangeEnd,
	)
}

// ListExceptionsInRange returns exceptions that show in [rangeStart,
// rangeEnd) by their own times, or that replace an occurrence slot in it.
func (r *CalendarEventRepository) ListExceptionsInRange(ctx context.Context, userID string, rangeStart, rangeEnd time.Time) ([]models.CalendarEvent, error) {
	return listCalendarEvents(ctx, r.db,
		`SELECT `+calendarEventColumnsOf("x")+`
		 FROM calendar_events x
		 JOIN calendar_events s ON s.id = x.recurring_event_id
		 WHERE x.user_id = $1 AND (
		       (x.status = 'confirmed' AND x.start_at < $3 AND x.end_at > $2)
		    OR (x.original_start_at < $3 AND x.original_start_at + (s.end_at - s.start_at) > $2))`,
		userID, rangeStart, rangeEnd,
	)
}

// ListExceptions returns every exception of one series.
func (r *CalendarEventRepository) ListExceptions(ctx context.Context, userID, seriesID string) ([]models.CalendarEvent, error) {
	return listCalendarEvents(ctx, r.db,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1 AND recurring_event_id = $2
		 ORDER BY original_start_at ASC`,
		userID, seriesID,
	)
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

func (r *CalendarEventRepository) GetByGoogleEventID(ctx context.Context, userID, googleEventID string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := r.db.QueryRowContext(ctx,
		`SELECT `+calendarEventColumns+` FROM calendar_events WHERE user_id = $1 AND google_event_id = $2`,
		userID, googleEventID,
	)
	err := scanCalendarEvent(row, &e)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

// GetException returns the exception replacing the series' occurrence that
// originally started at originalStart.
func (r *CalendarEventRepository) GetException(ctx context.Context, userID, seriesID string, originalStart time.Time) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := r.db.QueryRowContext(ctx,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1 AND recurring_event_id = $2 AND original_start_at = $3`,
		userID, seriesID, originalStart,
	)
	err := scanCalendarEvent(row, &e)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

func (r *CalendarEventRepository) CreateEvent(ctx context.Context, userID string, f CalendarEventFields) (models.CalendarEvent, error) {
	return createEvent(ctx, r.db, userID, f)
}

type rowQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func createEvent(ctx context.Context, q rowQueryer, userID string, f CalendarEventFields) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := q.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, recurrence, recurrence_end_at, color)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING `+calendarEventColumns,
		userID, f.Title, f.Description, f.Location, f.StartAt, f.EndAt, f.AllDay, f.Timezone, recurrenceArg(f.Recurrence), f.RecurrenceEndAt, f.Color,
	)
	err := scanCalendarEvent(row, &e)
	return e, err
}

// UpdateEvent rewrites a single event or a series row. Series edits that
// touch exceptions go through RewriteSeries instead.
func (r *CalendarEventRepository) UpdateEvent(ctx context.Context, userID, id string, f CalendarEventFields) (models.CalendarEvent, error) {
	return updateEvent(ctx, r.db, userID, id, f)
}

func updateEvent(ctx context.Context, q rowQueryer, userID, id string, f CalendarEventFields) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := q.QueryRowContext(ctx,
		`UPDATE calendar_events
		 SET title = $1, description = $2, location = $3, start_at = $4, end_at = $5,
		     all_day = $6, timezone = $7, recurrence = $8, recurrence_end_at = $9, color = $10, updated_at = NOW()
		 WHERE id = $11 AND user_id = $12
		 RETURNING `+calendarEventColumns,
		f.Title, f.Description, f.Location, f.StartAt, f.EndAt, f.AllDay, f.Timezone, recurrenceArg(f.Recurrence), f.RecurrenceEndAt, f.Color, id, userID,
	)
	err := scanCalendarEvent(row, &e)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

// UpsertException edits or cancels one occurrence of a series. The write
// bumps updated_at, which is what marks it for the next Google push.
func (r *CalendarEventRepository) UpsertException(ctx context.Context, userID, seriesID string, originalStart time.Time, f CalendarEventFields, status string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	// The WHERE must repeat idx_calendar_events_exception's predicate, or
	// Postgres can't match the partial index to the ON CONFLICT target.
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, color, recurring_event_id, original_start_at, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (recurring_event_id, original_start_at) WHERE recurring_event_id IS NOT NULL DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   location = EXCLUDED.location,
		   start_at = EXCLUDED.start_at,
		   end_at = EXCLUDED.end_at,
		   all_day = EXCLUDED.all_day,
		   timezone = EXCLUDED.timezone,
		   color = EXCLUDED.color,
		   status = EXCLUDED.status,
		   updated_at = NOW()
		 RETURNING `+calendarEventColumns,
		userID, f.Title, f.Description, f.Location, f.StartAt, f.EndAt, f.AllDay, f.Timezone, f.Color, seriesID, originalStart, status,
	)
	err := scanCalendarEvent(row, &e)
	return e, err
}

// ExceptionWrite is an exception row to insert during a RewriteSeries. It has
// no Google ID, so it pushes to Google as a fresh instance edit.
type ExceptionWrite struct {
	OriginalStartAt time.Time
	Fields          CalendarEventFields
	Status          string
	Source          string
}

// ExceptionPatch overwrites the flagged fields on a series' confirmed
// exceptions, so an "all events" edit reaches occurrences edited one by one.
type ExceptionPatch struct {
	SetTitle       bool
	Title          string
	SetDescription bool
	Description    *string
	SetLocation    bool
	Location       *string
	SetColor       bool
	Color          string
}

func (p ExceptionPatch) Any() bool {
	return p.SetTitle || p.SetDescription || p.SetLocation || p.SetColor
}

// SeriesRewrite is one scoped edit of a recurring series, applied in a single
// transaction.
type SeriesRewrite struct {
	SeriesID string
	// Series holds the existing series' new fields.
	Series CalendarEventFields
	// Create, when set, inserts a second series: the new half of a "this and
	// following" split.
	Create *CalendarEventFields
	// ReplaceExceptions deletes the series' exceptions that originally
	// started at or after ReplaceFrom (all of them when ReplaceFrom is nil).
	ReplaceExceptions bool
	ReplaceFrom       *time.Time
	// Exceptions are inserted under Create when it's set, else under SeriesID.
	Exceptions []ExceptionWrite
	// Patch is applied to the series' confirmed exceptions that remain.
	Patch ExceptionPatch
}

// RewriteSeries applies rw atomically. It returns the updated series, the
// series it created (if any), and the Google IDs of the exceptions it
// deleted, which the caller cancels on Google.
func (r *CalendarEventRepository) RewriteSeries(ctx context.Context, userID string, rw SeriesRewrite) (models.CalendarEvent, *models.CalendarEvent, []string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.CalendarEvent{}, nil, nil, err
	}
	defer tx.Rollback()

	series, err := updateEvent(ctx, tx, userID, rw.SeriesID, rw.Series)
	if err != nil {
		return series, nil, nil, err
	}

	var created *models.CalendarEvent
	if rw.Create != nil {
		e, err := createEvent(ctx, tx, userID, *rw.Create)
		if err != nil {
			return series, nil, nil, err
		}
		created = &e
	}

	var removedGoogleIDs []string
	if rw.ReplaceExceptions {
		rows, err := tx.QueryContext(ctx,
			`DELETE FROM calendar_events
			 WHERE user_id = $1 AND recurring_event_id = $2
			   AND ($3::timestamptz IS NULL OR original_start_at >= $3)
			 RETURNING google_event_id`,
			userID, rw.SeriesID, rw.ReplaceFrom,
		)
		if err != nil {
			return series, nil, nil, err
		}
		for rows.Next() {
			var googleID *string
			if err := rows.Scan(&googleID); err != nil {
				rows.Close()
				return series, nil, nil, err
			}
			if googleID != nil {
				removedGoogleIDs = append(removedGoogleIDs, *googleID)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return series, nil, nil, err
		}
	}

	if p := rw.Patch; p.Any() {
		_, err := tx.ExecContext(ctx,
			`UPDATE calendar_events SET
			   title = CASE WHEN $3 THEN $4 ELSE title END,
			   description = CASE WHEN $5 THEN $6 ELSE description END,
			   location = CASE WHEN $7 THEN $8 ELSE location END,
			   color = CASE WHEN $9 THEN $10 ELSE color END,
			   updated_at = NOW()
			 WHERE user_id = $1 AND recurring_event_id = $2 AND status = 'confirmed'`,
			userID, rw.SeriesID, p.SetTitle, p.Title, p.SetDescription, p.Description, p.SetLocation, p.Location, p.SetColor, p.Color,
		)
		if err != nil {
			return series, nil, nil, err
		}
	}

	parentID := rw.SeriesID
	if created != nil {
		parentID = created.ID
	}
	for _, x := range rw.Exceptions {
		f := x.Fields
		_, err := tx.ExecContext(ctx,
			`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, color, recurring_event_id, original_start_at, status, source)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			userID, f.Title, f.Description, f.Location, f.StartAt, f.EndAt, f.AllDay, f.Timezone, f.Color, parentID, x.OriginalStartAt, x.Status, x.Source,
		)
		if err != nil {
			return series, nil, nil, err
		}
	}

	return series, created, removedGoogleIDs, tx.Commit()
}

// UpsertFromGoogle is the pull side of sync for single events and series: it
// writes Google's version into calendar_events, keyed by (user_id,
// google_event_id). Marking google_synced_at = NOW() alongside the write is
// what keeps this from being picked up as "dirty" by ListDirtyForPush on the
// very next tick. color is only applied on insert — it's Arete-only (never
// pushed to Google), so an existing row keeps whatever color the user picked.
func (r *CalendarEventRepository) UpsertFromGoogle(ctx context.Context, userID, googleEventID, googleCalendarID string, f CalendarEventFields) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	// The WHERE must repeat idx_calendar_events_user_google_event's predicate,
	// or Postgres can't match the partial index to the ON CONFLICT target.
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, recurrence, recurrence_end_at, color, google_event_id, google_calendar_id, source, google_synced_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'google', NOW())
		 ON CONFLICT (user_id, google_event_id) WHERE google_event_id IS NOT NULL DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   location = EXCLUDED.location,
		   start_at = EXCLUDED.start_at,
		   end_at = EXCLUDED.end_at,
		   all_day = EXCLUDED.all_day,
		   timezone = EXCLUDED.timezone,
		   recurrence = EXCLUDED.recurrence,
		   recurrence_end_at = EXCLUDED.recurrence_end_at,
		   google_calendar_id = EXCLUDED.google_calendar_id,
		   updated_at = NOW(),
		   google_synced_at = NOW()
		 RETURNING `+calendarEventColumns,
		userID, f.Title, f.Description, f.Location, f.StartAt, f.EndAt, f.AllDay, f.Timezone, recurrenceArg(f.Recurrence), f.RecurrenceEndAt, f.Color, googleEventID, googleCalendarID,
	)
	if err := scanCalendarEvent(row, &e); err != nil {
		return e, err
	}

	// A series turned into a single event on Google's side leaves no
	// occurrences for its exceptions to replace.
	if len(f.Recurrence) == 0 {
		if _, err := r.db.ExecContext(ctx, `DELETE FROM calendar_events WHERE recurring_event_id = $1`, e.ID); err != nil {
			return e, err
		}
	}
	return e, nil
}

// UpsertExceptionFromGoogle is the pull side of sync for an exception (an
// edited or cancelled occurrence), keyed like UpsertFromGoogle. A new
// exception takes its series' color.
func (r *CalendarEventRepository) UpsertExceptionFromGoogle(ctx context.Context, userID, googleEventID, googleCalendarID, seriesID string, originalStart time.Time, f CalendarEventFields, status string) (models.CalendarEvent, error) {
	var e models.CalendarEvent
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO calendar_events (user_id, title, description, location, start_at, end_at, all_day, timezone, color, google_event_id, google_calendar_id, source, google_synced_at, recurring_event_id, original_start_at, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'google', NOW(), $12, $13, $14)
		 ON CONFLICT (user_id, google_event_id) WHERE google_event_id IS NOT NULL DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   location = EXCLUDED.location,
		   start_at = EXCLUDED.start_at,
		   end_at = EXCLUDED.end_at,
		   all_day = EXCLUDED.all_day,
		   timezone = EXCLUDED.timezone,
		   google_calendar_id = EXCLUDED.google_calendar_id,
		   recurring_event_id = EXCLUDED.recurring_event_id,
		   original_start_at = EXCLUDED.original_start_at,
		   status = EXCLUDED.status,
		   updated_at = NOW(),
		   google_synced_at = NOW()
		 RETURNING `+calendarEventColumns,
		userID, f.Title, f.Description, f.Location, f.StartAt, f.EndAt, f.AllDay, f.Timezone, f.Color, googleEventID, googleCalendarID, seriesID, originalStart, status,
	)
	err := scanCalendarEvent(row, &e)
	return e, err
}

// DeleteByGoogleEventID handles a pull-side cancellation (an event deleted on
// Google's end). A series' exceptions go with it via ON DELETE CASCADE.
func (r *CalendarEventRepository) DeleteByGoogleEventID(ctx context.Context, userID, googleEventID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM calendar_events WHERE user_id = $1 AND google_event_id = $2`,
		userID, googleEventID,
	)
	return err
}

// ListDirtyForPush returns events that either have never been pushed to
// Google, or were edited locally since their last push/pull — including
// Google-origin events, since the UI lets users edit those too. Series and
// single events come before exceptions, so a series exists on Google before
// its exceptions are pushed.
func (r *CalendarEventRepository) ListDirtyForPush(ctx context.Context, userID string) ([]models.CalendarEvent, error) {
	return listCalendarEvents(ctx, r.db,
		`SELECT `+calendarEventColumns+`
		 FROM calendar_events
		 WHERE user_id = $1
		   AND ((google_event_id IS NULL AND google_synced_at IS NULL) OR updated_at > google_synced_at)
		 ORDER BY (recurring_event_id IS NOT NULL), created_at ASC`,
		userID,
	)
}

// MarkSynced records that a local event was just pushed to Google, attaching
// the Google-assigned event ID (on first push) and resetting the dirty marker.
func (r *CalendarEventRepository) MarkSynced(ctx context.Context, id, googleEventID, googleCalendarID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE calendar_events SET google_event_id = $1, google_calendar_id = $2, google_synced_at = NOW() WHERE id = $3`,
		googleEventID, googleCalendarID, id,
	)
	return err
}

// MarkPushedWithoutRemote clears the dirty marker of an exception that has
// nothing to update on Google (its occurrence doesn't exist there).
func (r *CalendarEventRepository) MarkPushedWithoutRemote(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE calendar_events SET google_synced_at = NOW() WHERE id = $1`, id)
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
