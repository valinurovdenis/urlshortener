package urlstorage

import (
	"context"
	"errors"

	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

var ErrConflictURL = errors.New("conflicting long url")

//go:generate mockery --name URLStorage
type URLStorage interface {
	GetLongURLWithContext(context context.Context, shortURL string) (string, error)

	GetShortURLWithContext(context context.Context, longURL string) (string, error)

	StoreWithContext(context context.Context, longURL string, shortURL string, userID string) error

	StoreManyWithContext(context context.Context, long2ShortUrls []utils.URLPair, userID string) ([]error, error)

	GetUserURLs(context context.Context, userID string) ([]utils.URLPair, error)

	Clear() error

	Ping() error
}
