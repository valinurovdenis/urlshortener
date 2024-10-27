package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
)

type ShortenerHandler struct {
	Service service.ShortenerService
	Host    string
}

func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "url")
	url, err := h.Service.GetLongURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var rawURL []byte
	var url string
	var err error
	rawURL, err = io.ReadAll(r.Body)
	if err == nil {
		url, err = h.Service.GenerateShortURL(string(rawURL))
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Host + url))
}

func NewShortenerHandler(service service.ShortenerService, host string) *ShortenerHandler {
	return &ShortenerHandler{Service: service, Host: host}
}

func ShortenerRouter(handler ShortenerHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Post("/", handler.Generate)
	r.Get("/{url}", handler.Redirect)
	return r
}
