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

func (s *ShortenerService) GenerateShortURLWithContext(context context.Context, longURL string) (string, error) {
	longURL, err := sanitizeURL(longURL)
	if err != nil {
		return "", err
	}
	shortURL, err := s.URLStorage.GetShortURLWithContext(context, longURL)
	if err == nil {
		return shortURL, nil
	}
	if shortURL, err = s.Generator.Generate(); err == nil {
		err = s.URLStorage.StoreWithContext(context, longURL, shortURL)
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

func (s *ShortenerService) Ping() error {
	return s.URLStorage.Ping()
}

func NewShortenerService(storage urlstorage.URLStorage, generator shortcutgenerator.ShortCutGenerator) *ShortenerService {
	return &ShortenerService{
		URLStorage: storage,
		Generator:  generator,
	}
}
