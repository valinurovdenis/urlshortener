package utils

type URLPair struct {
	Short string
	Long  string
}

type URLsForDelete struct {
	UserID    string
	ShortURLs []string
}
