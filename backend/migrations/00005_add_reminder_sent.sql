-- +goose Up
-- +goose StatementBegin
ALTER TABLE rentals
ADD COLUMN reminder_sent BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rentals
DROP COLUMN reminder_sent;
-- +goose StatementEnd
