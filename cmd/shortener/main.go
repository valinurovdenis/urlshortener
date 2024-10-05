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
	shortLength := 8
	host := "http://localhost:8080/"
	generator := shortcutgenerator.NewRandBase64Generator(shortLength)
	storage := urlstorage.NewSimpleMapLockStorage(generator)
	handler := handlers.NewShortenerHandler(storage, host)
	return http.ListenAndServe(":8080", handlers.ShortenerRouter(*handler))
}
