package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const authUserKey contextKey = "authUser"

type AuthUser struct {
	ID   string
	Role string
}

func (app *application) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			app.WriteError(w, http.StatusUnauthorized, errors.New("missing auth cookie"))
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(os.Getenv("SECRET")), nil
		})
		if err != nil || !token.Valid {
			app.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			app.WriteError(w, http.StatusUnauthorized, errors.New("invalid claims"))
			return
		}

		expVal, ok := claims["exp"].(float64)
		if !ok || time.Now().Unix() > int64(expVal) {
			app.WriteError(w, http.StatusUnauthorized, errors.New("token expired"))
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			app.WriteError(w, http.StatusUnauthorized, errors.New("invalid subject"))
			return
		}

		role, _ := claims["role"].(string)

		authUser := AuthUser{
			ID:   sub,
			Role: role,
		}

		ctx := context.WithValue(r.Context(), authUserKey, authUser)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(authUserKey).(AuthUser)
		if !ok {
			app.WriteError(w, http.StatusUnauthorized, errors.New("user not authenticated"))
			return
		}

		if user.Role != "ADMIN" {
			app.WriteError(w, http.StatusFound, errors.New("forbidden: admin only"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

//func (app *application) RequireActivatedUser(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		userCtx, ok := r.Context().Value(authUserKey).(AuthUser)
//		if !ok {
//			app.WriteError(w, http.StatusUnauthorized, errors.New("user not authenticated"))
//			return
//		}
//
//		user, err := app.models.Users.GetByID(userCtx.ID)
//		if err != nil {
//			app.WriteError(w, http.StatusInternalServerError, err)
//			return
//		}
//
//		if !user.EmailVerified {
//			app.WriteError(w, http.StatusForbidden, errors.New("your account must be activated to perform this action"))
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}
// hz dlya chego eto
