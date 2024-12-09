package userstorage

import (
	"context"
	"database/sql"
)

type DatabaseUserStorage struct {
	DB *sql.DB
}

func (s *DatabaseUserStorage) Init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE user_id("id" SERIAL)`)
	return tx.Commit()
}

func (s *DatabaseUserStorage) GenerateUUID(ctx context.Context) (int64, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx, "INSERT into user_id DEFAULT VALUES RETURNING id").Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func NewDatabaseUserStorage(db *sql.DB) *DatabaseUserStorage {
	ret := &DatabaseUserStorage{DB: db}
	ret.Init()
	return ret
}
