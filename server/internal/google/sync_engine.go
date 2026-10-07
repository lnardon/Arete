package google

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/oauth2"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/lnardon/arete/internal/models"
	"github.com/lnardon/arete/internal/repository"
)

const defaultEventColor = "#6366f1"

type SyncEngine struct {
	accounts   *repository.GoogleAccountRepository
	events     *repository.CalendarEventRepository
	oauth      *OAuthService
	calendar   *CalendarClient
	interval   time.Duration
	defaultLoc *time.Location // used when neither Google nor the event names a usable zone

	// userLocks serializes sync passes per user (userID -> *sync.Mutex). The
	// ticker, the post-connect sync and "Sync now" can otherwise overlap, and
	// two concurrent passes would each insert the same not-yet-pushed event
	// into Google, duplicating it.
	userLocks sync.Map
}

func NewSyncEngine(accounts *repository.GoogleAccountRepository, events *repository.CalendarEventRepository, oauth *OAuthService, calendarClient *CalendarClient, interval time.Duration, defaultLoc *time.Location) *SyncEngine {
	return &SyncEngine{accounts: accounts, events: events, oauth: oauth, calendar: calendarClient, interval: interval, defaultLoc: defaultLoc}
}

// Start launches the periodic sync loop. Mirrors middleware.RateLimiter's
// ticker-based cleanup goroutine — the codebase's only prior "wake up
// periodically" precedent — repurposed for a multi-user batch job.
func (s *SyncEngine) Start() {
	go s.loop()
}

func (s *SyncEngine) loop() {
	ctx := context.Background()
	s.runOnce(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for range ticker.C {
		s.runOnce(ctx)
	}
}

func (s *SyncEngine) runOnce(ctx context.Context) {
	accounts, err := s.accounts.ListAllConnected(ctx)
	if err != nil {
		slog.Error("google sync: list accounts failed", "error", err)
		return
	}
	for _, account := range accounts {
		if err := s.syncAccount(ctx, account); err != nil {
			slog.Error("google sync: account sync failed", "userId", account.UserID, "error", err)
		}
	}
}

// SyncUser runs one sync pass for a single user — used right after a fresh
// OAuth connect and by the manual "sync now" endpoint, so a user doesn't have
// to wait for the next tick to see results.
func (s *SyncEngine) SyncUser(ctx context.Context, userID string) error {
	account, err := s.accounts.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.syncAccount(ctx, account)
}

// DeleteRemoteEvent deletes a single event on Google's side. Called
// synchronously when a user deletes a Google-linked event locally — this
// can't wait for the next sync tick, or the next pull would upsert-by-
// google_event_id and resurrect the event locally.
func (s *SyncEngine) DeleteRemoteEvent(ctx context.Context, userID, googleEventID string) error {
	account, err := s.accounts.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil // not connected — nothing to delete remotely
	}
	if err != nil {
		return err
	}
	return s.calendar.DeleteEvent(ctx, s.tokenSource(ctx, account), account.CalendarID, googleEventID)
}

func (s *SyncEngine) tokenSource(ctx context.Context, account models.GoogleAccount) oauth2.TokenSource {
	tok := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		Expiry:       account.TokenExpiry,
	}
	return s.oauth.TokenSource(ctx, tok)
}

func (s *SyncEngine) lockUser(userID string) (unlock func()) {
	m, _ := s.userLocks.LoadOrStore(userID, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// location resolves an IANA zone name, falling back to the app's default zone
// when it's empty or unknown.
func (s *SyncEngine) location(name string) *time.Location {
	if name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc
		}
	}
	return s.defaultLoc
}

