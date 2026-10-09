DROP INDEX IF EXISTS idx_calendar_events_user_series;
DROP INDEX IF EXISTS idx_calendar_events_exception;
DELETE FROM calendar_events WHERE recurring_event_id IS NOT NULL;
ALTER TABLE calendar_events ADD COLUMN recurrence_rule TEXT;
UPDATE calendar_events SET recurrence_rule = substring(recurrence[1] from 7) WHERE recurrence IS NOT NULL;
ALTER TABLE calendar_events
    DROP CONSTRAINT IF EXISTS calendar_events_exception_has_no_rule,
    DROP CONSTRAINT IF EXISTS calendar_events_exception_shape,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS original_start_at,
    DROP COLUMN IF EXISTS recurring_event_id,
    DROP COLUMN IF EXISTS recurrence_end_at,
    DROP COLUMN IF EXISTS recurrence;
