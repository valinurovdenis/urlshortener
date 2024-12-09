package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/mocks"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader, headers map[string]string) (*http.Response, string) {

	req, err := http.NewRequest(method, ts.URL+path, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
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
	userStorage := mocks.NewUserStorage(t)
	userStorage.On("GenerateUUID", mock.Anything).Return(int64(1), nil).Times(2)
	auth := auth.NewAuthenticator("SECRET_KEY", userStorage)
	mockStorage.On("GetLongURLWithContext", mock.Anything, "existing").Return(existingURL, nil).Once()
	mockStorage.On("GetLongURLWithContext", mock.Anything, "non-existing").Return("", errors.New("some error")).Once()
	mockUserStorage := mocks.NewUserURLStorage(t)
	shortenerService := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, *auth, "host/")
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
			resp, _ := testRequest(t, ts, tc.method, tc.shortURL, nil, nil)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			assert.Equal(t, tc.expectedLocation, resp.Header.Get("Location"), "Адрес редиректа не совпадает с ожидаемым")
		})
	}
}

func TestShortenerHandler_generateSimple(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockGenerator.On("Generate").Return("existing1", nil).Times(3)
	mockStorage := mocks.NewURLStorage(t)
	userStorage := mocks.NewUserStorage(t)
	userStorage.On("GenerateUUID", mock.Anything).Return(int64(1), nil).Times(5)
	auth := auth.NewAuthenticator("SECRET_KEY", userStorage)
	mockStorage.On("GetShortURLWithContext", mock.Anything, "http://existing1.ru").Return("existing1", nil).Twice()
	mockStorage.On("StoreWithContext", mock.Anything, "http://existing1.ru", "existing1", "1").Return(urlstorage.ErrConflictURL).Twice()
	mockStorage.On("StoreWithContext", mock.Anything, "https://existing1.ru", "existing1", "1").Return(nil).Once()
	shortURLHost := "host/"
	mockUserStorage := mocks.NewUserURLStorage(t)
	shortenerService := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, *auth, shortURLHost)
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
			expectedCode: http.StatusConflict, expectedShortURL: shortURLHost + "existing1"},
		{name: "empty scheme", method: http.MethodPost, URL: "existing1.ru",
			expectedCode: http.StatusConflict, expectedShortURL: shortURLHost + "existing1"},
		{name: "https", method: http.MethodPost, URL: "https://existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: shortURLHost + "existing1"},
		{name: "fake url", method: http.MethodPost, URL: "{:3fake-url:3}",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
		{name: "empty url", method: http.MethodPost, URL: "",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, resShortURL := testRequest(t, ts, tc.method, "/",
				bytes.NewBuffer([]byte(tc.URL)), nil)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			if tc.expectedCode == http.StatusCreated {
				assert.Equal(t, tc.expectedShortURL, resShortURL, "Короткая ссылка не совпадает с ожидаемой")
			}
		})
	}
}

func TestShortenerHandler_generateJSON(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockGenerator.On("Generate").Return("existing1", nil).Times(3)
	mockStorage := mocks.NewURLStorage(t)
	userStorage := mocks.NewUserStorage(t)
	userStorage.On("GenerateUUID", mock.Anything).Return(int64(1), nil).Times(5)
	auth := auth.NewAuthenticator("SECRET_KEY", userStorage)
	mockStorage.On("GetShortURLWithContext", mock.Anything, "http://existing1.ru").Return("existing1", nil).Twice()
	mockStorage.On("StoreWithContext", mock.Anything, "http://existing1.ru", "existing1", "1").Return(urlstorage.ErrConflictURL).Twice()
	mockStorage.On("StoreWithContext", mock.Anything, "https://existing1.ru", "existing1", "1").Return(nil).Once()
	shortURLHost := "host/"
	mockUserStorage := mocks.NewUserURLStorage(t)
	shortenerService := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, *auth, shortURLHost)
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
			expectedCode: http.StatusConflict, expectedShortURL: shortURLHost + "existing1"},
		{name: "empty scheme", method: http.MethodPost, URL: "existing1.ru",
			expectedCode: http.StatusConflict, expectedShortURL: shortURLHost + "existing1"},
		{name: "https", method: http.MethodPost, URL: "https://existing1.ru",
			expectedCode: http.StatusCreated, expectedShortURL: shortURLHost + "existing1"},
		{name: "fake url", method: http.MethodPost, URL: "{:3fake-url:3}",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
		{name: "empty url", method: http.MethodPost, URL: "",
			expectedCode: http.StatusBadRequest, expectedShortURL: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var input bytes.Buffer
			json.NewEncoder(&input).Encode(inputURL{tc.URL})
			resp, resShortURL := testRequest(t, ts, tc.method, "/api/shorten", &input, nil)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			if tc.expectedCode == http.StatusCreated {
				var getShortURL resultURL
				json.NewDecoder(strings.NewReader(resShortURL)).Decode(&getShortURL)
				assert.Equal(t, tc.expectedShortURL, getShortURL.URL, "Короткая ссылка не совпадает с ожидаемой")
			}
		})
	}
}

func TestShortenerHandler_generateGzip(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockStorage := mocks.NewURLStorage(t)
	userStorage := mocks.NewUserStorage(t)
	userStorage.On("GenerateUUID", mock.Anything).Return(int64(1), nil).Once()
	auth := auth.NewAuthenticator("SECRET_KEY", userStorage)
	mockGenerator.On("Generate").Return("existing1", nil).Once()
	mockStorage.On("StoreWithContext", mock.Anything, "http://existing1.ru", "existing1", "1").Return(nil).Once()
	shortURLHost := "host/"
	mockUserStorage := mocks.NewUserURLStorage(t)
	shortenerService := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, *auth, shortURLHost)
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var compressedInput bytes.Buffer
			writer := gzip.NewWriter(&compressedInput)
			writer.Write([]byte(tc.URL))
			writer.Close()
			resp, resCompressedURL := testRequest(t, ts, tc.method, "/", &compressedInput,
				map[string]string{"Content-Encoding": "gzip", "Accept-Encoding": "gzip"})
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			require.Contains(t, "gzip", resp.Header.Get("Accept-Encoding"), "Ответ не содержит Accept-Encoding: gzip")
			if tc.expectedShortURL != "" {
				gz, _ := gzip.NewReader(strings.NewReader(resCompressedURL))
				decompressedURL, _ := io.ReadAll(gz)
				assert.Equal(t, tc.expectedShortURL, string(decompressedURL), "Короткая ссылка не совпадает с ожидаемой")
			}
		})
	}
}

func TestShortenerHandler_ServeHTTPBadRequest(t *testing.T) {
	mockStorage := mocks.NewURLStorage(t)
	mockGenerator := mocks.NewShortCutGenerator(t)
	userStorage := mocks.NewUserStorage(t)
	userStorage.On("GenerateUUID", mock.Anything).Return(int64(1), nil).Times(3)
	auth := auth.NewAuthenticator("SECRET_KEY", userStorage)
	mockUserStorage := mocks.NewUserURLStorage(t)
	shortenerService := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	handler := NewShortenerHandler(*shortenerService, *auth, "host/")
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
			resp, _ := testRequest(t, ts, tc.method, tc.url, nil, nil)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
