package main

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/json"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) SignUp(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	_, err = mail.ParseAddress(input.Email)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if len(input.Password) < 6 {
		app.WriteError(w, http.StatusBadRequest, errors.New("password must be at least 6 characters"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	user := &data.User{
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	err = app.models.Users.Create(user)

	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	code, err := app.GenerateEmailCode()
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	token := &data.EmailToken{
		UserID:    user.ID,
		Code:      code,
		Purpose:   "EMAIL_VERIFY",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	err = app.models.EmailTokens.Create(token)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	app.background(func() {
		err := app.mailer.SendVerificationEmail(user.Email, code)
		if err != nil {
			slog.Error("failed to send verification email",
				"email", user.Email,
				"error", err,
			)
		}
	})

	err = json.WriteJSON(w, http.StatusOK, map[string]string{
		"verification_id": token.ID,
		"message":         "verification email sent",
	})
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, errors.New("invalid email or password"))
		return
	}

	cookie := CreateCookie(user.ID, user.Role)

	http.SetCookie(w, cookie)

	err = json.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "login successful",
	})
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		VerificationID string `json:"verification_id"`
		Code           string `json:"code"`
	}

	err := json.ReadJSON(r, &input)
	if err != nil {
		app.WriteError(w, http.StatusBadRequest, err)
		return
	}

	token, err := app.models.EmailTokens.GetByID(input.VerificationID)
	if err != nil || token.Code != input.Code || token.ExpiresAt.Before(time.Now()) {
		app.WriteError(w, http.StatusBadRequest, errors.New("invalid or expired code"))
		return
	}

	err = app.models.Users.MarkEmailVerified(token.UserID)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	user, err := app.models.Users.GetByID(token.UserID)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = app.models.EmailTokens.Delete(token.ID)
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	cookie := CreateCookie(token.UserID, user.ID)

	http.SetCookie(w, cookie)
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) Me(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(authUserKey).(AuthUser)

	err := json.WriteJSON(w, http.StatusOK, map[string]string{
		"id":   user.ID,
		"role": user.Role,
	})
	if err != nil {
		app.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func CreateCookie(userID, role string) *http.Cookie {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	secret := []byte(os.Getenv("SECRET"))
	tokenString, _ := jwtToken.SignedString(secret)

	cookie := &http.Cookie{
		Name:     "auth_token",
		Path:     "/",
		Value:    tokenString,
		MaxAge:   3600 * 24 * 30,
		Secure:   false,
		HttpOnly: true,
	}

	return cookie

}
