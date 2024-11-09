package urlstorage

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
)

type DatabaseStorage struct {
	DB    *sql.DB
	Mutex sync.Mutex
}

func (s *DatabaseStorage) Init() error {
	_, err := s.DB.Exec(`CREATE TABLE shortener("short_url" TEXT,"long_url" TEXT)`)
	return err
}

func (s *DatabaseStorage) GetLongURLWithContext(ctx context.Context, shortURL string) (string, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT long_url FROM videos WHERE short_url = ?", shortURL)
	var longURL string
	err := row.Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}

func (s *DatabaseStorage) GetShortURLWithContext(ctx context.Context, longURL string) (string, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT short_url FROM videos WHERE long_url = ?", longURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *DatabaseStorage) StoreWithContext(ctx context.Context, longURL string, shortURL string) error {
	_, err := s.DB.ExecContext(ctx,
		"INSERT into shortener (short_url, long_url) VALUES($1, $2)", shortURL, longURL)
	return err
}

func (s *DatabaseStorage) StoreMany(long2ShortUrls map[string]string) error {
	return errors.New("not implemented")
}

func (s *DatabaseStorage) Clear() error {
	return errors.New("not implemented")
}

func (s *DatabaseStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}

func NewDatabaseStorage(db *sql.DB) *DatabaseStorage {
	ret := &DatabaseStorage{DB: db}
	ret.Init()
	return ret
}
