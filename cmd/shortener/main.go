package main

import (
	"net/http"
	"os"

	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	config := new(Config)
	parseFlags(config)
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		config.LocalURL = serverAddress
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	generator := shortcutgenerator.NewRandBase64Generator(config.ShortLength)
	storage := urlstorage.NewSimpleMapLockStorage(generator)
	handler := handlers.NewShortenerHandler(storage, config.BaseURL+"/")
	return http.ListenAndServe(config.LocalURL, handlers.ShortenerRouter(*handler))
}
