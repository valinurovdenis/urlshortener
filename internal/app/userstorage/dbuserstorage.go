package userstorage

import (
	"context"
	"database/sql"
)

// Generates new user id by autoincrement in postgresql.
type DatabaseUserStorage struct {
	DB *sql.DB
}

// New postgresql user storage.
func NewDatabaseUserStorage(db *sql.DB) *DatabaseUserStorage {
	ret := &DatabaseUserStorage{DB: db}
	ret.init()
	return ret
}

// Create all tables if needed.
func (s *DatabaseUserStorage) init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE user_id("id" SERIAL)`)
	return tx.Commit()
}

// Generates uuid for new user with no collision.
func (s *DatabaseUserStorage) GenerateUUID(ctx context.Context) (int64, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx, "INSERT into user_id DEFAULT VALUES RETURNING id").Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
