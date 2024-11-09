package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/valinurovdenis/urlshortener/internal/app/gzip"
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

type inputURL struct {
	URL string `json:"url"`
}

type resultURL struct {
	URL string `json:"result"`
}

func (h *ShortenerHandler) GenerateJSON(w http.ResponseWriter, r *http.Request) {
	var shortURL string
	var longURL inputURL
	err := json.NewDecoder(r.Body).Decode(&longURL)
	if err == nil {
		shortURL, err = h.Service.GenerateShortURL(longURL.URL)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resultURL{h.Host + shortURL})
}

func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.Service.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func NewShortenerHandler(service service.ShortenerService, host string) *ShortenerHandler {
	return &ShortenerHandler{Service: service, Host: host}
}

func ShortenerRouter(handler ShortenerHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(gzip.GzipMiddleware)
	r.Post("/", handler.Generate)
	r.Post("/api/shorten", handler.GenerateJSON)
	r.Get("/{url}", handler.Redirect)
	r.Get("/ping", handler.Ping)
	return r
}
