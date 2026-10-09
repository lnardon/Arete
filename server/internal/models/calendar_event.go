package models

import "time"

const (
	CalendarEventConfirmed = "confirmed"
	CalendarEventCancelled = "cancelled"
)

// CalendarEvent is one row of calendar_events, which holds three kinds of
// rows, mirroring Google Calendar's model:
//   - a single event: Recurrence and RecurringEventID both nil;
//   - a recurring series: Recurrence holds RFC 5545 lines, and StartAt/EndAt
//     are its first occurrence;
//   - an exception: one edited or cancelled occurrence of a series, pointing
//     at it through RecurringEventID and OriginalStartAt.
//
// The list endpoint also returns computed occurrences in this shape, with ID
// set to an instance ID (see recurrence.InstanceID).
type CalendarEvent struct {
	ID               string     `json:"id"`
	UserID           string     `json:"userId"`
	Title            string     `json:"title"`
	Description      *string    `json:"description"`
	Location         *string    `json:"location"`
	StartAt          time.Time  `json:"startAt"`
	EndAt            time.Time  `json:"endAt"`
	AllDay           bool       `json:"allDay"`
	Timezone         string     `json:"timezone"`
	Recurrence       []string   `json:"recurrence"`
	RecurrenceEndAt  *time.Time `json:"-"`
	RecurringEventID *string    `json:"recurringEventId"`
	OriginalStartAt  *time.Time `json:"originalStartAt"`
	Status           string     `json:"status"`
	Color            string     `json:"color"`
	GoogleEventID    *string    `json:"googleEventId"`
	GoogleCalendarID *string    `json:"googleCalendarId"`
	Source           string     `json:"source"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
