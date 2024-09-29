package shortcutgenerator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandBase64Generator_Generate(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "random4", length: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := RandBase64Generator{Length: tt.length}
			got, err := generator.Generate()
			require.NoError(t, err)
			assert.Equal(t, len(got), tt.length)
		})
	}
}
