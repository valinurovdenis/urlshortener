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
	"github.com/valinurovdenis/urlshortener/internal/app/mocks"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) (*http.Response, string) {

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestShortenerHandler_redirect(t *testing.T) {
	existingURL := "http://existing.ru"
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockStorage := mocks.NewURLStorage(t)
	mockStorage.On("GetLongURL", "existing").Return(existingURL, nil).Once()
	mockStorage.On("GetLongURL", "non-existing").Return("", errors.New("some error")).Once()
	shortenerService := service.NewShortenerService(mockStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, "host/")
	ts := httptest.NewServer(ShortenerRouter(*handler))
	defer ts.Close()
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, tc.method, tc.shortURL, nil)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedLocation, resp.Header.Get("Location"), "Адрес редиректа не совпадает с ожидаемым")
		})
	}
}

func TestShortenerHandler_generate(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockGenerator.On("Generate").Return("existing1", nil).Once()
	mockStorage := mocks.NewURLStorage(t)
	mockStorage.On("GetShortURL", "http://existing1.ru").Return("existing1", nil).Twice()
	mockStorage.On("GetShortURL", "https://existing1.ru").Return("", errors.New("no such shortUrl")).Once()
	mockStorage.On("Store", "https://existing1.ru", "existing1").Return(nil).Once()
	shortURLHost := "host/"
	shortenerService := service.NewShortenerService(mockStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, shortURLHost)
	ts := httptest.NewServer(ShortenerRouter(*handler))
	defer ts.Close()
	testCases := []struct {
		name             string
		method           string
		URL              string
		expectedCode     int
		expectedShortURL string
	}{
		{name: "http", method: http.MethodPost, URL: "http://existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: shortURLHost + "existing1"},
		{name: "empty scheme", method: http.MethodPost, URL: "existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: shortURLHost + "existing1"},
		{name: "https", method: http.MethodPost, URL: "https://existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: shortURLHost + "existing1"},
		{name: "fake url", method: http.MethodPost, URL: "{:3fake-url:3}",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
		{name: "empty url", method: http.MethodPost, URL: "",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, resShortURL := testRequest(t, ts, tc.method, "/", bytes.NewBuffer([]byte(tc.URL)))
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			if tc.expectedShortURL != "" {
				assert.Equal(t, tc.expectedShortURL, resShortURL, "Короткая ссылка не совпадает с ожидаемой")
			}
		})
	}
}

func TestShortenerHandler_ServeHTTPBadRequest(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	mockGenerator := mocks.NewShortCutGenerator(t)
	shortenerService := service.NewShortenerService(mockStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, "host/")
	ts := httptest.NewServer(ShortenerRouter(*handler))
	defer ts.Close()
	testCases := []struct {
		name         string
		url          string
		method       string
		contentType  string
		expectedCode int
	}{
		{name: "method delete", url: "/asdf", method: http.MethodDelete, contentType: "text/plain"},
		{name: "method put", url: "/qwer", method: http.MethodPut, contentType: "text/plain"},
		{name: "method put", url: "/", method: http.MethodGet, contentType: "text/plain"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, tc.method, tc.url, nil)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
