package urlstorage

//go:generate mockery --name URLStorage
type URLStorage interface {
	Get(shortURL string) (string, error)
	Store(url string) (string, error)
}
