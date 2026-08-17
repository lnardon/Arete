package models

import "time"

type WhatsAppLink struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	PhoneNumber string    `json:"phoneNumber"`
	LinkedAt    time.Time `json:"linkedAt"`
}

type WhatsAppLinkCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}
