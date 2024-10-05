package urlstorage

import (
	"errors"
	"sync"

	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
)

type SimpleMapLockStorage struct {
	ShortURL2Url map[string]string
	URL2ShortURL map[string]string
	Mutex        sync.Mutex
	Generator    shortcutgenerator.ShortCutGenerator
}

func (s *SimpleMapLockStorage) Get(shortURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	val, has := s.ShortURL2Url[shortURL]
	if !has {
		return "", errors.New("no such shortUrl")
	} else {
		return val, nil
	}
}

func (s *SimpleMapLockStorage) Store(url string) (string, error) {
	if url == "" {
		return "", errors.New("cannot save empty url")
	}
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	shortURL, stored := s.URL2ShortURL[url]
	if stored {
		return shortURL, nil
	}
	shortURL, err := s.Generator.Generate()
	if err != nil {
		return "", errors.New("cannot save new url")
	}
	s.ShortURL2Url[shortURL] = url
	s.URL2ShortURL[url] = shortURL
	return shortURL, nil
}

func NewSimpleMapLockStorage(generator shortcutgenerator.ShortCutGenerator) *SimpleMapLockStorage {
	return &SimpleMapLockStorage{
		Generator:    generator,
		ShortURL2Url: make(map[string]string),
		URL2ShortURL: make(map[string]string)}
}
