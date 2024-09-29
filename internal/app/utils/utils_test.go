package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			got, err := SanitizeURL(tt.origURL)
			assert.Equal(t, got, tt.want)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
