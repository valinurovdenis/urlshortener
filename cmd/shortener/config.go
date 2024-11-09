package main

import (
	"flag"
	"os"
	"reflect"
)

type Config struct {
	LocalURL    string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	LogLevel    string `env:"LOG_LEVEL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	ShortLength int
}

func parseFlags(config *Config) {
	flag.StringVar(&config.LocalURL, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "address to return before short url")
	flag.IntVar(&config.ShortLength, "c", 8, "length of short url")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStorage, "f", "url_storage", "file storage path")
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
