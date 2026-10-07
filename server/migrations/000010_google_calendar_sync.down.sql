DROP INDEX IF EXISTS idx_calendar_events_user_google_event;
CREATE UNIQUE INDEX IF NOT EXISTS idx_calendar_events_google_event_id ON calendar_events(google_event_id) WHERE google_event_id IS NOT NULL;
ALTER TABLE calendar_events DROP COLUMN IF EXISTS google_synced_at;
DROP TABLE IF EXISTS google_oauth_tokens;
