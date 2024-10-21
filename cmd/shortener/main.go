package main

import (
	"net/http"

	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
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
	config.LocalURL = updateFromEnv("SERVER_ADDRESS", config.LocalURL)
	config.BaseURL = updateFromEnv("BASE_URL", config.BaseURL)
	generator := shortcutgenerator.NewRandBase64Generator(config.ShortLength)
	storage := urlstorage.NewSimpleMapLockStorage()
	service := service.NewShortenerService(storage, generator)
	handler := handlers.NewShortenerHandler(*service, config.BaseURL+"/")
	return http.ListenAndServe(config.LocalURL, handlers.ShortenerRouter(*handler))
}
