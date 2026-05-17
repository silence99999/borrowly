package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func readJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	_ = writeJSON(w, status, map[string]string{"error": err.Error()})
}

func genCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (app *application) bgRun(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if v := recover(); v != nil {
				slog.Error("background panic", "err", v)
			}
		}()
		fn()
	}()
}

func authCookie(userID, role string) *http.Cookie {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "role": role,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	s, _ := tok.SignedString([]byte(os.Getenv("SECRET")))
	return &http.Cookie{
		Name: "auth_token", Value: s, Path: "/",
		MaxAge: 86400 * 30, HttpOnly: true,
	}
}

func (app *application) signUp(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := mail.ParseAddress(in.Email); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid email"))
		return
	}
	if len(in.Password) < 6 {
		writeError(w, http.StatusBadRequest, errors.New("password must be at least 6 characters"))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	user, err := app.dbCreateUser(in.Email, string(hash))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	code, err := genCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	token, err := app.dbCreateEmailToken(user.ID, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	app.bgRun(func() {
		if err := app.sendVerificationEmail(in.Email, code); err != nil {
			slog.Error("send verify email", "err", err)
		}
	})
	_ = writeJSON(w, http.StatusOK, map[string]string{
		"verification_id": token.ID,
		"message":         "verification email sent",
	})
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	user, err := app.dbGetUserByEmail(in.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		writeError(w, http.StatusUnauthorized, errors.New("invalid email or password"))
		return
	}
	http.SetCookie(w, authCookie(user.ID, user.Role))
	_ = writeJSON(w, http.StatusOK, map[string]string{"message": "login successful"})
}

func (app *application) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var in struct {
		VerificationID string `json:"verification_id"`
		Code           string `json:"code"`
	}
	if err := readJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	tok, err := app.dbGetEmailToken(in.VerificationID)
	if err != nil || tok.Code != in.Code || tok.ExpiresAt.Before(time.Now()) {
		writeError(w, http.StatusBadRequest, errors.New("invalid or expired code"))
		return
	}
	_ = app.dbMarkEmailVerified(tok.UserID)
	user, err := app.dbGetUserByID(tok.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_ = app.dbDeleteEmailToken(tok.ID)
	http.SetCookie(w, authCookie(user.ID, user.Role))
	_ = writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "auth_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) me(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(authUserKey).(AuthUser)
	_ = writeJSON(w, http.StatusOK, map[string]string{"id": u.ID, "role": u.Role})
}

// internalGetUser is a service-to-service endpoint (no JWT required).
// It is NOT routed through the gateway — only reachable within the Docker network.
func (app *application) internalGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := app.dbGetUserByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	_ = writeJSON(w, http.StatusOK, map[string]string{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}
