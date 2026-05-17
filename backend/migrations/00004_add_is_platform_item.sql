-- +goose Up
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN is_platform_item boolean NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE items DROP COLUMN is_platform_item;
-- +goose StatementEnd
