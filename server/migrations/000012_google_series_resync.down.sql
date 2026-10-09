-- The old code pulls expanded instances again on its next full sync.
UPDATE google_oauth_tokens SET sync_token = NULL;
