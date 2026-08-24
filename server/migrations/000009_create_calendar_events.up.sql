CREATE TABLE calendar_events (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title              VARCHAR(255) NOT NULL,
    description        TEXT,
    location           VARCHAR(255),
    start_at           TIMESTAMPTZ NOT NULL,
    end_at             TIMESTAMPTZ NOT NULL,
    all_day            BOOLEAN NOT NULL DEFAULT false,
    timezone           VARCHAR(64) NOT NULL,
    recurrence_rule    TEXT,
    color              VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    -- google_event_id / google_calendar_id / source are reserved for a future
    -- Google Calendar sync integration; unused until that feature is built.
    google_event_id    VARCHAR(255),
    google_calendar_id VARCHAR(255),
    source              VARCHAR(10) NOT NULL DEFAULT 'app' CHECK (source IN ('app', 'google')),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_at > start_at)
);
CREATE INDEX idx_calendar_events_user_id ON calendar_events(user_id);
CREATE INDEX idx_calendar_events_user_start ON calendar_events(user_id, start_at);
CREATE UNIQUE INDEX idx_calendar_events_google_event_id ON calendar_events(google_event_id) WHERE google_event_id IS NOT NULL;
