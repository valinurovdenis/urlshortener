package main

import (
	"net/http"

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
	generator := shortcutgenerator.NewRandBase64Generator(config.ShortLength)
	storage := urlstorage.NewSimpleMapLockStorage(generator)
	handler := handlers.NewShortenerHandler(storage, config.BaseURL+"/")
	return http.ListenAndServe(config.LocalURL, handlers.ShortenerRouter(*handler))
}
