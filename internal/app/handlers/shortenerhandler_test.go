package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage/mocks"
)

func TestShortenerHandler_redirect(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	existingURL := "http://existing.ru"
	handler := ShortenerHandler{
		Storage: mockStorage,
		Host:    "http://localhost:8080/"}
	testCases := []struct {
		name             string
		method           string
		shortURL         string
		expectedCode     int
		expectedLocation string
	}{
		{name: "existing", method: http.MethodGet, shortURL: "/existing",
			expectedCode: http.StatusTemporaryRedirect, expectedLocation: existingURL},
		{name: "non-existing", method: http.MethodGet, shortURL: "/non-existing",
			expectedCode: http.StatusBadRequest, expectedLocation: ""},
		{name: "empty", method: http.MethodGet, shortURL: "/",
			expectedCode: http.StatusBadRequest, expectedLocation: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.shortURL, nil)
			w := httptest.NewRecorder()

			if tc.expectedLocation != "" {
				mockStorage.On("Get", tc.shortURL[1:]).Return(existingURL, nil).Once()
			} else {
				mockStorage.On("Get", tc.shortURL[1:]).Return("", errors.New("some error")).Once()
			}
			handler.redirect(w, r)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedLocation, res.Header.Get("Location"), "Адрес редиректа не совпадает с ожидаемым")
		})
	}
}

func TestShortenerHandler_generate(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	mockStorage.On("Store", "http://existing1.ru").Return("existing1", nil).Twice()
	mockStorage.On("Store", "https://existing2.ru").Return("existing2", nil).Once()
	handler := ShortenerHandler{
		Storage: mockStorage,
		Host:    "http://localhost:8080/"}
	testCases := []struct {
		name             string
		method           string
		URL              string
		expectedCode     int
		expectedShortURL string
	}{
		{name: "http", method: http.MethodPost, URL: "http://existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: handler.Host + "existing1"},
		{name: "empty scheme", method: http.MethodPost, URL: "existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: handler.Host + "existing1"},
		{name: "https", method: http.MethodPost, URL: "https://existing2.ru",
			expectedCode: http.StatusCreated, expectedShortURL: handler.Host + "existing2"},
		{name: "fake url", method: http.MethodPost, URL: "{:3fake-url:3}",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
		{name: "empty url", method: http.MethodPost, URL: "",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", bytes.NewBuffer([]byte(tc.URL)))
			w := httptest.NewRecorder()

			handler.generate(w, r)
			res := w.Result()
			defer res.Body.Close()
			resShortURL, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			require.Equal(t, tc.expectedCode, res.StatusCode, "Код ответа не совпадает с ожидаемым")
			if tc.expectedShortURL != "" {
				assert.Equal(t, tc.expectedShortURL, string(resShortURL), "Короткая ссылка не совпадает с ожидаемой")
			}
		})
	}
}

func TestShortenerHandler_ServeHTTPBadRequest(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	handler := ShortenerHandler{
		Storage: mockStorage,
		Host:    "http://localhost:8080/"}
	testCases := []struct {
		name         string
		method       string
		contentType  string
		expectedCode int
	}{
		{name: "method delete", method: http.MethodDelete, contentType: "text/plain"},
		{name: "method put", method: http.MethodPut, contentType: "text/plain"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", nil)
			w := httptest.NewRecorder()

			r.Header.Set("Content-Type", tc.contentType)
			handler.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, "Not supported method\n", string(resBody), "Не совпадает текст ошибки")
		})
	}
}
