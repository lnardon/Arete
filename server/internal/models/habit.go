package models

import "time"

type Habit struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type HabitCompletion struct {
	HabitID string `json:"habitId"`
	Date    string `json:"date"` // YYYY-MM-DD
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}
