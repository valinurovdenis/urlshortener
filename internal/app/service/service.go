package service

import (
	"context"
	"errors"
	"net/url"

	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
)

func sanitizeURL(origURL string) (string, error) {
	if origURL == "" {
		return "", errors.New("empty string is not url")
	}
	parsed, err := url.Parse(origURL)
	if err != nil {
		return "", errors.New("given string is not url")
	}
	if !parsed.IsAbs() {
		parsed.Scheme = "http"
	}
	return parsed.String(), nil
}

type ShortenerService struct {
	URLStorage urlstorage.URLStorage
	Generator  shortcutgenerator.ShortCutGenerator
}

var ErrConflictURL = errors.New("conflict long url")

func (s *ShortenerService) GenerateShortURLWithContext(context context.Context, longURL string) (string, error) {
	longURL, err := sanitizeURL(longURL)
	if err != nil {
		return "", err
	}
	shortURL, err := s.Generator.Generate()
	if err != nil {
		return "", errors.New("cannot generate new url")
	}
	err = s.URLStorage.StoreWithContext(context, longURL, shortURL)
	if errors.Is(err, urlstorage.ErrConflictURL) {
		shortURL, err := s.URLStorage.GetShortURLWithContext(context, longURL)
		if err == nil {
			return shortURL, ErrConflictURL
		}
	}

	if err != nil {
		return "", errors.New("cannot save new url")
	}

	return shortURL, nil
}

func (s *ShortenerService) GetLongURLWithContext(context context.Context, shortURL string) (string, error) {
	longURL, err := s.URLStorage.GetLongURLWithContext(context, shortURL)
	if err != nil {
		return "", errors.New("no such short url")
	}
	return longURL, nil
}

func (s *ShortenerService) GenerateShortURLBatchWithContext(context context.Context, longURLs []string) ([]string, error) {
	var shortURLs []string
	urls2Store := make(map[string]string)
	for _, longURL := range longURLs {
		longURL, err := sanitizeURL(longURL)
		if err != nil {
			return []string{}, err
		}

		if shortURL, err := s.Generator.Generate(); err == nil {
			urls2Store[longURL] = shortURL
			shortURLs = append(shortURLs, shortURL)
		}
	}
	errs, err := s.URLStorage.StoreManyWithContext(context, urls2Store)
	if err != nil {
		return []string{}, err
	}
	for i, longURL := range longURLs {
		if errs[i] != nil {
			shortURL, err := s.URLStorage.GetShortURLWithContext(context, longURL)
			if errors.Is(err, urlstorage.ErrConflictURL) {
				shortURLs[i] = shortURL
			} else {
				shortURLs[i] = "error when saving"
			}
		}
	}
	return shortURLs, nil
}

func (s *ShortenerService) Ping() error {
	return s.URLStorage.Ping()
}

func NewShortenerService(storage urlstorage.URLStorage, generator shortcutgenerator.ShortCutGenerator) *ShortenerService {
	return &ShortenerService{
		URLStorage: storage,
		Generator:  generator,
	}
}
