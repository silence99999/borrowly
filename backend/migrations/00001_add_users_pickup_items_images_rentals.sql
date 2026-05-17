-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       role VARCHAR(16) NOT NULL DEFAULT 'USER',
                       email_verified BOOLEAN NOT NULL DEFAULT FALSE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                       CONSTRAINT chk_users_role CHECK (role IN ('USER', 'ADMIN'))
);


CREATE TABLE pickup_points (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               address TEXT NOT NULL,
                               city VARCHAR(128) NOT NULL,
                               is_active BOOLEAN NOT NULL DEFAULT TRUE,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE items (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       pickup_point_id UUID REFERENCES pickup_points(id) ON DELETE SET NULL,

                       title VARCHAR(255) NOT NULL,
                       description TEXT,
                       category VARCHAR(64),

                       price_per_hour INT,
                       price_per_day INT,

                       delivery_type VARCHAR(16) NOT NULL DEFAULT 'SELF',
                       status VARCHAR(16) NOT NULL DEFAULT 'AVAILABLE',

                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                       CONSTRAINT chk_items_status CHECK (
                           status IN ('AVAILABLE', 'UNAVAILABLE', 'BLOCKED')
                           ),
                       CONSTRAINT chk_items_delivery CHECK (
                           delivery_type IN ('SELF', 'COURIER')
                           ),
                       CONSTRAINT chk_items_price CHECK (
                           price_per_hour IS NOT NULL OR price_per_day IS NOT NULL
                           )
);


CREATE TABLE item_images (
                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
                             image_url TEXT NOT NULL,
                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE rentals (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         item_id UUID NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
                         renter_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                         pickup_point_id UUID NOT NULL REFERENCES pickup_points(id) ON DELETE RESTRICT,

                         start_at TIMESTAMPTZ NOT NULL,
                         end_at TIMESTAMPTZ NOT NULL,

                         total_price INT NOT NULL,
                         platform_fee INT NOT NULL,
                         owner_income INT NOT NULL,

                         status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                         CONSTRAINT chk_rentals_time CHECK (end_at > start_at),
                         CONSTRAINT chk_rentals_status CHECK (
                             status IN ('PENDING', 'APPROVED', 'ACTIVE', 'COMPLETED', 'CANCELLED')
                             ),
                         CONSTRAINT chk_rentals_money CHECK (
                             total_price = platform_fee + owner_income
                             )
);


-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rentals;
DROP TABLE IF EXISTS item_images;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS pickup_points;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
