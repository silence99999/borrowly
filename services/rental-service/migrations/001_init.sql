CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- item_id, renter_id, pickup_point_id are UUIDs owned by other services.
-- No FK constraints here — cross-service boundary.
CREATE TABLE rentals (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id         UUID        NOT NULL,
    renter_id       UUID        NOT NULL,
    pickup_point_id UUID        NOT NULL,
    start_at        TIMESTAMPTZ NOT NULL,
    end_at          TIMESTAMPTZ NOT NULL,
    total_price     INT         NOT NULL,
    platform_fee    INT         NOT NULL,
    owner_income    INT         NOT NULL,
    status          VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    reminder_sent   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_rentals_time   CHECK (end_at > start_at),
    CONSTRAINT chk_rentals_status CHECK (status IN ('PENDING','APPROVED','ACTIVE','COMPLETED','CANCELLED')),
    CONSTRAINT chk_rentals_money  CHECK (total_price = platform_fee + owner_income)
);
