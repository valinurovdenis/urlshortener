// Package contains string generator.
package shortcutgenerator

//go:generate mockery --name ShortCutGenerator
type ShortCutGenerator interface {
	Generate() (string, error)
}
