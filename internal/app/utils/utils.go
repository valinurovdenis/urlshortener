package utils

import "bytes"

type URLPair struct {
	Short string
	Long  string
}

type URLsForDelete struct {
	UserID    string
	ShortURLs []string
}

func AddStrings(strings ...string) string {
	bufferLength := 0
	for _, s := range strings {
		bufferLength += len(s)
	}
	buffer := bytes.NewBuffer(make([]byte, 0, bufferLength))
	for _, s := range strings {
		buffer.WriteString(s)
	}
	return buffer.String()
}
