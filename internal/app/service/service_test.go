package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/mocks"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		origURL string
		want    string
		wantErr string
	}{
		{origURL: "http://sanitized.ru", want: "http://sanitized.ru", wantErr: ""},
		{origURL: "http://sanitized", want: "http://sanitized", wantErr: ""},
		{origURL: "sanitized.ru", want: "http://sanitized.ru", wantErr: ""},
		{origURL: "sanitized/asdf/qwer?user_id=1", want: "http://sanitized/asdf/qwer?user_id=1", wantErr: ""},
		{origURL: "https://sanitized.ru", want: "https://sanitized.ru", wantErr: ""},
		{origURL: "", want: "", wantErr: "empty string is not url"},
		{origURL: "://asdf://", want: "", wantErr: "given string is not url"},
	}
	for _, tt := range tests {
		t.Run("test", func(t *testing.T) {
			got, err := service.SanitizeURL(tt.origURL)
			assert.Equal(t, got, tt.want)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestShortenerService_GenerateShortURL(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockGenerator.On("Generate").Return("non-existing", nil).Twice()
	mockStorage := mocks.NewURLStorage(t)
	mockStorage.On("GetShortURLWithContext", mock.Anything, "http://existing.ru").Return("existing", nil).Once()
	mockStorage.On("StoreWithContext", mock.Anything, "http://non-existing.ru", "non-existing", "").Return(nil).Once()
	mockStorage.On("StoreWithContext", mock.Anything, "http://existing.ru", "non-existing", "").Return(urlstorage.ErrConflictURL).Once()
	mockUserStorage := mocks.NewUserURLStorage(t)
	service := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	testCases := []struct {
		name          string
		longURL       string
		expectedShort string
		expectedError error
	}{
		{name: "existing", longURL: "existing.ru", expectedShort: "existing", expectedError: urlstorage.ErrConflictURL},
		{name: "non-existing", longURL: "non-existing.ru", expectedShort: "non-existing", expectedError: nil},
		{name: "empty", longURL: "", expectedShort: "", expectedError: errors.New("empty string is not url")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shortURL, err := service.GenerateShortURLWithContext(context.Background(), tc.longURL, "")
			require.Equal(t, tc.expectedError, err, "Ошибка не совпадает")
			require.Equal(t, tc.expectedShort, shortURL, "Короткий урл не совпадает")
		})
	}
}

func TestShortenerService_GetShortURL(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	mockStorage := mocks.NewURLStorage(t)
	storageErr := errors.New("not found")
	mockStorage.On("GetLongURLWithContext", mock.Anything, "existing").Return("existing.ru", nil).Once()
	mockStorage.On("GetLongURLWithContext", mock.Anything, "non-existing").Return("", storageErr).Once()
	mockUserStorage := mocks.NewUserURLStorage(t)
	service := service.NewShortenerService(mockStorage, mockUserStorage, mockGenerator)
	testCases := []struct {
		name          string
		shortURL      string
		expectedLong  string
		expectedError error
	}{
		{name: "existing", shortURL: "existing", expectedLong: "existing.ru", expectedError: nil},
		{name: "non-existing", shortURL: "non-existing", expectedLong: "",
			expectedError: storageErr},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			longURL, err := service.GetLongURLWithContext(context.Background(), tc.shortURL)
			require.ErrorIs(t, err, tc.expectedError, "Ошибка не совпадает")
			require.Equal(t, tc.expectedLong, longURL, "Короткий урл не совпадает")
		})
	}
}
