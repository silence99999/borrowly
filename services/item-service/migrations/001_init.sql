CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE pickup_points (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    address    TEXT         NOT NULL,
    city       VARCHAR(128) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- owner_id references users in auth-service (no FK constraint — cross-service boundary)
CREATE TABLE items (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID        NOT NULL,
    pickup_point_id UUID        REFERENCES pickup_points(id) ON DELETE SET NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    category        VARCHAR(64),
    price_per_hour  INT,
    price_per_day   INT,
    status          VARCHAR(16)  NOT NULL DEFAULT 'AVAILABLE',
    is_platform_item BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_items_status CHECK (status IN ('AVAILABLE', 'UNAVAILABLE', 'BLOCKED')),
    CONSTRAINT chk_items_price  CHECK (price_per_hour IS NOT NULL OR price_per_day IS NOT NULL)
);

CREATE TABLE item_images (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id    UUID        NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    image_url  TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
