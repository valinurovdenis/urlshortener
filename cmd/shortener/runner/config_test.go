package runner

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetConfig(t *testing.T) {
	f, _ := os.CreateTemp("", "sample")
	defer os.Remove(f.Name())

	t.Setenv("CONFIG", f.Name())
	jsonConf := map[string]string{"server_address": "json"}
	content, _ := json.Marshal(jsonConf)
	os.WriteFile(f.Name(), content, os.ModeAppend)

	t.Setenv("BASE_URL", "env")
	os.Args = append(os.Args, "-d", "flag")
	config := GetConfig()

	require.Equal(t, "json", config.LocalURL)
	require.Equal(t, "env", config.BaseURL)
	require.Equal(t, "flag", config.Database)
	require.Equal(t, false, config.EnableHTTPS)
}
