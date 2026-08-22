package models

import "time"

type PomodoroProject struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
}

type PomodoroEntry struct {
	ID             string     `json:"id"`
	ProjectID      *string    `json:"projectId"`
	PlannedMinutes int        `json:"plannedMinutes"`
	StartedAt      time.Time  `json:"startedAt"`
	EndedAt        *time.Time `json:"endedAt"`
	LocalDate      string     `json:"localDate"`
	CreatedAt      time.Time  `json:"createdAt"`
}
