package urlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Storage storing urls in postgresql.
type DatabaseStorage struct {
	DB *sql.DB
}

// New postgresql storage.
func NewDatabaseStorage(db *sql.DB) *DatabaseStorage {
	ret := &DatabaseStorage{DB: db}
	ret.init()
	return ret
}

// Creates all tables if needed.
func (s *DatabaseStorage) init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE shortener("user_id" TEXT, "short_url" TEXT, "long_url" TEXT, "deleted" BOOLEAN DEFAULT false)`)
	tx.Exec(`CREATE INDEX user_id_index ON shortener USING btree(user_id)`)
	tx.Exec(`CREATE UNIQUE INDEX long_url_index ON shortener USING btree(long_url)`)
	return tx.Commit()
}

// Returns longURL from shortURL.
func (s *DatabaseStorage) GetLongURLWithContext(ctx context.Context, shortURL string) (string, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT long_url, deleted FROM shortener WHERE short_url = $1", shortURL)
	var longURL string
	var deleted bool
	err := row.Scan(&longURL, &deleted)
	if err != nil {
		return "", err
	}
	if deleted {
		return "", ErrDeletedURL
	}
	return longURL, nil
}

// Returns shortURL from longURL.
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

// Adds mapping longURL -> shortURL.
func (s *DatabaseStorage) StoreWithContext(ctx context.Context, longURL string, shortURL string, userID string) error {
	if longURL == "" {
		return ErrEmptyLongURL
	}
	_, err := s.DB.ExecContext(ctx,
		"INSERT into shortener (user_id, short_url, long_url) VALUES($1, $2, $3)", userID, shortURL, longURL)
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		err = ErrConflictURL
	}
	return err
}

// Adds number of mappings longURL -> shortURL.
func (s *DatabaseStorage) StoreManyWithContext(ctx context.Context, long2ShortUrls []URLPair, userID string) ([]error, error) {
	var errs []error
	tx, errDB := s.DB.Begin()
	if errDB != nil {
		return nil, errDB
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO shortener (user_id, short_url, long_url) VALUES($1, $2, $3) ON CONFLICT do nothing")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i := range long2ShortUrls {
		res, errExec := stmt.ExecContext(ctx, userID, long2ShortUrls[i].Short, long2ShortUrls[i].Long)
		if c, _ := res.RowsAffected(); c == 0 {
			errExec = ErrConflictURL
		}
		errs = append(errs, errExec)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return errs, nil
}

// Returns all urls saved by user.
func (s *DatabaseStorage) GetUserURLs(ctx context.Context, userID string) ([]URLPair, error) {
	var res []URLPair
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
		var userURL URLPair
		err = rows.Scan(&userURL.Short, &userURL.Long)
		if err != nil {
			return nil, err
		}

		res = append(res, userURL)
	}

	return res, nil
}

// Deletes given urls previously saved by user.
func (s *DatabaseStorage) DeleteUserURLs(ctx context.Context, urlsByUser ...URLsForDelete) error {
	query :=
		`UPDATE shortener SET deleted=true WHERE user_id = $1 and short_url = $2`

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, urls := range urlsByUser {
		for _, url := range urls.ShortURLs {
			stmt.ExecContext(ctx, urls.UserID, url)
		}
	}

	return tx.Commit()
}

// Clear all mappings.
func (s *DatabaseStorage) Clear() error {
	return errors.New("not implemented")
}

// Storage contains urls saved and deleted by user.
func (s *DatabaseStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}
