package main

import (
	"net/http"

	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
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
	config.updateFromEnv()

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	generator := shortcutgenerator.NewRandBase64Generator(config.ShortLength)
	storage := urlstorage.NewSimpleMapLockStorage()
	service := service.NewShortenerService(storage, generator)
	handler := handlers.NewShortenerHandler(*service, config.BaseURL+"/")
	return http.ListenAndServe(config.LocalURL, handlers.ShortenerRouter(*handler))
}
