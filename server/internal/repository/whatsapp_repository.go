package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

type WhatsAppRepository struct {
	db *database.DB
}

func NewWhatsAppRepository(db *database.DB) *WhatsAppRepository {
	return &WhatsAppRepository{db: db}
}

func (r *WhatsAppRepository) CreateLinkCode(ctx context.Context, userID, code string, expiresAt time.Time) (models.WhatsAppLinkCode, error) {
	var c models.WhatsAppLinkCode
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO whatsapp_link_codes (user_id, code, expires_at) VALUES ($1, $2, $3)
		 RETURNING id, user_id, code, expires_at`,
		userID, code, expiresAt,
	).Scan(&c.ID, &c.UserID, &c.Code, &c.ExpiresAt)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return c, ErrCodeTaken
	}
	return c, err
}

func (r *WhatsAppRepository) ResolveLinkCode(ctx context.Context, code string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`UPDATE whatsapp_link_codes SET used_at = now()
		 WHERE code = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING user_id`,
		code,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return userID, err
}

func (r *WhatsAppRepository) CreateLink(ctx context.Context, userID, phoneNumber string) (models.WhatsAppLink, error) {
	var link models.WhatsAppLink

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return link, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM whatsapp_links WHERE user_id = $1`, userID); err != nil {
		return link, err
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO whatsapp_links (user_id, phone_number) VALUES ($1, $2)
		 ON CONFLICT (phone_number) DO UPDATE SET user_id = EXCLUDED.user_id, linked_at = now()
		 RETURNING id, user_id, phone_number, linked_at`,
		userID, phoneNumber,
	).Scan(&link.ID, &link.UserID, &link.PhoneNumber, &link.LinkedAt)
	if err != nil {
		return link, err
	}

	return link, tx.Commit()
}

func (r *WhatsAppRepository) GetLinkByPhone(ctx context.Context, phoneNumber string) (models.WhatsAppLink, error) {
	var link models.WhatsAppLink
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, phone_number, linked_at FROM whatsapp_links WHERE phone_number = $1`,
		phoneNumber,
	).Scan(&link.ID, &link.UserID, &link.PhoneNumber, &link.LinkedAt)
	if err == sql.ErrNoRows {
		return link, ErrNotFound
	}
	return link, err
}

func (r *WhatsAppRepository) GetLinkByUser(ctx context.Context, userID string) (models.WhatsAppLink, error) {
	var link models.WhatsAppLink
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, phone_number, linked_at FROM whatsapp_links WHERE user_id = $1`,
		userID,
	).Scan(&link.ID, &link.UserID, &link.PhoneNumber, &link.LinkedAt)
	if err == sql.ErrNoRows {
		return link, ErrNotFound
	}
	return link, err
}

func (r *WhatsAppRepository) DeleteLink(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM whatsapp_links WHERE user_id = $1`, userID)
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
