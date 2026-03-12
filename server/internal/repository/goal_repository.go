package repository

import (
	"context"
	"database/sql"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

type GoalRepository struct {
	db *database.DB
}

func NewGoalRepository(db *database.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) ListGoals(ctx context.Context, userID, periodType, periodKey string) ([]models.Goal, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, title, period_type, period_key, completed, created_at
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
		if err := rows.Scan(&g.ID, &g.UserID, &g.Title, &g.PeriodType, &g.PeriodKey, &g.Completed, &g.CreatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	if goals == nil {
		goals = []models.Goal{}
	}
	return goals, rows.Err()
}

func (r *GoalRepository) CreateGoal(ctx context.Context, userID, title, periodType, periodKey string) (models.Goal, error) {
	var g models.Goal
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO goals (user_id, title, period_type, period_key)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, title, period_type, period_key, completed, created_at`,
		userID, title, periodType, periodKey,
	).Scan(&g.ID, &g.UserID, &g.Title, &g.PeriodType, &g.PeriodKey, &g.Completed, &g.CreatedAt)
	return g, err
}

func (r *GoalRepository) ToggleGoal(ctx context.Context, userID, id string) (models.Goal, error) {
	var g models.Goal
	err := r.db.QueryRowContext(ctx,
		`UPDATE goals SET completed = NOT completed
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, title, period_type, period_key, completed, created_at`,
		id, userID,
	).Scan(&g.ID, &g.UserID, &g.Title, &g.PeriodType, &g.PeriodKey, &g.Completed, &g.CreatedAt)

	if err == sql.ErrNoRows {
		return g, ErrNotFound
	}
	return g, err
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
