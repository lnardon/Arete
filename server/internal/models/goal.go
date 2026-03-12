package models

import "time"

type Goal struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Title      string    `json:"title"`
	PeriodType string    `json:"periodType"` // month|quarter|semester|year
	PeriodKey  string    `json:"periodKey"`  // "2026-03", "2026-Q1", "2026-H1", "2026"
	Completed  bool      `json:"completed"`
	CreatedAt  time.Time `json:"createdAt"`
}
