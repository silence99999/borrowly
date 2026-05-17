-- +goose Up
-- +goose StatementBegin

ALTER TABLE items
    DROP COLUMN delivery_type;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE items
    ADD COLUMN delivery_type VARCHAR(16) NOT NULL DEFAULT 'SELF';

ALTER TABLE items
    ADD CONSTRAINT chk_items_delivery
        CHECK (delivery_type IN ('SELF', 'COURIER'));

-- +goose StatementEnd
