package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

type ShortenerHandler struct {
	Storage urlstorage.URLStorage
	Host    string
}

func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "url")
	url, err := h.Storage.Get(shortURL)
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
		url, err = utils.SanitizeURL(string(rawURL))
		if err == nil {
			url, err = h.Storage.Store(url)
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Host + url))
}

func NewShortenerHandler(storage urlstorage.URLStorage, host string) *ShortenerHandler {
	return &ShortenerHandler{Storage: storage, Host: host}
}

func ShortenerRouter(handler ShortenerHandler) chi.Router {
	r := chi.NewRouter()
	r.Post("/", handler.Generate)
	r.Get("/{url}", handler.Redirect)
	return r
}
