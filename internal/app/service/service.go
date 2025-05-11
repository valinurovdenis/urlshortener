// Package service contains main shortener processing service.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"go.uber.org/zap"
)

// Sanitizes url to fixed form.
func SanitizeURL(origURL string) (string, error) {
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

// Interface for shortner service methods.
type ShortenerService interface {
	// Generating short url.
	GenerateShortURLWithContext(context context.Context, longURL string, userID string) (string, error)
	// Get long url from short.
	GetLongURLWithContext(context context.Context, shortURL string) (string, error)
	// Generate short url in batch mode.
	GenerateShortURLBatchWithContext(context context.Context, longURLs []string, userID string) ([]string, error)
	// Returns all user urls.
	GetUserURLs(context context.Context, userID string) ([]urlstorage.URLPair, error)
	// Deletes all user urls.
	DeleteUserURLs(ctx context.Context, userID string, shortURLs ...string) error
	// Get service stats.
	GetStats(ctx context.Context) (urlstorage.StorageStats, error)
	// Check whether service is alive.
	Ping() error
}

// Facade for service datastorages and generators.
type ShortenerServiceImpl struct {
	URLStorage     urlstorage.URLStorage
	UserURLStorage urlstorage.UserURLStorage
	Generator      shortcutgenerator.ShortCutGenerator
	deleteChan     chan urlstorage.URLsForDelete
	Stop           func()
	Stopped        chan struct{}
}

// New shortener service that facades storages and short url generator.
func NewShortenerService(storage urlstorage.URLStorage, userStorage urlstorage.UserURLStorage, generator shortcutgenerator.ShortCutGenerator) *ShortenerServiceImpl {
	ctx, stop := context.WithCancel(context.Background())
	ret := &ShortenerServiceImpl{
		URLStorage:     storage,
		UserURLStorage: userStorage,
		Generator:      generator,
		deleteChan:     make(chan urlstorage.URLsForDelete, 1024),
		Stop:           stop,
		Stopped:        make(chan struct{}, 1),
	}
	go ret.FlushDeletedUserURLs(ctx)
	return ret
}

// Generates shortURL from longURL for given user.
func (s ShortenerServiceImpl) GenerateShortURLWithContext(context context.Context, longURL string, userID string) (string, error) {
	longURL, err := SanitizeURL(longURL)
	if err != nil {
		return "", err
	}

	shortURL, err := s.Generator.Generate()
	if err != nil {
		return "", fmt.Errorf("cannot generate new url: %w", err)
	}
	err = s.URLStorage.StoreWithContext(context, longURL, shortURL, userID)
	if errors.Is(err, urlstorage.ErrConflictURL) {
		existingShortURL, errGet := s.URLStorage.GetShortURLWithContext(context, longURL)
		if errGet == nil {
			return existingShortURL, urlstorage.ErrConflictURL
		}
	}

	if err != nil {
		return "", fmt.Errorf("cannot save new url: %w", err)
	}

	return shortURL, nil
}

// Error in case of long url is already deleted.
var ErrDeletedURL = errors.New("conflict long url")

// Gets longURL from shortURL.
func (s ShortenerServiceImpl) GetLongURLWithContext(context context.Context, shortURL string) (string, error) {
	longURL, err := s.URLStorage.GetLongURLWithContext(context, shortURL)
	if errors.Is(err, urlstorage.ErrDeletedURL) {
		return "", ErrDeletedURL
	}
	if err != nil {
		return "", fmt.Errorf("no such short url: %w", err)
	}
	return longURL, nil
}

// Generates batch of shortURLs for user.
func (s ShortenerServiceImpl) GenerateShortURLBatchWithContext(context context.Context, longURLs []string, userID string) ([]string, error) {
	var shortURLs []string
	var urls2Store []urlstorage.URLPair
	for _, longURL := range longURLs {
		sanitizedLongURL, err := SanitizeURL(longURL)
		if err != nil {
			return []string{}, err
		}

		shortURL, err := s.Generator.Generate()
		if err != nil {
			return nil, errors.New("cannot generate new short url")
		}
		userURL := urlstorage.URLPair{Short: shortURL, Long: sanitizedLongURL}
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

// Deletes given user urls.
func (s ShortenerServiceImpl) DeleteUserURLs(ctx context.Context, userID string, shortURLs ...string) error {
	urls := urlstorage.URLsForDelete{UserID: userID, ShortURLs: shortURLs}
	s.deleteChan <- urls
	return nil
}

// Collects urls for deleting.
// Calls deleting function for collected urls every 10 seconds.
func (s ShortenerServiceImpl) FlushDeletedUserURLs(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	var urlsByUser []urlstorage.URLsForDelete

	for {
		select {
		case <-ctx.Done():
			if len(urlsByUser) != 0 {
				if err := s.UserURLStorage.DeleteUserURLs(context.TODO(), urlsByUser...); err != nil {
					logger.Log.Error("cannot delete urls", zap.Error(err))
				}
			}
			close(s.Stopped)
			return
		case urls := <-s.deleteChan:
			urlsByUser = append(urlsByUser, urls)
		case <-ticker.C:
			if len(urlsByUser) == 0 {
				continue
			}
			err := s.UserURLStorage.DeleteUserURLs(context.TODO(), urlsByUser...)
			if err != nil {
				logger.Log.Error("cannot delete urls", zap.Error(err))
				continue
			}
			urlsByUser = nil
		}
	}
}

// Check whether service is alive.
func (s ShortenerServiceImpl) Ping() error {
	err := s.UserURLStorage.Ping()
	if err == nil {
		return s.URLStorage.Ping()
	}
	return err
}

// Returns all user urls.
func (s ShortenerServiceImpl) GetUserURLs(context context.Context, userID string) ([]urlstorage.URLPair, error) {
	return s.UserURLStorage.GetUserURLs(context, userID)
}

// Get service stats.
func (s ShortenerServiceImpl) GetStats(ctx context.Context) (urlstorage.StorageStats, error) {
	return s.UserURLStorage.GetStats(ctx)
}
