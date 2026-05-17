CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- item_id and reviewed_user_id are owned by other services — no FK constraints.
CREATE TABLE reviews (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id          UUID        NOT NULL,
    reviewed_user_id UUID        NOT NULL,
    rating           INT         NOT NULL,
    comment          TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_reviews_rating CHECK (rating >= 0 AND rating <= 5)
);
