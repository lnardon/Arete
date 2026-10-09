-- Series rules are stored as RFC 5545 lines, exactly like Google's
-- Event.recurrence, so push/pull is a copy rather than a translation.
ALTER TABLE calendar_events ADD COLUMN recurrence TEXT[];
UPDATE calendar_events
   SET recurrence = ARRAY[CASE WHEN recurrence_rule LIKE 'RRULE:%' THEN recurrence_rule
                               ELSE 'RRULE:' || recurrence_rule END]
 WHERE recurrence_rule IS NOT NULL AND recurrence_rule <> '';
ALTER TABLE calendar_events DROP COLUMN recurrence_rule;

-- End of the series' last occurrence, for range queries. NULL = repeats forever.
ALTER TABLE calendar_events ADD COLUMN recurrence_end_at TIMESTAMPTZ;

-- An exception overrides one occurrence of a series (Google's
-- recurringEventId + originalStartTime). Cancelled exceptions hide an
-- occurrence; Google says to keep them for the life of the series.
ALTER TABLE calendar_events
    ADD COLUMN recurring_event_id UUID REFERENCES calendar_events(id) ON DELETE CASCADE,
    ADD COLUMN original_start_at  TIMESTAMPTZ,
    ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'confirmed'
        CHECK (status IN ('confirmed', 'cancelled')),
    ADD CONSTRAINT calendar_events_exception_shape
        CHECK ((recurring_event_id IS NULL) = (original_start_at IS NULL)),
    ADD CONSTRAINT calendar_events_exception_has_no_rule
        CHECK (recurring_event_id IS NULL OR recurrence IS NULL);

CREATE UNIQUE INDEX idx_calendar_events_exception
    ON calendar_events(recurring_event_id, original_start_at)
    WHERE recurring_event_id IS NOT NULL;
CREATE INDEX idx_calendar_events_user_series
    ON calendar_events(user_id, start_at)
    WHERE recurrence IS NOT NULL;
