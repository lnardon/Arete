CREATE TABLE whatsapp_links (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number TEXT        NOT NULL UNIQUE,
    linked_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX whatsapp_links_user_id_idx ON whatsapp_links (user_id);

CREATE TABLE whatsapp_link_codes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX whatsapp_link_codes_user_id_idx ON whatsapp_link_codes (user_id);
