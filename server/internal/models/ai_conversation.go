package models

import (
	"encoding/json"
	"time"
)

type AIConversation struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	StartedAt     time.Time `json:"startedAt"`
	LastMessageAt time.Time `json:"lastMessageAt"`
}

type AIMessage struct {
	ID                string          `json:"id"`
	ConversationID    string          `json:"conversationId"`
	Role              string          `json:"role"` // "user" | "assistant"
	Content           json.RawMessage `json:"content"`
	WhatsAppMessageID *string         `json:"whatsappMessageId,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
}
