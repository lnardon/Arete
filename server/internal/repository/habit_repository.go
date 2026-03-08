package repository

import (
	"context"
	"database/sql"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

type HabitRepository struct {
	db *database.DB
}

func NewHabitRepository(db *database.DB) *HabitRepository {
	return &HabitRepository{db: db}
}

func (r *HabitRepository) ListHabits(ctx context.Context) ([]models.Habit, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at FROM habits ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []models.Habit
	for rows.Next() {
		var h models.Habit
		if err := rows.Scan(&h.ID, &h.Name, &h.CreatedAt); err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}
	if habits == nil {
		habits = []models.Habit{}
	}
	return habits, rows.Err()
}

func (r *HabitRepository) CreateHabit(ctx context.Context, name string) (models.Habit, error) {
	var h models.Habit
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO habits (name) VALUES ($1) RETURNING id, name, created_at`,
		name,
	).Scan(&h.ID, &h.Name, &h.CreatedAt)
	return h, err
}

func (r *HabitRepository) UpdateHabit(ctx context.Context, id string, name string) (models.Habit, error) {
	var h models.Habit
	err := r.db.QueryRowContext(ctx,
		`UPDATE habits SET name = $1 WHERE id = $2 RETURNING id, name, created_at`,
		name, id,
	).Scan(&h.ID, &h.Name, &h.CreatedAt)

	if err == sql.ErrNoRows {
		return h, ErrNotFound
	}
	return h, err
}

func (r *HabitRepository) DeleteHabit(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM habits WHERE id = $1`, id)
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

func (r *HabitRepository) GetCompletionsForDate(ctx context.Context, date string) ([]models.HabitCompletion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT habit_id, date::text FROM habit_completions WHERE date = $1`,
		date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var completions []models.HabitCompletion
	for rows.Next() {
		var c models.HabitCompletion
		if err := rows.Scan(&c.HabitID, &c.Date); err != nil {
			return nil, err
		}
		completions = append(completions, c)
	}
	if completions == nil {
		completions = []models.HabitCompletion{}
	}
	return completions, rows.Err()
}

func (r *HabitRepository) GetCompletionsForRange(ctx context.Context, startDate, endDate string) ([]models.HabitCompletion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT habit_id, date::text FROM habit_completions WHERE date >= $1 AND date <= $2 ORDER BY date ASC`,
		startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var completions []models.HabitCompletion
	for rows.Next() {
		var c models.HabitCompletion
		if err := rows.Scan(&c.HabitID, &c.Date); err != nil {
			return nil, err
		}
		completions = append(completions, c)
	}
	if completions == nil {
		completions = []models.HabitCompletion{}
	}
	return completions, rows.Err()
}

func (r *HabitRepository) ToggleCompletion(ctx context.Context, habitID string, date string) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO habit_completions (habit_id, date) VALUES ($1, $2) ON CONFLICT (habit_id, date) DO NOTHING`,
		habitID, date,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		_, err = r.db.ExecContext(ctx,
			`DELETE FROM habit_completions WHERE habit_id = $1 AND date = $2`,
			habitID, date,
		)
		return err
	}
	return nil
}
