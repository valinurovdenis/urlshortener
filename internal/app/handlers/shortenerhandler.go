package handlers

import (
	"encoding/json"
	"errors"
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
	url, err := h.Service.GetLongURLWithContext(r.Context(), shortURL)
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
		url, err = h.Service.GenerateShortURLWithContext(r.Context(), string(rawURL))
	}

	if err == nil {
		w.WriteHeader(http.StatusCreated)
	} else if errors.Is(err, service.ErrConflictURL) {
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err = h.Service.GenerateShortURLWithContext(r.Context(), longURL.URL)

	if err == nil {
		w.WriteHeader(http.StatusCreated)
	} else if errors.Is(err, service.ErrConflictURL) {
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultURL{h.Host + shortURL})
}

type InputBatch struct {
	URL string `json:"original_url"`
	ID  string `json:"correlation_id"`
}

type ResultBatch struct {
	URL string `json:"result_url"`
	ID  string `json:"correlation_id"`
}

func (h *ShortenerHandler) GenerateBatch(w http.ResponseWriter, r *http.Request) {
	var inputBatch []InputBatch
	var resultBatch []ResultBatch
	var longURLs []string
	if err := json.NewDecoder(r.Body).Decode(&inputBatch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, v := range inputBatch {
		longURLs = append(longURLs, v.URL)
	}
	shortURLs, err := h.Service.GenerateShortURLBatchWithContext(r.Context(), longURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i, shortURL := range shortURLs {
		resultBatch = append(resultBatch, ResultBatch{ID: inputBatch[i].ID, URL: h.Host + shortURL})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resultBatch)
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
	r.Post("/api/shorten/batch", handler.GenerateBatch)
	r.Get("/{url}", handler.Redirect)
	r.Get("/ping", handler.Ping)
	return r
}
