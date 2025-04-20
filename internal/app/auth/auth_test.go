// Package auth for authorization middlewares.
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/mocks"
)

func TestJwtAuthenticator_testJWTToken(t *testing.T) {
	mockUserStorage := mocks.NewUserStorage(t)
	secretKey := "asdf"
	var userID int64 = 1
	authenticator := auth.NewAuthenticator(secretKey, mockUserStorage)
	validTokenString, err := authenticator.BuildJWTString(userID)
	require.NoError(t, err)
	wrongSigningMethodTokenString, _ := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.RegisteredClaims{}).SignedString([]byte(secretKey))

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
	}{
		{name: "valid", tokenString: validTokenString, wantErr: false},
		{name: "invalid", tokenString: "asdf", wantErr: true},
		{name: "wrongMethod", tokenString: wrongSigningMethodTokenString, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := authenticator.GetUserID(tt.tokenString)
			require.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, userID, res)
			}
		})
	}
}

func TestJwtAuthenticator_CreateUserIfNeeded(t *testing.T) {
	userID := int64(1)
	var authCookie string
	mockUserStorage := mocks.NewUserStorage(t)
	mockUserStorage.On("GenerateUUID", mock.Anything).Return(userID, nil).Once()
	authenticator := auth.NewAuthenticator("asdf", mockUserStorage)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		strUserID := r.Header.Get("user_id")
		authCookie = w.Header().Get("Set-Cookie")
		getUserID, _ := strconv.Atoi(strUserID)
		require.Equal(t, int64(getUserID), userID)
	})

	handlerToTest := authenticator.CreateUserIfNeeded(nextHandler)

	req1 := httptest.NewRequest("GET", "http://testing", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req1)

	req2 := httptest.NewRequest("GET", "http://testing", nil)
	cookie := http.Cookie{Name: "Authorization", Value: authCookie[len("Authorization="):]}
	req2.AddCookie(&cookie)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req2)
}

func TestJwtAuthenticator_OnlyWithAuth(t *testing.T) {
	userID := int64(1)
	var authCookie string
	mockUserStorage := mocks.NewUserStorage(t)
	mockUserStorage.On("GenerateUUID", mock.Anything).Return(userID, nil).Once()
	authenticator := auth.NewAuthenticator("asdf", mockUserStorage)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		strUserID := r.Header.Get("user_id")
		authCookie = w.Header().Get("Set-Cookie")
		getUserID, _ := strconv.Atoi(strUserID)
		require.Equal(t, int64(getUserID), userID)
	})

	handlerToAuth := authenticator.CreateUserIfNeeded(nextHandler)
	req1 := httptest.NewRequest("GET", "http://testing", nil)
	handlerToAuth.ServeHTTP(httptest.NewRecorder(), req1)

	handlerToTest := authenticator.OnlyWithAuth(nextHandler)
	req2 := httptest.NewRequest("GET", "http://testing", nil)
	cookie := http.Cookie{Name: "Authorization", Value: authCookie[len("Authorization="):]}
	req2.AddCookie(&cookie)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req2)
}
