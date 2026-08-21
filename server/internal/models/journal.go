package models

import "time"

type JournalEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	EntryDate string    `json:"entryDate"` // YYYY-MM-DD
	Mood      int       `json:"mood"`      // 1 (rough) - 5 (great)
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
