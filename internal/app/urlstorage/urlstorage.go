package urlstorage

import "context"

//go:generate mockery --name URLStorage
type URLStorage interface {
	GetLongURLWithContext(context context.Context, shortURL string) (string, error)

	GetShortURLWithContext(context context.Context, longURL string) (string, error)

	StoreWithContext(context context.Context, longURL string, shortURL string) error

	StoreManyWithContext(context context.Context, long2ShortUrls map[string]string) error

	Clear() error

	Ping() error
}
