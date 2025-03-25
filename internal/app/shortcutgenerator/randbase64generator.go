package shortcutgenerator

import (
	"math/rand"
)

type RandBase64Generator struct {
	Length int
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (s RandBase64Generator) Generate() (string, error) {
	b := make([]rune, s.Length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b), nil
}

func NewRandBase64Generator(length int) *RandBase64Generator {
	return &RandBase64Generator{Length: length}
}
