// Package contains urls storage.
package urlstorage

import (
	"context"
	"errors"
)

// Error in case url already has been saved.
var ErrConflictURL = errors.New("conflicting long url")

// Error in case url has been already deleted.
var ErrDeletedURL = errors.New("url has been deleted")

// Empty URL
var ErrEmptyLongURL = errors.New("cannot save empty url")

// Auxiliary struct for mapping longURL <-> shortURL.
type URLPair struct {
	Short string
	Long  string
}

// Auxiliary struct for user urls for delete.
type URLsForDelete struct {
	UserID    string
	ShortURLs []string
}

// Storage that contains mapping longURL <-> shortURL.
//
//go:generate mockery --name URLStorage
type URLStorage interface {
	// Returns longURL from shortURL.
	GetLongURLWithContext(context context.Context, shortURL string) (string, error)

	// Returns shortURL from longURL.
	GetShortURLWithContext(context context.Context, longURL string) (string, error)

	// Adds mapping longURL -> shortURL.
	StoreWithContext(context context.Context, longURL string, shortURL string, userID string) error

	// Adds number of mappings longURL -> shortURL.
	StoreManyWithContext(context context.Context, long2ShortUrls []URLPair, userID string) ([]error, error)

	// Clear all mappings.
	Clear() error

	// Check whether storage alive.
	Ping() error
}

// Storage contains urls saved and deleted by user.
//
//go:generate mockery --name UserURLStorage
type UserURLStorage interface {
	// Returns all urls saved by user.
	GetUserURLs(context context.Context, userID string) ([]URLPair, error)

	// Deletes given urls previously saved by user.
	DeleteUserURLs(context context.Context, urls ...URLsForDelete) error

	// Clear all user urls.
	Clear() error

	// Check whether storage alive.
	Ping() error
}
