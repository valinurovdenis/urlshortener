package userstorage

import (
	"context"
)

//go:generate mockery --name UserStorage
type UserStorage interface {
	GenerateUUID(context context.Context) (int64, error)
}
