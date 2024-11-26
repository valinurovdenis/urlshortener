package urlstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

type SimpleMapLockStorage struct {
	ShortURL2Url map[string]string
	URL2ShortURL map[string]string
	Mutex        sync.Mutex
}

func (s *SimpleMapLockStorage) GetLongURLWithContext(_ context.Context, shortURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	val, has := s.ShortURL2Url[shortURL]
	if !has {
		return "", errors.New("no such shortUrl")
	} else {
		return val, nil
	}
}

func (s *SimpleMapLockStorage) GetShortURLWithContext(_ context.Context, longURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	val, has := s.URL2ShortURL[longURL]
	if !has {
		return "", errors.New("no such longUrl")
	} else {
		return val, nil
	}
}

func (s *SimpleMapLockStorage) StoreWithContext(_ context.Context, longURL string, shortURL string, _ string) error {
	if shortURL == "" {
		return errors.New("cannot save empty url")
	}
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	_, has := s.URL2ShortURL[longURL]
	if has {
		return ErrConflictURL
	}

	s.ShortURL2Url[shortURL] = longURL
	s.URL2ShortURL[longURL] = shortURL
	return nil
}

func (s *SimpleMapLockStorage) StoreManyWithContext(_ context.Context, long2ShortUrls []utils.URLPair, _ string) ([]error, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	var errs []error
	for i := range long2ShortUrls {
		longURL := long2ShortUrls[i].Long
		shortURL := long2ShortUrls[i].Short
		if shortURL == "" {
			continue
		}
		_, has := s.URL2ShortURL[longURL]
		if has {
			errs = append(errs, ErrConflictURL)
		} else {
			s.ShortURL2Url[shortURL] = longURL
			s.URL2ShortURL[longURL] = shortURL
			errs = append(errs, nil)
		}
	}
	return errs, nil
}

func (s *SimpleMapLockStorage) Clear() error {
	s.ShortURL2Url = make(map[string]string)
	s.URL2ShortURL = make(map[string]string)
	return nil
}

func (s *SimpleMapLockStorage) Ping() error {
	return nil
}

func (s *SimpleMapLockStorage) GetUserURLs(ctx context.Context, userID string) ([]utils.URLPair, error) {
	return nil, errors.New("not implemented")
}

func (s *SimpleMapLockStorage) DeleteUserURLs(context context.Context, urls ...utils.URLsForDelete) error {
	return errors.New("not implemented")
}

func NewSimpleMapLockStorage() *SimpleMapLockStorage {
	return &SimpleMapLockStorage{
		ShortURL2Url: make(map[string]string),
		URL2ShortURL: make(map[string]string)}
}
