package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type EmailToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailTokenModel struct {
	DB *sql.DB
}

func (et *EmailTokenModel) Create(token *EmailToken) error {
	query := `INSERT INTO email_tokens (user_id,code, purpose, expires_at) VALUES ($1,$2,$3,$4) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{token.UserID, token.Code, token.Purpose, token.ExpiresAt}

	err := et.DB.QueryRowContext(ctx, query, args...).Scan(
		&token.ID,
	)

	if err != nil {
		return err
	}
	return nil
}

func (et *EmailTokenModel) GetByID(verificationID string) (*EmailToken, error) {
	query := `SELECT id,user_id,code,expires_at FROM email_tokens WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var emailToken EmailToken

	err := et.DB.QueryRowContext(ctx, query, verificationID).Scan(
		&emailToken.ID,
		&emailToken.UserID,
		&emailToken.Code,
		&emailToken.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}
	return &emailToken, nil
}

func (et *EmailTokenModel) Delete(verificationID string) error {
	query := `DELETE FROM email_tokens WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := et.DB.ExecContext(ctx, query, verificationID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("token not found")
	}

	return nil
}
