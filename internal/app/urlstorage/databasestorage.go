package urlstorage

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseStorage struct {
	DB    *sql.DB
	Mutex sync.Mutex
}

func (s *DatabaseStorage) Init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE shortener("short_url" TEXT,"long_url" TEXT)`)
	tx.Exec(`CREATE UNIQUE INDEX long_url_index ON shortener USING btree(long_url)`)
	return tx.Commit()
}

func (s *DatabaseStorage) GetLongURLWithContext(ctx context.Context, shortURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	row := s.DB.QueryRowContext(ctx,
		"SELECT long_url FROM shortener WHERE short_url = $1", shortURL)
	var longURL string
	err := row.Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}

func (s *DatabaseStorage) GetShortURLWithContext(ctx context.Context, longURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	row := s.DB.QueryRowContext(ctx,
		"SELECT short_url FROM shortener WHERE long_url = $1", longURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *DatabaseStorage) StoreWithContext(ctx context.Context, longURL string, shortURL string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	_, err := s.DB.ExecContext(ctx,
		"INSERT into shortener (short_url, long_url) VALUES($1, $2)", shortURL, longURL)
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		err = ErrConflictURL
	}
	return err
}

func (s *DatabaseStorage) StoreManyWithContext(ctx context.Context, long2ShortUrls map[string]string) ([]error, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	var errs []error
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO shortener (short_url, long_url) VALUES($1, $2) ON CONFLICT do nothing")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for longURL, shortURL := range long2ShortUrls {
		res, err := stmt.ExecContext(ctx, shortURL, longURL)
		if c, _ := res.RowsAffected(); c == 0 {
			err = ErrConflictURL
		}
		errs = append(errs, err)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return errs, nil
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
