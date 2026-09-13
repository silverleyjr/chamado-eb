package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"

	"chamadoApi/api"

	"github.com/golang-jwt/jwt/v5"
)

var UnAuthorizedError = errors.New("Invalid token.")

func secretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	return []byte(secret)
}

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := r.Header.Get("Authorization")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return secretKey(), nil
		})
		if err != nil {
			api.RequestErrorHandler(w, UnAuthorizedError)
			return
		}

		if !token.Valid {
			api.RequestErrorHandler(w, UnAuthorizedError)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
