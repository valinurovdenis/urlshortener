package userstorage

import (
	"context"
	"sync/atomic"
)

// Generates new user ids by atomic increment.
type SimpleUserStorage struct {
	ID int64
}

// New atomic user storage.
func NewSimpleUserStorage() *SimpleUserStorage {
	return &SimpleUserStorage{
		ID: 0,
	}
}

// Generates uuid for new user with no collision.
func (s *SimpleUserStorage) GenerateUUID(_ context.Context) (int64, error) {
	resID := atomic.AddInt64(&s.ID, 1)
	return resID, nil
}
