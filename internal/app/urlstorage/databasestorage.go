package urlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

type DatabaseStorage struct {
	DB *sql.DB
}

func (s *DatabaseStorage) Init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE shortener("user_id" TEXT, "short_url" TEXT,"long_url" TEXT)`)
	tx.Exec(`CREATE INDEX user_id_index ON shortener USING btree(user_id)`)
	tx.Exec(`CREATE UNIQUE INDEX long_url_index ON shortener USING btree(long_url)`)
	return tx.Commit()
}

func (s *DatabaseStorage) GetLongURLWithContext(ctx context.Context, shortURL string) (string, error) {
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
	row := s.DB.QueryRowContext(ctx,
		"SELECT short_url FROM shortener WHERE long_url = $1", longURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *DatabaseStorage) StoreWithContext(ctx context.Context, longURL string, shortURL string, userID string) error {
	_, err := s.DB.ExecContext(ctx,
		"INSERT into shortener (user_id, short_url, long_url) VALUES($1, $2, $3)", userID, shortURL, longURL)
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		err = ErrConflictURL
	}
	return err
}

func (s *DatabaseStorage) StoreManyWithContext(ctx context.Context, long2ShortUrls []utils.URLPair, userID string) ([]error, error) {
	var errs []error
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO shortener (user_id, short_url, long_url) VALUES($1, $2, $3) ON CONFLICT do nothing")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i := range long2ShortUrls {
		res, err := stmt.ExecContext(ctx, userID, long2ShortUrls[i].Short, long2ShortUrls[i].Long)
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

func (s *DatabaseStorage) GetUserURLs(ctx context.Context, userID string) ([]utils.URLPair, error) {
	var res []utils.URLPair
	rows, err := s.DB.QueryContext(ctx,
		"SELECT short_url, long_url FROM shortener WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		var userURL utils.URLPair
		err = rows.Scan(&userURL.Short, &userURL.Long)
		if err != nil {
			return nil, err
		}

		res = append(res, userURL)
	}

	return res, nil
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
