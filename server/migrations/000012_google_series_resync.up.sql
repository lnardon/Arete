-- Google pull moves from expanded instances (singleEvents=true) to series +
-- exceptions. The old per-instance rows have Google IDs ending in _YYYYMMDD or
-- _YYYYMMDDTHHMMSSZ; the full resync below re-creates them as series.
DELETE FROM calendar_events
 WHERE source = 'google'
   AND google_event_id ~ '_[0-9]{8}(T[0-9]{6}Z)?$';

-- A sync token is bound to the parameters of its initial request, so changing
-- singleEvents requires a fresh full sync.
UPDATE google_oauth_tokens SET sync_token = NULL;
