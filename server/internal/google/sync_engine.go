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
	"github.com/lnardon/arete/internal/recurrence"
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

// PushUser runs only the push half of a sync pass for one user. Scoped edits
// to a recurring series call it right away: a regular tick pulls before it
// pushes, and pulling Google's stale copy of the series between the local
// edit and its push could re-apply exceptions to the wrong occurrences.
func (s *SyncEngine) PushUser(ctx context.Context, userID string) error {
	account, err := s.accounts.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil // not connected — nothing to push
	}
	if err != nil {
		return err
	}
	defer s.lockUser(userID)()
	return s.pushDirty(ctx, s.tokenSource(ctx, account), account)
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

	// Exceptions (edited or cancelled occurrences) are applied after
	// everything else: one can arrive on an earlier page than its series.
	calendarLoc := s.location(calendarTZ)
	var exceptions []*calendar.Event
	for _, ev := range events {
		if ev.RecurringEventId != "" {
			exceptions = append(exceptions, ev)
			continue
		}
		if ev.Status == "cancelled" {
			if err := s.events.DeleteByGoogleEventID(ctx, account.UserID, ev.Id); err != nil {
				slog.Error("google sync: delete pulled event failed", "userId", account.UserID, "eventId", ev.Id, "error", err)
			}
			continue
		}
		if _, err := s.applyRemoteEvent(ctx, account, ev, calendarLoc); err != nil {
			slog.Error("google sync: apply pulled event failed", "userId", account.UserID, "eventId", ev.Id, "error", err)
		}
	}
	for _, ev := range exceptions {
		if err := s.applyRemoteException(ctx, ts, account, ev, calendarLoc); err != nil {
			slog.Error("google sync: apply pulled exception failed", "userId", account.UserID, "eventId", ev.Id, "error", err)
		}
	}

	if err := s.pushDirty(ctx, ts, account); err != nil {
		return err
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

// pushDirty sends every locally created or edited event to Google.
func (s *SyncEngine) pushDirty(ctx context.Context, ts oauth2.TokenSource, account models.GoogleAccount) error {
	dirty, err := s.events.ListDirtyForPush(ctx, account.UserID)
	if err != nil {
		return fmt.Errorf("list dirty events: %w", err)
	}
	for _, e := range dirty {
		if err := s.pushEvent(ctx, ts, account, e); err != nil {
			slog.Error("google sync: push event failed", "userId", account.UserID, "eventId", e.ID, "error", err)
		}
	}
	return nil
}

// applyRemoteEvent writes Google's version of one single event or series into
// the local calendar — the pull side of sync, also used to revert a push
// Google refused.
func (s *SyncEngine) applyRemoteEvent(ctx context.Context, account models.GoogleAccount, ev *calendar.Event, loc *time.Location) (models.CalendarEvent, error) {
	fields, err := googleEventToLocal(ev, loc)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("parse event: %w", err)
	}
	fields.Color = defaultEventColor
	if len(ev.Recurrence) > 0 {
		fields.Recurrence = ev.Recurrence
		if fields.RecurrenceEndAt, err = recurrence.End(fields.Event()); err != nil {
			return models.CalendarEvent{}, fmt.Errorf("expand recurrence: %w", err)
		}
	}
	return s.events.UpsertFromGoogle(ctx, account.UserID, ev.Id, account.CalendarID, fields)
}

// applyRemoteException writes one edited or cancelled occurrence of a Google
// series into the local calendar, attached to the local copy of its series.
func (s *SyncEngine) applyRemoteException(ctx context.Context, ts oauth2.TokenSource, account models.GoogleAccount, ev *calendar.Event, loc *time.Location) error {
	series, err := s.events.GetByGoogleEventID(ctx, account.UserID, ev.RecurringEventId)
	if errors.Is(err, repository.ErrNotFound) {
		// The series wasn't in this pull (unchanged, or outside the initial
		// sync window), so fetch it before attaching the exception.
		master, err := s.calendar.GetEvent(ctx, ts, account.CalendarID, ev.RecurringEventId)
		if err != nil {
			return fmt.Errorf("fetch series: %w", err)
		}
		if master.Status == "cancelled" {
			return nil
		}
		if series, err = s.applyRemoteEvent(ctx, account, master, loc); err != nil {
			return fmt.Errorf("apply series: %w", err)
		}
	} else if err != nil {
		return err
	}
	if len(series.Recurrence) == 0 {
		return nil // no longer recurring; nothing for the exception to replace
	}

	seriesLoc := s.location(series.Timezone)
	originalStart, err := googleOriginalStart(ev.OriginalStartTime, seriesLoc)
	if err != nil {
		return fmt.Errorf("parse original start: %w", err)
	}

	var fields repository.CalendarEventFields
	status := models.CalendarEventConfirmed
	if ev.Status == "cancelled" {
		// A cancelled exception only guarantees id, recurringEventId and
		// originalStartTime. One whose slot isn't an occurrence of the series
		// any more hides nothing, so it isn't worth keeping.
		if !recurrence.IsOccurrence(series, originalStart) {
			return nil
		}
		status = models.CalendarEventCancelled
		fields = repository.FieldsOf(recurrence.Occurrence(series, originalStart))
		fields.Recurrence, fields.RecurrenceEndAt = nil, nil
	} else {
		if fields, err = googleEventToLocal(ev, seriesLoc); err != nil {
			return fmt.Errorf("parse event: %w", err)
		}
		fields.Color = series.Color
	}

	_, err = s.events.UpsertExceptionFromGoogle(ctx, account.UserID, ev.Id, account.CalendarID, series.ID, originalStart, fields, status)
	return err
}

