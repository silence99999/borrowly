package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

type EmailToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (app *application) dbCreateUser(email, hash string) (*User, error) {
	var u User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := app.db.QueryRowContext(ctx,
		`INSERT INTO users (email,password_hash) VALUES ($1,$2) RETURNING id,created_at`,
		email, hash,
	).Scan(&u.ID, &u.CreatedAt)
	u.Email = email
	return &u, err
}

func (app *application) dbGetUserByEmail(email string) (*User, error) {
	var u User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := app.db.QueryRowContext(ctx,
		`SELECT id,email,password_hash,role,email_verified,created_at FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.EmailVerified, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return &u, err
}

func (app *application) dbGetUserByID(id string) (*User, error) {
	var u User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := app.db.QueryRowContext(ctx,
		`SELECT id,email,role,email_verified,created_at FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Email, &u.Role, &u.EmailVerified, &u.CreatedAt)
	return &u, err
}

func (app *application) dbMarkEmailVerified(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := app.db.ExecContext(ctx, `UPDATE users SET email_verified=true WHERE id=$1`, userID)
	return err
}

func (app *application) dbCreateEmailToken(userID, code string) (*EmailToken, error) {
	t := &EmailToken{UserID: userID, Code: code, ExpiresAt: time.Now().Add(10 * time.Minute)}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := app.db.QueryRowContext(ctx,
		`INSERT INTO email_tokens(user_id,code,purpose,expires_at) VALUES($1,$2,'EMAIL_VERIFY',$3) RETURNING id`,
		userID, code, t.ExpiresAt,
	).Scan(&t.ID)
	return t, err
}

func (app *application) dbGetEmailToken(id string) (*EmailToken, error) {
	var t EmailToken
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := app.db.QueryRowContext(ctx,
		`SELECT id,user_id,code,expires_at FROM email_tokens WHERE id=$1`, id,
	).Scan(&t.ID, &t.UserID, &t.Code, &t.ExpiresAt)
	return &t, err
}

func (app *application) dbDeleteEmailToken(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := app.db.ExecContext(ctx, `DELETE FROM email_tokens WHERE id=$1`, id)
	return err
}
