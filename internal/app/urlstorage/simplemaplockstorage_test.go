package urlstorage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleMapLockStorage_GetLongURL(t *testing.T) {
	storage := SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name     string
		s        *SimpleMapLockStorage
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
			got, err := tt.s.GetLongURL(tt.shortURL)
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
	storage := SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name    string
		s       *SimpleMapLockStorage
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
			got, err := tt.s.GetShortURL(tt.longURL)
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
	storage := SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name          string
		longURL       string
		shortURL      string
		expectedError error
	}{
		{name: "store_a", longURL: "url_a", shortURL: "a", expectedError: nil},
		{name: "store_b", longURL: "url_b", shortURL: "b", expectedError: nil},
		{name: "store_empty", longURL: "", shortURL: "", expectedError: errors.New("cannot save empty url")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Store(tt.longURL, tt.shortURL)
			require.Equal(t, tt.expectedError, err)
			if err == nil {
				assert.Subset(t, storage.URL2ShortURL, map[string]string{tt.longURL: tt.shortURL})
				assert.Subset(t, storage.ShortURL2Url, map[string]string{tt.shortURL: tt.longURL})
			}
		})
	}
}

func TestSimpleMapLockStorage_StoreMany(t *testing.T) {
	storage := SimpleMapLockStorage{
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name          string
		urlsToStore   map[string]string
		expectedError error
	}{
		{name: "store_many", urlsToStore: map[string]string{"url_a": "a", "url_b": "b"},
			expectedError: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.StoreMany(tt.urlsToStore)
			require.Equal(t, tt.expectedError, err)
			assert.Equal(t, storage.URL2ShortURL,
				map[string]string{"url_a": "a", "url_b": "b"})
			assert.Subset(t, storage.ShortURL2Url,
				map[string]string{"a": "url_a", "b": "url_b"})
		})
	}
}
