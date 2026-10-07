package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
	"github.com/lnardon/arete/internal/tokencrypt"
)

const googleAccountColumns = `id, user_id, google_email, access_token, refresh_token, token_expiry, scope, calendar_id, sync_token, last_synced_at, created_at, updated_at`

type GoogleAccountRepository struct {
	db     *database.DB
	cipher *tokencrypt.Cipher
}

func NewGoogleAccountRepository(db *database.DB, cipher *tokencrypt.Cipher) *GoogleAccountRepository {
	return &GoogleAccountRepository{db: db, cipher: cipher}
}

func (r *GoogleAccountRepository) scan(row interface {
	Scan(dest ...any) error
}, a *models.GoogleAccount) error {
	var encAccess, encRefresh string
	if err := row.Scan(
		&a.ID, &a.UserID, &a.GoogleEmail, &encAccess, &encRefresh, &a.TokenExpiry,
		&a.Scope, &a.CalendarID, &a.SyncToken, &a.LastSyncedAt, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return err
	}

	accessToken, err := r.cipher.Decrypt(encAccess)
	if err != nil {
		return err
	}
	refreshToken, err := r.cipher.Decrypt(encRefresh)
	if err != nil {
		return err
	}
	a.AccessToken = accessToken
	a.RefreshToken = refreshToken
	return nil
}

// Upsert stores a newly connected (or reconnected) account. One row per
// user, matching whatsapp_links' one-link-per-user pattern.
func (r *GoogleAccountRepository) Upsert(ctx context.Context, userID, email, accessToken, refreshToken string, expiry time.Time, scope string) (models.GoogleAccount, error) {
	var a models.GoogleAccount

	encAccess, err := r.cipher.Encrypt(accessToken)
	if err != nil {
		return a, err
	}
	encRefresh, err := r.cipher.Encrypt(refreshToken)
	if err != nil {
		return a, err
	}

	row := r.db.QueryRowContext(ctx,
		`INSERT INTO google_oauth_tokens (user_id, google_email, access_token, refresh_token, token_expiry, scope)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id) DO UPDATE SET
		   google_email = EXCLUDED.google_email,
		   access_token = EXCLUDED.access_token,
		   refresh_token = EXCLUDED.refresh_token,
		   token_expiry = EXCLUDED.token_expiry,
		   scope = EXCLUDED.scope,
		   -- a fresh connect should also force a full resync
		   sync_token = NULL,
		   updated_at = NOW()
		 RETURNING `+googleAccountColumns,
		userID, email, encAccess, encRefresh, expiry, scope,
	)
	err = r.scan(row, &a)
	return a, err
}

func (r *GoogleAccountRepository) GetByUserID(ctx context.Context, userID string) (models.GoogleAccount, error) {
	var a models.GoogleAccount
	row := r.db.QueryRowContext(ctx,
		`SELECT `+googleAccountColumns+` FROM google_oauth_tokens WHERE user_id = $1`,
		userID,
	)
	err := r.scan(row, &a)
	if err == sql.ErrNoRows {
		return a, ErrNotFound
	}
	return a, err
}

func (r *GoogleAccountRepository) ListAllConnected(ctx context.Context) ([]models.GoogleAccount, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+googleAccountColumns+` FROM google_oauth_tokens ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.GoogleAccount
	for rows.Next() {
		var a models.GoogleAccount
		if err := r.scan(rows, &a); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	if accounts == nil {
		accounts = []models.GoogleAccount{}
	}
	return accounts, rows.Err()
}

// UpdateTokens persists a refreshed access token. refreshToken should be
// whatever oauth2.Token.RefreshToken holds after a refresh — the oauth2
// library already preserves the prior refresh token when Google's response
// omits a new one, so this never needs special-casing an empty value.
func (r *GoogleAccountRepository) UpdateTokens(ctx context.Context, userID, accessToken, refreshToken string, expiry time.Time) error {
	encAccess, err := r.cipher.Encrypt(accessToken)
	if err != nil {
		return err
	}
	encRefresh, err := r.cipher.Encrypt(refreshToken)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE google_oauth_tokens SET access_token = $1, refresh_token = $2, token_expiry = $3, updated_at = NOW() WHERE user_id = $4`,
		encAccess, encRefresh, expiry, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GoogleAccountRepository) UpdateSyncState(ctx context.Context, userID string, syncToken *string, lastSyncedAt time.Time) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE google_oauth_tokens SET sync_token = $1, last_synced_at = $2, updated_at = NOW() WHERE user_id = $3`,
		syncToken, lastSyncedAt, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GoogleAccountRepository) Delete(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM google_oauth_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
