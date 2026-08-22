package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

type PomodoroRepository struct {
	db *database.DB
}

func NewPomodoroRepository(db *database.DB) *PomodoroRepository {
	return &PomodoroRepository{db: db}
}

func (r *PomodoroRepository) ListProjects(ctx context.Context, userID string) ([]models.PomodoroProject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, created_at FROM pomodoro_projects WHERE user_id = $1 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.PomodoroProject
	for rows.Next() {
		var p models.PomodoroProject
		if err := rows.Scan(&p.ID, &p.Name, &p.Color, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []models.PomodoroProject{}
	}
	return projects, rows.Err()
}

func (r *PomodoroRepository) CreateProject(ctx context.Context, userID, name, color string) (models.PomodoroProject, error) {
	var p models.PomodoroProject
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO pomodoro_projects (user_id, name, color) VALUES ($1, $2, $3) RETURNING id, name, color, created_at`,
		userID, name, color,
	).Scan(&p.ID, &p.Name, &p.Color, &p.CreatedAt)
	return p, err
}

func (r *PomodoroRepository) UpdateProject(ctx context.Context, userID, id, name, color string) (models.PomodoroProject, error) {
	var p models.PomodoroProject
	err := r.db.QueryRowContext(ctx,
		`UPDATE pomodoro_projects SET name = $1, color = $2 WHERE id = $3 AND user_id = $4
		 RETURNING id, name, color, created_at`,
		name, color, id, userID,
	).Scan(&p.ID, &p.Name, &p.Color, &p.CreatedAt)

	if err == sql.ErrNoRows {
		return p, ErrNotFound
	}
	return p, err
}

func (r *PomodoroRepository) DeleteProject(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM pomodoro_projects WHERE id = $1 AND user_id = $2`,
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

func (r *PomodoroRepository) GetActiveEntry(ctx context.Context, userID string) (*models.PomodoroEntry, error) {
	var e models.PomodoroEntry
	err := r.db.QueryRowContext(ctx,
		`SELECT id, project_id, planned_minutes, started_at, ended_at, local_date::text, created_at
		 FROM pomodoro_entries WHERE user_id = $1 AND ended_at IS NULL`,
		userID,
	).Scan(&e.ID, &e.ProjectID, &e.PlannedMinutes, &e.StartedAt, &e.EndedAt, &e.LocalDate, &e.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PomodoroRepository) StartEntry(ctx context.Context, userID string, projectID *string, plannedMinutes int, localDate string) (models.PomodoroEntry, error) {
	var e models.PomodoroEntry
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO pomodoro_entries (user_id, project_id, planned_minutes, local_date)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, project_id, planned_minutes, started_at, ended_at, local_date::text, created_at`,
		userID, projectID, plannedMinutes, localDate,
	).Scan(&e.ID, &e.ProjectID, &e.PlannedMinutes, &e.StartedAt, &e.EndedAt, &e.LocalDate, &e.CreatedAt)

	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return e, ErrActiveEntryExists
	}
	return e, err
}

func (r *PomodoroRepository) StopActiveEntry(ctx context.Context, userID string) (models.PomodoroEntry, error) {
	var e models.PomodoroEntry
	err := r.db.QueryRowContext(ctx,
		`UPDATE pomodoro_entries SET ended_at = NOW() WHERE user_id = $1 AND ended_at IS NULL
		 RETURNING id, project_id, planned_minutes, started_at, ended_at, local_date::text, created_at`,
		userID,
	).Scan(&e.ID, &e.ProjectID, &e.PlannedMinutes, &e.StartedAt, &e.EndedAt, &e.LocalDate, &e.CreatedAt)

	if err == sql.ErrNoRows {
		return e, ErrNotFound
	}
	return e, err
}

func (r *PomodoroRepository) ListEntries(ctx context.Context, userID, startDate, endDate string, projectID *string) ([]models.PomodoroEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, project_id, planned_minutes, started_at, ended_at, local_date::text, created_at
		 FROM pomodoro_entries
		 WHERE user_id = $1 AND local_date >= $2 AND local_date <= $3
		   AND ($4::uuid IS NULL OR project_id = $4)
		 ORDER BY started_at DESC`,
		userID, startDate, endDate, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.PomodoroEntry
	for rows.Next() {
		var e models.PomodoroEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.PlannedMinutes, &e.StartedAt, &e.EndedAt, &e.LocalDate, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []models.PomodoroEntry{}
	}
	return entries, rows.Err()
}

func (r *PomodoroRepository) DeleteEntry(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM pomodoro_entries WHERE id = $1 AND user_id = $2`,
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
