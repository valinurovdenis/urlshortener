package urlstorage

import (
	"context"
	"errors"
	"sync"
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

func (s *SimpleMapLockStorage) StoreWithContext(_ context.Context, longURL string, shortURL string) error {
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

func (s *SimpleMapLockStorage) StoreManyWithContext(_ context.Context, long2ShortUrls map[string]string) ([]error, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	var errs []error
	for longURL, shortURL := range long2ShortUrls {
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

func NewSimpleMapLockStorage() *SimpleMapLockStorage {
	return &SimpleMapLockStorage{
		ShortURL2Url: make(map[string]string),
		URL2ShortURL: make(map[string]string)}
}
