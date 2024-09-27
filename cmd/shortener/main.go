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
	mux := http.NewServeMux()
	generator := shortcutgenerator.RandBase64Generator{Length: 8}
	storage := urlstorage.MakeSimpleMapLockStorage(generator)
	handler := handlers.ShortenerHandler{
		Storage: storage,
		Host:    "http://localhost:8080/"}
	mux.Handle("/", &handler)
	return http.ListenAndServe(`:8080`, mux)
}
