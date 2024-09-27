package utils

import (
	"errors"
	"net/url"
)

func SanitizeURL(origURL string) (string, error) {
	parsed, err := url.Parse(origURL)
	if err != nil {
		return "", errors.New("given string is not url")
	}
	if !parsed.IsAbs() {
		parsed.Scheme = "http"
	}
	return parsed.String(), nil
}
