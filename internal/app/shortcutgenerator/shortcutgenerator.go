package shortcutgenerator

import (
	"crypto/rand"
	"encoding/base64"
)

type ShortCutGenerator interface {
	Generate() (string, error)
}

type RandBase64Generator struct {
	Length int
}

func (s RandBase64Generator) Generate() (string, error) {
	buffer := make([]byte, s.Length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buffer)[:s.Length], nil
}
