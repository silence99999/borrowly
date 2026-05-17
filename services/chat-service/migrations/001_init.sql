CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- sender_id and receiver_id reference users owned by auth-service — no FK constraints.
CREATE TABLE chat_messages (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id   UUID        NOT NULL,
    receiver_id UUID        NOT NULL,
    content     TEXT        NOT NULL,
    is_read     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_chat_no_self_message CHECK (sender_id <> receiver_id)
);

CREATE INDEX idx_chat_sender   ON chat_messages(sender_id);
CREATE INDEX idx_chat_receiver ON chat_messages(receiver_id);
CREATE INDEX idx_chat_created  ON chat_messages(created_at);
