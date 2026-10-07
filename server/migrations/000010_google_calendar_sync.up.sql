CREATE TABLE google_oauth_tokens (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    google_email   TEXT,
    access_token   TEXT NOT NULL,
    refresh_token  TEXT NOT NULL,
    token_expiry   TIMESTAMPTZ NOT NULL,
    scope          TEXT NOT NULL,
    calendar_id    TEXT NOT NULL DEFAULT 'primary',
    sync_token     TEXT,
    last_synced_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_google_oauth_tokens_user_id ON google_oauth_tokens(user_id);

-- Tracks when a calendar_events row was last pushed/pulled, distinct from
-- updated_at, so the sync engine can tell "user edited this" (needs push)
-- apart from "we just synced this" (does not) without an endless push/pull loop.
ALTER TABLE calendar_events ADD COLUMN google_synced_at TIMESTAMPTZ;

-- Google event IDs are only unique per calendar, not globally: invitees to the
-- same meeting see the same ID. 000009's global unique index would reject the
-- second Arete user to sync a shared meeting, so scope uniqueness per user.
DROP INDEX IF EXISTS idx_calendar_events_google_event_id;
CREATE UNIQUE INDEX idx_calendar_events_user_google_event ON calendar_events(user_id, google_event_id) WHERE google_event_id IS NOT NULL;
