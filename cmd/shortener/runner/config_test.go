package runner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_updateFromEnv(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", "not_localhost")
	config := new(Config)
	parseFlags(config)
	require.Equal(t, "localhost:8080", config.LocalURL)
	config.updateFromEnv()
	require.Equal(t, "not_localhost", config.LocalURL)
}
