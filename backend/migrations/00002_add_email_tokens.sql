-- +goose Up
-- +goose StatementBegin

CREATE TABLE email_tokens (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                              code VARCHAR(6) NOT NULL,
                              purpose VARCHAR(32) NOT NULL,
                              expires_at TIMESTAMPTZ NOT NULL,
                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                              CONSTRAINT chk_email_tokens_purpose CHECK (
                                  purpose IN ('EMAIL_VERIFY', 'LOGIN_2FA')
                              )
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS email_tokens;
-- +goose StatementEnd
