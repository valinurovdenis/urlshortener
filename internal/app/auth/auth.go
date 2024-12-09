package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

const tokenExpiration = time.Hour * 3

type JwtAuthenticator struct {
	SecretKey   string
	UserStorage userstorage.UserStorage
}

func NewAuthenticator(secretKey string, userStorage userstorage.UserStorage) *JwtAuthenticator {
	return &JwtAuthenticator{
		SecretKey:   secretKey,
		UserStorage: userStorage,
	}
}

func (a *JwtAuthenticator) buildJWTString(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *JwtAuthenticator) getUserID(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.SecretKey), nil
		})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

func (a *JwtAuthenticator) CreateUserIfNeeded(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		var userID int64
		if err == nil {
			userID, err = a.getUserID(cookie.Value)
		}

		if err != nil {
			userID, _ = a.UserStorage.GenerateUUID(r.Context())
			token, _ := a.buildJWTString(userID)
			newCookie := http.Cookie{Name: "Authorization", Value: token}
			http.SetCookie(w, &newCookie)
		}
		strID := strconv.FormatInt(userID, 10)
		r.Header.Set("user_id", strID)

		h.ServeHTTP(w, r)
	})
}

func (a *JwtAuthenticator) OnlyWithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		var userID int64
		if err == nil {
			userID, err = a.getUserID(cookie.Value)
		}

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		strID := strconv.FormatInt(userID, 10)
		r.Header.Set("user_id", strID)

		h.ServeHTTP(w, r)
	})
}
