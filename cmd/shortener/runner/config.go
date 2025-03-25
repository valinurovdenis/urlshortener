package runner

import (
	"flag"
	"os"
	"reflect"
)

// Struct contains all service settings.
//
// Settings read both from env and from args.
type Config struct {
	LocalURL     string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
	LogLevel     string `env:"LOG_LEVEL"`
	FileStorage  string `env:"FILE_STORAGE_PATH"`
	Database     string `env:"DATABASE_DSN"`
	SecretKey    string `env:"SECRET_KEY"`
	ShortLength  int
	IsProduction bool
}

func parseFlags(config *Config) {
	flag.StringVar(&config.LocalURL, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "address to return before short url")
	flag.IntVar(&config.ShortLength, "c", 8, "length of short url")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStorage, "f", "", "file storage path")
	flag.StringVar(&config.Database, "d", "", "database address")
	flag.StringVar(&config.SecretKey, "k", "SECRET_KEY", "secret key")
	flag.BoolVar(&config.IsProduction, "p", false, "is production")
	flag.Parse()
}

func (config *Config) updateFromEnv() {
	v := reflect.Indirect(reflect.ValueOf(config))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var envName string
		if envName = field.Tag.Get("env"); envName == "" {
			continue
		}
		if envVal := os.Getenv(envName); envVal != "" {
			v.Field(i).SetString(envVal)
		}
	}
}
