package urlstorage

//go:generate mockery --name URLStorage
type URLStorage interface {
	GetLongURL(shortURL string) (string, error)
	GetShortURL(longURL string) (string, error)
	Store(longURL string, shortURL string) error
	StoreMany(long2ShortUrls map[string]string) error
	Clear() error
}
