package shortcutgenerator

import (
	"math/rand"
)

// Generates random string from sequence of runs.
type RandBase64Generator struct {
	Length int
}

// New random generator. Generates strings of length `Length`.
func NewRandBase64Generator(length int) *RandBase64Generator {
	return &RandBase64Generator{Length: length}
}

// Sequence of runes to generate from.
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Generates random string of given length from above rune sequence.
func (s RandBase64Generator) Generate() (string, error) {
	b := make([]rune, s.Length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b), nil
}