// googleOriginalStart reads an instance's originalStartTime: a date for
// all-day series (anchored to loc, like their start), a date-time otherwise.
func googleOriginalStart(original *calendar.EventDateTime, loc *time.Location) (time.Time, error) {
	if original == nil {
		return time.Time{}, errors.New("missing originalStartTime")
	}
	if original.Date != "" {
		return time.ParseInLocation("2006-01-02", original.Date, loc)
	}
	return time.Parse(time.RFC3339, original.DateTime)
}

// pushEvent sends one locally created or edited event to Google and marks it
// synced.
func (s *SyncEngine) pushEvent(ctx context.Context, ts oauth2.TokenSource, account models.GoogleAccount, e models.CalendarEvent) error {
	if e.RecurringEventID != nil {
		return s.pushException(ctx, ts, account, e)
	}
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
		_, err := s.applyRemoteEvent(ctx, account, current, loc)
		return err
	}
	if err != nil {
		return err
	}
	return s.events.MarkSynced(ctx, e.ID, result.Id, account.CalendarID)
}

// pushException sends one edited or cancelled occurrence to Google as an edit
// of the matching instance of its series.
func (s *SyncEngine) pushException(ctx context.Context, ts oauth2.TokenSource, account models.GoogleAccount, e models.CalendarEvent) error {
	series, err := s.events.GetEvent(ctx, account.UserID, *e.RecurringEventID)
	if err != nil {
		return fmt.Errorf("load series: %w", err)
	}
	if series.GoogleEventID == nil {
		return nil // the series isn't on Google yet; it goes first, next tick
	}

	var instance *calendar.Event
	if e.GoogleEventID != nil {
		instance, err = s.calendar.GetEvent(ctx, ts, account.CalendarID, *e.GoogleEventID)
		if isNotFound(err) {
			err = ErrInstanceNotFound
		}
	} else {
		instance, err = s.calendar.GetInstance(ctx, ts, account.CalendarID, *series.GoogleEventID, *e.OriginalStartAt, series.AllDay)
	}
	if errors.Is(err, ErrInstanceNotFound) {
		// Google has no occurrence there, so there's nothing to change.
		return s.events.MarkPushedWithoutRemote(ctx, e.ID)
	}
	if err != nil {
		return err
	}

	if e.Status == models.CalendarEventCancelled {
		if err := s.calendar.DeleteEvent(ctx, ts, account.CalendarID, instance.Id); err != nil {
			return err
		}
		return s.events.MarkSynced(ctx, e.ID, instance.Id, account.CalendarID)
	}

	loc := s.location(e.Timezone)
	edited := *instance
	applyLocalFields(&edited, e, loc)
	result, err := s.calendar.UpdateEvent(ctx, ts, account.CalendarID, instance.Id, &edited)
	if errors.Is(err, ErrEditForbidden) {
		slog.Warn("google sync: edit not permitted on google, reverting local copy", "userId", account.UserID, "eventId", e.ID)
		return s.applyRemoteException(ctx, ts, account, instance, s.location(series.Timezone))
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
func googleEventToLocal(ev *calendar.Event, loc *time.Location) (f repository.CalendarEventFields, err error) {
	f.Title = ev.Summary
	if f.Title == "" {
		f.Title = "(untitled event)"
	}
	if ev.Description != "" {
		d := ev.Description
		f.Description = &d
	}
	if ev.Location != "" {
		l := ev.Location
		f.Location = &l
	}

	if ev.Start.Date != "" {
		f.AllDay = true
		f.Timezone = loc.String()
		if f.StartAt, err = time.ParseInLocation("2006-01-02", ev.Start.Date, loc); err != nil {
			return
		}
		f.EndAt, err = time.ParseInLocation("2006-01-02", ev.End.Date, loc)
		return
	}

	if f.StartAt, err = time.Parse(time.RFC3339, ev.Start.DateTime); err != nil {
		return
	}
	f.EndAt, err = time.Parse(time.RFC3339, ev.End.DateTime)
	// Google requires timeZone on recurring events: it's the zone the rule
	// expands in, so it has to be the one Arete expands in too.
	f.Timezone = ev.Start.TimeZone
	if f.Timezone == "" {
		f.Timezone = loc.String()
	}
	return
}

// applyLocalFields copies the fields Arete owns onto a Google event, assigning
// fresh Start/End values rather than mutating the existing ones (pushEvent
// relies on that to keep an untouched copy for reverts). All-day dates are
// read in loc for the same reason googleEventToLocal parses them in it.
// Recurrence is only set on series; an instance never carries one.
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
		// Google requires timeZone on every recurring event, all-day included.
		if len(e.Recurrence) > 0 {
			ev.Start.TimeZone, ev.End.TimeZone = e.Timezone, e.Timezone
		}
	} else {
		ev.Start = &calendar.EventDateTime{DateTime: e.StartAt.Format(time.RFC3339), TimeZone: e.Timezone}
		ev.End = &calendar.EventDateTime{DateTime: e.EndAt.Format(time.RFC3339), TimeZone: e.Timezone}
	}

	switch {
	case len(e.Recurrence) > 0:
		ev.Recurrence = e.Recurrence
	case len(ev.Recurrence) > 0 && e.RecurringEventID == nil:
		// A series turned back into a single event. The client library
		// omits empty fields, so the removal has to be sent as an explicit null.
		ev.Recurrence = nil
		ev.NullFields = append(ev.NullFields, "Recurrence")
	}
}
