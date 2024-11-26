package userstorage

import (
	"context"
	"sync/atomic"
)

type SimpleUserStorage struct {
	ID int64
}

func (s *SimpleUserStorage) GenerateUUID(_ context.Context) (int64, error) {
	resID := atomic.AddInt64(&s.ID, 1)
	return resID, nil
}

func NewSimpleUserStorage() *SimpleUserStorage {
	return &SimpleUserStorage{
		ID: 0,
	}
}
