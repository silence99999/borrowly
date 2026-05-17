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

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("auth_token")
		if err != nil {
			writeError(w, http.StatusUnauthorized, errors.New("missing auth cookie"))
			return
		}
		tok, err := jwt.Parse(c.Value, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(os.Getenv("SECRET")), nil
		})
		if err != nil || !tok.Valid {
			writeError(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}
		claims, _ := tok.Claims.(jwt.MapClaims)
		exp, _ := claims["exp"].(float64)
		if time.Now().Unix() > int64(exp) {
			writeError(w, http.StatusUnauthorized, errors.New("token expired"))
			return
		}
		sub, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)
		ctx := context.WithValue(r.Context(), authUserKey, AuthUser{ID: sub, Role: role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
