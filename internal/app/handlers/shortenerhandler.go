package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/gzip"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
)

type ShortenerHandler struct {
	Service service.ShortenerService
	Auth    auth.JwtAuthenticator
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
	userID := r.Header.Get("user_id")
	var rawURL []byte
	var url string
	var err error
	rawURL, err = io.ReadAll(r.Body)
	if err == nil {
		url, err = h.Service.GenerateShortURLWithContext(r.Context(), string(rawURL), userID)
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
	userID := r.Header.Get("user_id")
	var shortURL string
	var longURL inputURL
	err := json.NewDecoder(r.Body).Decode(&longURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err = h.Service.GenerateShortURLWithContext(r.Context(), longURL.URL, userID)

	w.Header().Set("Content-Type", "application/json")
	if err == nil {
		w.WriteHeader(http.StatusCreated)
	} else if errors.Is(err, service.ErrConflictURL) {
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(resultURL{h.Host + shortURL})
}

type InputBatch struct {
	URL string `json:"original_url"`
	ID  string `json:"correlation_id"`
}

type ResultBatch struct {
	URL string `json:"short_url"`
	ID  string `json:"correlation_id"`
}

func (h *ShortenerHandler) GenerateBatch(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user_id")
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
	shortURLs, err := h.Service.GenerateShortURLBatchWithContext(r.Context(), longURLs, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i, shortURL := range shortURLs {
		if shortURL != "" {
			resultBatch = append(resultBatch, ResultBatch{ID: inputBatch[i].ID, URL: h.Host + shortURL})
		}
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

type UserURL struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

func (h *ShortenerHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user_id")
	userURLs, err := h.Service.GetUserURLs(r.Context(), userID)
	var resultURLs []UserURL
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	for i := range userURLs {
		url := UserURL{ShortURL: h.Host + userURLs[i].Short, LongURL: userURLs[i].Long}
		resultURLs = append(resultURLs, url)
	}
	json.NewEncoder(w).Encode(resultURLs)
}

func NewShortenerHandler(service service.ShortenerService, auth auth.JwtAuthenticator, host string) *ShortenerHandler {
	return &ShortenerHandler{Service: service, Auth: auth, Host: host}
}

func ShortenerRouter(handler ShortenerHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(gzip.GzipMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Use(handler.Auth.MaybeWithAuth)
		r.Post("/", handler.Generate)
		r.Post("/api/shorten", handler.GenerateJSON)
		r.Post("/api/shorten/batch", handler.GenerateBatch)
		r.Get("/{url}", handler.Redirect)
		r.Get("/ping", handler.Ping)
	})

	r.With(handler.Auth.OnlyWithAuth).Get("/api/user/urls", handler.GetUserURLs)
	return r
}
