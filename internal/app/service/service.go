package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/utils"
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

func (s *ShortenerService) GenerateShortURLWithContext(context context.Context, longURL string, userID string) (string, error) {
	longURL, err := sanitizeURL(longURL)
	if err != nil {
		return "", err
	}

	shortURL, err := s.Generator.Generate()
	if err != nil {
		return "", fmt.Errorf("cannot generate new url: %w", err)
	}
	err = s.URLStorage.StoreWithContext(context, longURL, shortURL, userID)
	if errors.Is(err, urlstorage.ErrConflictURL) {
		shortURL, err := s.URLStorage.GetShortURLWithContext(context, longURL)
		if err == nil {
			return shortURL, ErrConflictURL
		}
	}

	if err != nil {
		return "", fmt.Errorf("cannot save new url: %w", err)
	}

	return shortURL, nil
}

func (s *ShortenerService) GetLongURLWithContext(context context.Context, shortURL string) (string, error) {
	longURL, err := s.URLStorage.GetLongURLWithContext(context, shortURL)
	if err != nil {
		return "", fmt.Errorf("no such short url: %w", err)
	}
	return longURL, nil
}

func (s *ShortenerService) GenerateShortURLBatchWithContext(context context.Context, longURLs []string, userID string) ([]string, error) {
	var shortURLs []string
	var urls2Store []utils.URLPair
	for _, longURL := range longURLs {
		longURL, err := sanitizeURL(longURL)
		if err != nil {
			return []string{}, err
		}

		shortURL, err := s.Generator.Generate()
		if err != nil {
			return nil, errors.New("cannot generate new short url")
		}
		userURL := utils.URLPair{Short: shortURL, Long: longURL}
		urls2Store = append(urls2Store, userURL)
		shortURLs = append(shortURLs, shortURL)
	}
	errs, err := s.URLStorage.StoreManyWithContext(context, urls2Store, userID)
	if err != nil {
		return []string{}, err
	}
	for i, longURL := range longURLs {
		if errs[i] != nil {
			shortURL, err := s.URLStorage.GetShortURLWithContext(context, longURL)
			if err == nil {
				shortURLs[i] = shortURL
			} else {
				shortURLs[i] = ""
			}
		}
	}
	return shortURLs, nil
}

func (s *ShortenerService) Ping() error {
	return s.URLStorage.Ping()
}

func (s *ShortenerService) GetUserURLs(context context.Context, userID string) ([]utils.URLPair, error) {
	return s.URLStorage.GetUserURLs(context, userID)
}

func NewShortenerService(storage urlstorage.URLStorage, generator shortcutgenerator.ShortCutGenerator) *ShortenerService {
	return &ShortenerService{
		URLStorage: storage,
		Generator:  generator,
	}
}
