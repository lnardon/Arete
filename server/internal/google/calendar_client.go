package google

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// ErrSyncTokenInvalid indicates Google rejected the stored sync token (410
// Gone) — the caller should clear the stored token and retry a full resync.
var ErrSyncTokenInvalid = errors.New("google sync token invalid, full resync required")

// ErrEditForbidden indicates Google refused an edit on permission grounds —
// typically an invite the user doesn't organize, where guests can't modify
// the event. Unlike a rate limit, retrying will never succeed.
var ErrEditForbidden = errors.New("google refused the edit: not permitted on this event")

type CalendarClient struct{}

func NewCalendarClient() *CalendarClient {
	return &CalendarClient{}
}

func (c *CalendarClient) service(ctx context.Context, ts oauth2.TokenSource) (*calendar.Service, error) {
	svc, err := calendar.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("build calendar service: %w", err)
	}
	return svc, nil
}

// ListEvents lists events changed since syncToken. When syncToken is empty it
// does an initial sync bounded to roughly the last 6 months through the next
// 6 months, per Google's guidance for seeding an incremental sync. Events are
// expanded to single instances (SingleEvents), so a Google-side recurring
// event becomes individual local rows rather than one row with a recurrence
// rule — a deliberate simplification for this base sync. timeZone is the
// calendar's own zone, which all-day dates are anchored to.
func (c *CalendarClient) ListEvents(ctx context.Context, ts oauth2.TokenSource, calendarID, syncToken string) (events []*calendar.Event, nextSyncToken, timeZone string, err error) {
	svc, err := c.service(ctx, ts)
	if err != nil {
		return nil, "", "", err
	}

	call := svc.Events.List(calendarID).SingleEvents(true).ShowDeleted(true)
	if syncToken != "" {
		call = call.SyncToken(syncToken)
	} else {
		now := time.Now()
		call = call.TimeMin(now.AddDate(0, -6, 0).Format(time.RFC3339)).
			TimeMax(now.AddDate(0, 6, 0).Format(time.RFC3339))
	}

	err = call.Pages(ctx, func(page *calendar.Events) error {
		events = append(events, page.Items...)
		nextSyncToken = page.NextSyncToken
		timeZone = page.TimeZone
		return nil
	})
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 410 {
			return nil, "", "", ErrSyncTokenInvalid
		}
		return nil, "", "", fmt.Errorf("list events: %w", err)
	}
	return events, nextSyncToken, timeZone, nil
}

func (c *CalendarClient) GetEvent(ctx context.Context, ts oauth2.TokenSource, calendarID, eventID string) (*calendar.Event, error) {
	svc, err := c.service(ctx, ts)
	if err != nil {
		return nil, err
	}
	ev, err := svc.Events.Get(calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return ev, nil
}

func (c *CalendarClient) InsertEvent(ctx context.Context, ts oauth2.TokenSource, calendarID string, event *calendar.Event) (*calendar.Event, error) {
	svc, err := c.service(ctx, ts)
	if err != nil {
		return nil, err
	}
	created, err := svc.Events.Insert(calendarID, event).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("insert event: %w", err)
	}
	return created, nil
}

func (c *CalendarClient) UpdateEvent(ctx context.Context, ts oauth2.TokenSource, calendarID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	svc, err := c.service(ctx, ts)
	if err != nil {
		return nil, err
	}
	updated, err := svc.Events.Update(calendarID, eventID, event).Context(ctx).Do()
	if err != nil {
		if isPermissionDenied(err) {
			return nil, ErrEditForbidden
		}
		return nil, fmt.Errorf("update event: %w", err)
	}
	return updated, nil
}

// isPermissionDenied reports whether err is a 403 that retrying won't fix. The
// Calendar API also uses 403 for rate and quota limits, which are worth
// retrying on the next tick, so those are excluded.
func isPermissionDenied(err error) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) || apiErr.Code != 403 {
		return false
	}
	for _, item := range apiErr.Errors {
		switch item.Reason {
		case "rateLimitExceeded", "userRateLimitExceeded", "quotaExceeded", "dailyLimitExceeded":
			return false
		}
	}
	return true
}

func (c *CalendarClient) DeleteEvent(ctx context.Context, ts oauth2.TokenSource, calendarID, eventID string) error {
	svc, err := c.service(ctx, ts)
	if err != nil {
		return err
	}
	if err := svc.Events.Delete(calendarID, eventID).Context(ctx).Do(); err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 410 {
			return nil // already gone on Google's side — not an error for us
		}
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}