func (s *SyncEngine) syncAccount(ctx context.Context, account models.GoogleAccount) error {
	defer s.lockUser(account.UserID)()

	ts := s.tokenSource(ctx, account)

	// .Token() only actually refreshes when the current token is expired or
	// near expiry; persist it if it rotated.
	refreshed, err := ts.Token()
	if err != nil {
		return fmt.Errorf("refresh token: %w", err)
	}
	if refreshed.AccessToken != account.AccessToken {
		if err := s.accounts.UpdateTokens(ctx, account.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry); err != nil {
			slog.Error("google sync: persist refreshed token failed", "userId", account.UserID, "error", err)
		}
	}

	syncToken := ""
	if account.SyncToken != nil {
		syncToken = *account.SyncToken
	}

	events, nextSyncToken, calendarTZ, err := s.calendar.ListEvents(ctx, ts, account.CalendarID, syncToken)
	if errors.Is(err, ErrSyncTokenInvalid) {
		// Google invalidated our cursor — clear it so the next tick does a
		// full resync instead of erroring forever.
		if err := s.accounts.UpdateSyncState(ctx, account.UserID, nil, time.Now()); err != nil {
			slog.Error("google sync: clear invalid sync token failed", "userId", account.UserID, "error", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("pull events: %w", err)
	}

	calendarLoc := s.location(calendarTZ)
	for _, ev := range events {
		if ev.Status == "cancelled" {
			if err := s.events.DeleteByGoogleEventID(ctx, account.UserID, ev.Id); err != nil {
				slog.Error("google sync: delete pulled event failed", "userId", account.UserID, "eventId", ev.Id, "error", err)
			}
			continue
		}
		if err := s.applyRemoteEvent(ctx, account, ev, calendarLoc); err != nil {
			slog.Error("google sync: apply pulled event failed", "userId", account.UserID, "eventId", ev.Id, "error", err)
		}
	}

	dirty, err := s.events.ListDirtyForPush(ctx, account.UserID)
	if err != nil {
		return fmt.Errorf("list dirty events: %w", err)
	}
	for _, e := range dirty {
		if err := s.pushEvent(ctx, ts, account, e); err != nil {
			slog.Error("google sync: push event failed", "userId", account.UserID, "eventId", e.ID, "error", err)
		}
	}

	var newSyncToken *string
	if nextSyncToken != "" {
		newSyncToken = &nextSyncToken
	}
	if err := s.accounts.UpdateSyncState(ctx, account.UserID, newSyncToken, time.Now()); err != nil {
		slog.Error("google sync: update sync state failed", "userId", account.UserID, "error", err)
	}

	return nil
}

// applyRemoteEvent writes Google's version of one event into the local
// calendar — the pull side of sync, also used to revert a push Google refused.
func (s *SyncEngine) applyRemoteEvent(ctx context.Context, account models.GoogleAccount, ev *calendar.Event, loc *time.Location) error {
	title, description, location, startAt, endAt, allDay, timezone, err := googleEventToLocal(ev, loc)
	if err != nil {
		return fmt.Errorf("parse event: %w", err)
	}
	_, err = s.events.UpsertFromGoogle(ctx, account.UserID, ev.Id, account.CalendarID, title, description, location, startAt, endAt, allDay, timezone, defaultEventColor)
	return err
}

// pushEvent sends one locally created or edited event to Google and marks it
// synced.
func (s *SyncEngine) pushEvent(ctx context.Context, ts oauth2.TokenSource, account models.GoogleAccount, e models.CalendarEvent) error {
	loc := s.location(e.Timezone)

	if e.GoogleEventID == nil {
		created := &calendar.Event{}
		applyLocalFields(created, e, loc)
		result, err := s.calendar.InsertEvent(ctx, ts, account.CalendarID, created)
		if err != nil {
			return err
		}
		return s.events.MarkSynced(ctx, e.ID, result.Id, account.CalendarID)
	}

	// Update replaces the whole resource, so start from Google's current copy
	// and overwrite only the fields Arete edits — otherwise attendees, Meet
	// links and reminders that exist only on Google's side would be wiped.
	current, err := s.calendar.GetEvent(ctx, ts, account.CalendarID, *e.GoogleEventID)
	if err != nil {
		return err
	}
	edited := *current
	applyLocalFields(&edited, e, loc)

	result, err := s.calendar.UpdateEvent(ctx, ts, account.CalendarID, *e.GoogleEventID, &edited)
	if errors.Is(err, ErrEditForbidden) {
		// Retrying every tick would never succeed, so put Google's version
		// back locally rather than let the two copies silently diverge.
		slog.Warn("google sync: edit not permitted on google, reverting local copy", "userId", account.UserID, "eventId", e.ID)
		return s.applyRemoteEvent(ctx, account, current, loc)
	}
	if err != nil {
		return err
	}
	return s.events.MarkSynced(ctx, e.ID, result.Id, account.CalendarID)
}

// googleEventToLocal maps a Google event onto our schema's fields. All-day
// events use EventDateTime.Date with an exclusive end date, matching this
// app's convention, but the app stores them as midnight in the user's zone
// (see calendar-event-dialog.tsx) — so the dates are anchored to loc, not UTC,
// or they'd straddle two days in the UI.
func googleEventToLocal(ev *calendar.Event, loc *time.Location) (title string, description, location *string, startAt, endAt time.Time, allDay bool, timezone string, err error) {
	title = ev.Summary
	if title == "" {
		title = "(untitled event)"
	}
	if ev.Description != "" {
		d := ev.Description
		description = &d
	}
	if ev.Location != "" {
		l := ev.Location
		location = &l
	}

	if ev.Start.Date != "" {
		allDay = true
		timezone = loc.String()
		if startAt, err = time.ParseInLocation("2006-01-02", ev.Start.Date, loc); err != nil {
			return
		}
		endAt, err = time.ParseInLocation("2006-01-02", ev.End.Date, loc)
		return
	}

	if startAt, err = time.Parse(time.RFC3339, ev.Start.DateTime); err != nil {
		return
	}
	endAt, err = time.Parse(time.RFC3339, ev.End.DateTime)
	timezone = ev.Start.TimeZone
	if timezone == "" {
		timezone = loc.String()
	}
	return
}

// applyLocalFields copies the fields Arete owns onto a Google event, assigning
// fresh Start/End values rather than mutating the existing ones (pushEvent
// relies on that to keep an untouched copy for reverts). All-day dates are
// read in loc for the same reason googleEventToLocal parses them in it.
func applyLocalFields(ev *calendar.Event, e models.CalendarEvent, loc *time.Location) {
	ev.Summary = e.Title
	ev.Description = ""
	if e.Description != nil {
		ev.Description = *e.Description
	}
	ev.Location = ""
	if e.Location != nil {
		ev.Location = *e.Location
	}

	if e.AllDay {
		ev.Start = &calendar.EventDateTime{Date: e.StartAt.In(loc).Format("2006-01-02")}
		ev.End = &calendar.EventDateTime{Date: e.EndAt.In(loc).Format("2006-01-02")}
	} else {
		ev.Start = &calendar.EventDateTime{DateTime: e.StartAt.Format(time.RFC3339), TimeZone: e.Timezone}
		ev.End = &calendar.EventDateTime{DateTime: e.EndAt.Format(time.RFC3339), TimeZone: e.Timezone}
	}
}
