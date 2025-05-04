package urlstorage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
)

func TestSimpleMapLockStorage_GetLongURL(t *testing.T) {
	storage := urlstorage.SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name     string
		s        *urlstorage.SimpleMapLockStorage
		shortURL string
		want     string
		wantErr  error
	}{
		{name: "get_a", s: &storage, shortURL: "a", want: "url_a", wantErr: nil},
		{name: "get_b", s: &storage, shortURL: "b", want: "", wantErr: errors.New("no such shortUrl")},
		{name: "get_empty", s: &storage, shortURL: "", want: "", wantErr: errors.New("no such shortUrl")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetLongURLWithContext(context.Background(), tt.shortURL)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSimpleMapLockStorage_GetShortURL(t *testing.T) {
	storage := urlstorage.SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name    string
		s       *urlstorage.SimpleMapLockStorage
		longURL string
		want    string
		wantErr error
	}{
		{name: "get_a", s: &storage, longURL: "url_a", want: "a", wantErr: nil},
		{name: "get_b", s: &storage, longURL: "url_b", want: "", wantErr: errors.New("no such longUrl")},
		{name: "get_empty", s: &storage, longURL: "", want: "", wantErr: errors.New("no such longUrl")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetShortURLWithContext(context.Background(), tt.longURL)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSimpleMapLockStorage_Store(t *testing.T) {
	storage := urlstorage.SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name          string
		longURL       string
		shortURL      string
		expectedError error
	}{
		{name: "store_a", longURL: "url_a", shortURL: "a", expectedError: urlstorage.ErrConflictURL},
		{name: "store_b", longURL: "url_b", shortURL: "b", expectedError: nil},
		{name: "store_empty", longURL: "", shortURL: "", expectedError: errors.New("cannot save empty url")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.StoreWithContext(context.Background(), tt.longURL, tt.shortURL, "")
			require.Equal(t, tt.expectedError, err)
			if err == nil {
				assert.Subset(t, storage.URL2ShortURL, map[string]string{tt.longURL: tt.shortURL})
				assert.Subset(t, storage.ShortURL2Url, map[string]string{tt.shortURL: tt.longURL})
			}
		})
	}
}

func TestSimpleMapLockStorage_StoreMany(t *testing.T) {
	storage := urlstorage.SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name          string
		urlsToStore   []urlstorage.URLPair
		expectedError error
	}{
		{name: "store_many",
			urlsToStore: []urlstorage.URLPair{
				{Long: "url_a", Short: "a"},
				{Long: "url_b", Short: "b"}},
			expectedError: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := storage.StoreManyWithContext(context.Background(), tt.urlsToStore, "")
			require.Equal(t, tt.expectedError, err)
			assert.Equal(t, errs, []error{urlstorage.ErrConflictURL, nil})
			assert.Equal(t, storage.URL2ShortURL,
				map[string]string{"url_a": "a", "url_b": "b"})
			assert.Subset(t, storage.ShortURL2Url,
				map[string]string{"a": "url_a", "b": "url_b"})
		})
	}
}

func TestSimpleMapLockStorage_Clear(t *testing.T) {
	storage := urlstorage.NewSimpleMapLockStorage()
	storage.StoreManyWithContext(context.Background(), []urlstorage.URLPair{
		{Long: "url_a", Short: "a"},
		{Long: "url_b", Short: "b"}}, "")

	err := storage.Clear()
	require.Equal(t, nil, err)

	assert.Equal(t, storage.URL2ShortURL,
		map[string]string{})
	assert.Equal(t, storage.ShortURL2Url,
		map[string]string{})
}

func TestSimpleMapLockStorage_Ping(t *testing.T) {
	storage := urlstorage.NewSimpleMapLockStorage()
	err := storage.Ping()
	require.NoError(t, err)
}

func TestSimpleMapLockStorage_GetUserURLs(t *testing.T) {
	storage := urlstorage.NewSimpleMapLockStorage()
	_, err := storage.GetUserURLs(context.Background(), "user_1")
	// TODO: write tests
	require.Error(t, err)
}

func TestSimpleMapLockStorage_DeleteUserURLs(t *testing.T) {
	storage := urlstorage.NewSimpleMapLockStorage()
	err := storage.DeleteUserURLs(context.Background(), urlstorage.URLsForDelete{UserID: "user", ShortURLs: []string{"short1", "short2"}})
	// TODO: write tests
	require.Error(t, err)
}
