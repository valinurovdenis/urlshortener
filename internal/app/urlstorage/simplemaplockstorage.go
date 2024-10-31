package urlstorage

import (
	"errors"
	"sync"
)

type SimpleMapLockStorage struct {
	ShortURL2Url map[string]string
	URL2ShortURL map[string]string
	Mutex        sync.Mutex
}

func (s *SimpleMapLockStorage) GetLongURL(shortURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	val, has := s.ShortURL2Url[shortURL]
	if !has {
		return "", errors.New("no such shortUrl")
	} else {
		return val, nil
	}
}

func (s *SimpleMapLockStorage) GetShortURL(longURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	val, has := s.URL2ShortURL[longURL]
	if !has {
		return "", errors.New("no such longUrl")
	} else {
		return val, nil
	}
}

func (s *SimpleMapLockStorage) Store(longURL string, shortURL string) error {
	if shortURL == "" {
		return errors.New("cannot save empty url")
	}
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.ShortURL2Url[shortURL] = longURL
	s.URL2ShortURL[longURL] = shortURL
	return nil
}

func (s *SimpleMapLockStorage) StoreMany(long2ShortUrls map[string]string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for longURL, shortURL := range long2ShortUrls {
		if shortURL == "" {
			continue
		}
		s.ShortURL2Url[shortURL] = longURL
		s.URL2ShortURL[longURL] = shortURL
	}
	return nil
}

func (s *SimpleMapLockStorage) Clear() error {
	s.ShortURL2Url = make(map[string]string)
	s.URL2ShortURL = make(map[string]string)
	return nil
}

func NewSimpleMapLockStorage() *SimpleMapLockStorage {
	return &SimpleMapLockStorage{
		ShortURL2Url: make(map[string]string),
		URL2ShortURL: make(map[string]string)}
}
