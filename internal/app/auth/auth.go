// Package auth for authorization middlewares.
package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
)

// Struct for parsing user id from jwt token.
type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

const tokenExpiration = time.Hour * 3

// Class for authentication via jwt tokens.
type JwtAuthenticator struct {
	SecretKey   string
	UserStorage userstorage.UserStorage
}

// Returns new authenticator.
// Requires secret key for jwt and storage for generatings user ids.
func NewAuthenticator(secretKey string, userStorage userstorage.UserStorage) *JwtAuthenticator {
	return &JwtAuthenticator{
		SecretKey:   secretKey,
		UserStorage: userStorage,
	}
}

// Builds jwt string from given user id.
func (a *JwtAuthenticator) BuildJWTString(userID int64) (string, error) {
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

// Parses user id from jwt token string.
// Returns error if no jwt token is not valid.
func (a *JwtAuthenticator) GetUserID(tokenString string) (int64, error) {
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

// Middleware creates new user if no authorization cookie with valid user provided.
// Returns new authorization cookie if new user has been created.
func (a *JwtAuthenticator) CreateUserIfNeeded(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		var userID int64
		if err == nil {
			userID, err = a.GetUserID(cookie.Value)
		}

		if err != nil {
			userID, _ = a.UserStorage.GenerateUUID(r.Context())
			token, _ := a.BuildJWTString(userID)
			newCookie := http.Cookie{Name: "Authorization", Value: token}
			http.SetCookie(w, &newCookie)
		}
		strID := strconv.FormatInt(userID, 10)
		r.Header.Set("user_id", strID)

		h.ServeHTTP(w, r)
	})
}

// Middleware checks whether there is authorization cookie with valid user.
// Returns 401 if valid user not found.
func (a *JwtAuthenticator) OnlyWithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		var userID int64
		if err == nil {
			userID, err = a.GetUserID(cookie.Value)
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
