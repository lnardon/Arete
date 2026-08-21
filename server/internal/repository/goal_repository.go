package repository

import (
	"context"
	"database/sql"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

const goalColumns = `id, user_id, title, period_type, period_key, goal_type, target_value, current_value, completed, created_at`

type GoalRepository struct {
	db *database.DB
}

func NewGoalRepository(db *database.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func scanGoal(row interface {
	Scan(dest ...any) error
}, g *models.Goal) error {
	return row.Scan(&g.ID, &g.UserID, &g.Title, &g.PeriodType, &g.PeriodKey, &g.GoalType, &g.TargetValue, &g.CurrentValue, &g.Completed, &g.CreatedAt)
}

func (r *GoalRepository) ListGoals(ctx context.Context, userID, periodType, periodKey string) ([]models.Goal, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+goalColumns+`
		 FROM goals
		 WHERE user_id = $1 AND period_type = $2 AND period_key = $3
		 ORDER BY created_at ASC`,
		userID, periodType, periodKey,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.Goal
	for rows.Next() {
		var g models.Goal
		if err := scanGoal(rows, &g); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	if goals == nil {
		goals = []models.Goal{}
	}
	return goals, rows.Err()
}

func (r *GoalRepository) CreateGoal(ctx context.Context, userID, title, periodType, periodKey, goalType string, targetValue *int) (models.Goal, error) {
	var g models.Goal
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO goals (user_id, title, period_type, period_key, goal_type, target_value)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+goalColumns,
		userID, title, periodType, periodKey, goalType, targetValue,
	)
	err := scanGoal(row, &g)
	return g, err
}

func (r *GoalRepository) ToggleGoal(ctx context.Context, userID, id string) (models.Goal, error) {
	var g models.Goal
	row := r.db.QueryRowContext(ctx,
		`UPDATE goals SET completed = NOT completed
		 WHERE id = $1 AND user_id = $2 AND goal_type = 'binary'
		 RETURNING `+goalColumns,
		id, userID,
	)
	err := scanGoal(row, &g)

	if err == sql.ErrNoRows {
		return g, ErrNotFound
	}
	return g, err
}

// UpdateGoal renames a goal and, for numeric goals, optionally corrects its
// target and/or current progress value. targetValue/currentValue must both
// be nil for binary goals — passing either returns ErrInvalidGoalType.
func (r *GoalRepository) UpdateGoal(ctx context.Context, userID, id, title string, targetValue, currentValue *int) (models.Goal, error) {
	var g models.Goal

	if targetValue == nil && currentValue == nil {
		row := r.db.QueryRowContext(ctx,
			`UPDATE goals SET title = $1 WHERE id = $2 AND user_id = $3
			 RETURNING `+goalColumns,
			title, id, userID,
		)
		err := scanGoal(row, &g)
		if err == sql.ErrNoRows {
			return g, ErrNotFound
		}
		return g, err
	}

	row := r.db.QueryRowContext(ctx,
		`UPDATE goals
		 SET title = $1,
		     target_value = COALESCE($2, target_value),
		     current_value = COALESCE($3, current_value),
		     completed = (COALESCE($3, current_value) >= COALESCE($2, target_value))
		 WHERE id = $4 AND user_id = $5 AND goal_type = 'numeric'
		 RETURNING `+goalColumns,
		title, targetValue, currentValue, id, userID,
	)
	err := scanGoal(row, &g)
	if err == sql.ErrNoRows {
		exists, existsErr := r.exists(ctx, userID, id)
		if existsErr != nil {
			return g, existsErr
		}
		if exists {
			return g, ErrInvalidGoalType
		}
		return g, ErrNotFound
	}
	return g, err
}

// AddGoalProgress atomically adjusts a numeric goal's current value by delta
// (which may be negative), clamped at zero, and recomputes Completed.
func (r *GoalRepository) AddGoalProgress(ctx context.Context, userID, id string, delta int) (models.Goal, error) {
	var g models.Goal
	row := r.db.QueryRowContext(ctx,
		`UPDATE goals
		 SET current_value = GREATEST(0, current_value + $1),
		     completed = (GREATEST(0, current_value + $1) >= target_value)
		 WHERE id = $2 AND user_id = $3 AND goal_type = 'numeric'
		 RETURNING `+goalColumns,
		delta, id, userID,
	)
	err := scanGoal(row, &g)
	if err == sql.ErrNoRows {
		return g, ErrNotFound
	}
	return g, err
}

func (r *GoalRepository) exists(ctx context.Context, userID, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM goals WHERE id = $1 AND user_id = $2)`,
		id, userID,
	).Scan(&exists)
	return exists, err
}

func (r *GoalRepository) DeleteGoal(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM goals WHERE id = $1 AND user_id = $2`,
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
