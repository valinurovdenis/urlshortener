// Storage for storing and generating user ids.
package userstorage

import (
	"context"
)

// Storage can generate uuid for new user with no collision.
//
//go:generate mockery --name UserStorage
type UserStorage interface {
	// Method for geerating new user uuid.
	GenerateUUID(context context.Context) (int64, error)
}
