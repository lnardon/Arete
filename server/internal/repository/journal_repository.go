package repository

import (
	"context"
	"database/sql"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

const journalColumns = `id, user_id, entry_date::text, mood, content, created_at, updated_at`

type JournalRepository struct {
	db *database.DB
}

func NewJournalRepository(db *database.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

func scanJournalEntry(row interface {
	Scan(dest ...any) error
}, e *models.JournalEntry) error {
	return row.Scan(&e.ID, &e.UserID, &e.EntryDate, &e.Mood, &e.Content, &e.CreatedAt, &e.UpdatedAt)
}

func (r *JournalRepository) ListEntries(ctx context.Context, userID string, limit int) ([]models.JournalEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+journalColumns+`
		 FROM journal_entries
		 WHERE user_id = $1
		 ORDER BY entry_date DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.JournalEntry
	for rows.Next() {
		var e models.JournalEntry
		if err := scanJournalEntry(rows, &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []models.JournalEntry{}
	}
	return entries, rows.Err()
}

func (r *JournalRepository) GetEntry(ctx context.Context, userID, entryDate string) (models.JournalEntry, error) {
	var e models.JournalEntry
	row := r.db.QueryRowContext(ctx,
		`SELECT `+journalColumns+`
		 FROM journal_entries
		 WHERE user_id = $1 AND entry_date = $2`,
		userID, entryDate,
	)
	err := scanJournalEntry(row, &e)
	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

// UpsertEntry creates the entry for entryDate, or overwrites its mood and
// content if one already exists — journal entries are one-per-day, so a
// second save for the same day is always an edit, never a duplicate.
func (r *JournalRepository) UpsertEntry(ctx context.Context, userID, entryDate string, mood int, content string) (models.JournalEntry, error) {
	var e models.JournalEntry
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO journal_entries (user_id, entry_date, mood, content)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, entry_date)
		 DO UPDATE SET mood = EXCLUDED.mood, content = EXCLUDED.content, updated_at = NOW()
		 RETURNING `+journalColumns,
		userID, entryDate, mood, content,
	)
	err := scanJournalEntry(row, &e)
	return e, err
}

func (r *JournalRepository) DeleteEntry(ctx context.Context, userID, entryDate string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM journal_entries WHERE user_id = $1 AND entry_date = $2`,
		userID, entryDate,
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
