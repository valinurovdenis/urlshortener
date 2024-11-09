package urlstorage

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

type DatabaseStorage struct {
	DB    *sql.DB
	Mutex sync.Mutex
}

func (s *DatabaseStorage) GetLongURL(shortURL string) (string, error) {
	return "", nil
}

func (s *DatabaseStorage) GetShortURL(longURL string) (string, error) {
	return "", nil
}

func (s *DatabaseStorage) Store(longURL string, shortURL string) error {
	return nil
}

func (s *DatabaseStorage) StoreMany(long2ShortUrls map[string]string) error {
	return nil
}

func (s *DatabaseStorage) Clear() error {
	return nil
}

func (s *DatabaseStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}

func NewDatabaseStorage(db *sql.DB) *DatabaseStorage {
	return &DatabaseStorage{DB: db}
}
