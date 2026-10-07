package models

import "time"

type GoogleAccount struct {
	ID           string     `json:"id"`
	UserID       string     `json:"userId"`
	GoogleEmail  *string    `json:"googleEmail"`
	AccessToken  string     `json:"-"`
	RefreshToken string     `json:"-"`
	TokenExpiry  time.Time  `json:"-"`
	Scope        string     `json:"-"`
	CalendarID   string     `json:"-"`
	SyncToken    *string    `json:"-"`
	LastSyncedAt *time.Time `json:"lastSyncedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
