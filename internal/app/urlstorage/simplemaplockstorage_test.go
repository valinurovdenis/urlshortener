package urlstorage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator/mocks"
)

func TestSimpleMapLockStorage_Get(t *testing.T) {
	mockGenerator := mocks.NewShortCutGenerator(t)
	storage := SimpleMapLockStorage{
		Generator:    mockGenerator,
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
			got, err := tt.s.Get(tt.shortURL)
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
	mockGenerator := mocks.NewShortCutGenerator(t)
	storage := SimpleMapLockStorage{
		Generator:    mockGenerator,
		ShortURL2Url: map[string]string{"a": "url_a"},
		URL2ShortURL: map[string]string{"url_a": "a"}}
	tests := []struct {
		name           string
		s              *SimpleMapLockStorage
		URL            string
		want           string
		wantErr        error
		shortGenerated bool
	}{
		{name: "store_a", s: &storage, URL: "url_a", want: "a", wantErr: nil, shortGenerated: false},
		{name: "store_b", s: &storage, URL: "url_b", want: "b", wantErr: nil, shortGenerated: true},
		{name: "store_empty", s: &storage, URL: "", want: "", wantErr: errors.New("cannot save empty url"), shortGenerated: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shortGenerated {
				mockGenerator.On("Generate").Return(tt.want, nil).Once()
			}
			got, err := tt.s.Store(tt.URL)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
			if err == nil {
				assert.Subset(t, tt.s.URL2ShortURL, map[string]string{tt.URL: tt.want})
				assert.Subset(t, tt.s.ShortURL2Url, map[string]string{tt.want: tt.URL})
			}
		})
	}
}
