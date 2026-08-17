CREATE TABLE ai_conversations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ai_conversations_user_id_idx ON ai_conversations (user_id);

CREATE TABLE ai_messages (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id      UUID        NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role                 TEXT        NOT NULL CHECK (role IN ('user', 'assistant')),
    content              JSONB       NOT NULL,
    whatsapp_message_id  TEXT        UNIQUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ai_messages_conversation_id_idx ON ai_messages (conversation_id, created_at);
