package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/models"
)

type ConversationRepository struct {
	db *database.DB
}

func NewConversationRepository(db *database.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) GetOrCreateActiveConversation(ctx context.Context, userID string, inactivityWindow time.Duration) (models.AIConversation, error) {
	var conv models.AIConversation
	cutoff := time.Now().Add(-inactivityWindow)

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, started_at, last_message_at FROM ai_conversations
		 WHERE user_id = $1 AND last_message_at > $2
		 ORDER BY last_message_at DESC LIMIT 1`,
		userID, cutoff,
	).Scan(&conv.ID, &conv.UserID, &conv.StartedAt, &conv.LastMessageAt)
	if err == nil {
		return conv, nil
	}
	if err != sql.ErrNoRows {
		return conv, err
	}

	err = r.db.QueryRowContext(ctx,
		`INSERT INTO ai_conversations (user_id) VALUES ($1) RETURNING id, user_id, started_at, last_message_at`,
		userID,
	).Scan(&conv.ID, &conv.UserID, &conv.StartedAt, &conv.LastMessageAt)
	return conv, err
}

func (r *ConversationRepository) AppendMessage(ctx context.Context, conversationID, role string, content json.RawMessage, whatsappMessageID *string) (models.AIMessage, error) {
	var msg models.AIMessage

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return msg, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO ai_messages (conversation_id, role, content, whatsapp_message_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, conversation_id, role, content, whatsapp_message_id, created_at`,
		conversationID, role, content, whatsappMessageID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.WhatsAppMessageID, &msg.CreatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return msg, ErrDuplicateMessage
		}
		return msg, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE ai_conversations SET last_message_at = now() WHERE id = $1`, conversationID); err != nil {
		return msg, err
	}

	return msg, tx.Commit()
}

func (r *ConversationRepository) RecentMessages(ctx context.Context, conversationID string, limit int) ([]models.AIMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, whatsapp_message_id, created_at FROM (
			SELECT id, conversation_id, role, content, whatsapp_message_id, created_at
			FROM ai_messages WHERE conversation_id = $1
			ORDER BY created_at DESC LIMIT $2
		 ) recent ORDER BY created_at ASC`,
		conversationID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.AIMessage
	for rows.Next() {
		var m models.AIMessage
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.WhatsAppMessageID, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []models.AIMessage{}
	}
	return messages, rows.Err()
}
