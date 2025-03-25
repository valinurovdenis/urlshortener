// Main shortener processing service.
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

// Facade for service datastorages and generators.
type ShortenerService struct {
	URLStorage     urlstorage.URLStorage
	UserURLStorage urlstorage.UserURLStorage
	Generator      shortcutgenerator.ShortCutGenerator
	deleteChan     chan urlstorage.URLsForDelete
	Stop           func()
}

// New shortener service that facades storages and short url generator.
func NewShortenerService(storage urlstorage.URLStorage, userStorage urlstorage.UserURLStorage, generator shortcutgenerator.ShortCutGenerator) *ShortenerService {
	ctx, stop := context.WithCancel(context.Background())
	ret := &ShortenerService{
		URLStorage:     storage,
		UserURLStorage: userStorage,
		Generator:      generator,
		deleteChan:     make(chan urlstorage.URLsForDelete, 1024),
		Stop:           stop,
	}
	go ret.flushDeletedUserURLs(ctx)
	return ret
}

// Error in case of long url is already saved.
var ErrConflictURL = errors.New("conflict long url")

// Generates shortURL from longURL for given user.
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

// Error in case of long url is already deleted.
var ErrDeletedURL = errors.New("conflict long url")

// Gets longURL from shortURL.
func (s *ShortenerService) GetLongURLWithContext(context context.Context, shortURL string) (string, error) {
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
func (s *ShortenerService) GenerateShortURLBatchWithContext(context context.Context, longURLs []string, userID string) ([]string, error) {
	var shortURLs []string
	var urls2Store []urlstorage.URLPair
	for _, longURL := range longURLs {
		longURL, err := sanitizeURL(longURL)
		if err != nil {
			return []string{}, err
		}

		shortURL, err := s.Generator.Generate()
		if err != nil {
			return nil, errors.New("cannot generate new short url")
		}
		userURL := urlstorage.URLPair{Short: shortURL, Long: longURL}
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
func (s *ShortenerService) DeleteUserURLs(ctx context.Context, userID string, shortURLs ...string) error {
	urls := urlstorage.URLsForDelete{UserID: userID, ShortURLs: shortURLs}
	s.deleteChan <- urls
	return nil
}

// Collects urls for deleting.
// Calls deleting function for collected urls every 10 seconds.
func (s *ShortenerService) flushDeletedUserURLs(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	var urlsByUser []urlstorage.URLsForDelete

	for {
		select {
		case <-ctx.Done():
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
func (s *ShortenerService) Ping() error {
	return s.URLStorage.Ping()
}

// Returns all user urls.
func (s *ShortenerService) GetUserURLs(context context.Context, userID string) ([]urlstorage.URLPair, error) {
	return s.UserURLStorage.GetUserURLs(context, userID)
}
