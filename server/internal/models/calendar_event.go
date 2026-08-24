package models

import "time"

type CalendarEvent struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	Title            string    `json:"title"`
	Description      *string   `json:"description"`
	Location         *string   `json:"location"`
	StartAt          time.Time `json:"startAt"`
	EndAt            time.Time `json:"endAt"`
	AllDay           bool      `json:"allDay"`
	Timezone         string    `json:"timezone"`
	RecurrenceRule   *string   `json:"recurrenceRule"`
	Color            string    `json:"color"`
	GoogleEventID    *string   `json:"googleEventId"`
	GoogleCalendarID *string   `json:"googleCalendarId"`
	Source           string    `json:"source"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
