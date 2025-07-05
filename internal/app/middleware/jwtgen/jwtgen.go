// Package jwtgen пакет с описание jwt токена и авторизацией по Cookie
package jwtgen

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
)

// Claims требования
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildJWTString построение jwt строки по заданной строке
func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExp)),
		},
		UserID: 2,
	})

	tokenString, err := token.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID получить userID из jwt
func GetUserID(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.SecretKey), nil
		})
	if err != nil {
		return 0
	}

	if !token.Valid {
		return 0
	}

	return claims.UserID
}

// AuthMiddleware прослойка для авторизации
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			token, err := BuildJWTString()
			if err != nil {
				http.Error(w, err.Error(), 400)
				next.ServeHTTP(w, r)
			}
			cookie := &http.Cookie{
				Name:     "token",
				Value:    token,
				HttpOnly: true,
			}
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
		}
		next.ServeHTTP(w, r)
	})
}
