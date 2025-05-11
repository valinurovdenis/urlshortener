// Package handlers contains service handlers for chi router.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/gzip"
	"github.com/valinurovdenis/urlshortener/internal/app/ipchecker"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

// Main class for chi handlers.
type ShortenerHandler struct {
	Service   service.ShortenerService
	Auth      auth.JwtAuthenticator
	IPChecker ipchecker.IPChecker
	Host      string
}

// Shortener handler contains shortener service, authenticator.
func NewShortenerHandler(service service.ShortenerService, auth auth.JwtAuthenticator, host string, ipchecker ipchecker.IPChecker) *ShortenerHandler {
	return &ShortenerHandler{Service: service, Auth: auth, Host: host, IPChecker: ipchecker}
}

// Handler for redirecting to long url by short url.
func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "url")
	url, err := h.Service.GetLongURLWithContext(r.Context(), shortURL)
	if errors.Is(err, service.ErrDeletedURL) {
		w.WriteHeader(http.StatusGone)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Handler for generating short url from long url.
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
	} else if errors.Is(err, urlstorage.ErrConflictURL) {
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(utils.AddStrings(h.Host, url)))
}

// Input type for json handler.
type InputURL struct {
	URL string `json:"url"`
}

// Output type for json handler.
type ResultURL struct {
	URL string `json:"result"`
}

// Same as Generate, but in json format.
func (h *ShortenerHandler) GenerateJSON(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user_id")
	var shortURL string
	var longURL InputURL
	err := json.NewDecoder(r.Body).Decode(&longURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err = h.Service.GenerateShortURLWithContext(r.Context(), longURL.URL, userID)

	w.Header().Set("Content-Type", "application/json")
	if err == nil {
		w.WriteHeader(http.StatusCreated)
	} else if errors.Is(err, urlstorage.ErrConflictURL) {
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(ResultURL{utils.AddStrings(h.Host, shortURL)})
}

// Input type for generating batch.
type InputBatch struct {
	URL string `json:"original_url"`
	ID  string `json:"correlation_id"`
}

// Output type for generating batch.
type ResultBatch struct {
	URL string `json:"short_url"`
	ID  string `json:"correlation_id"`
}

// Handler for generating long urls in batch mode.
func (h *ShortenerHandler) GenerateBatch(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user_id")
	var input []InputBatch
	var result []ResultBatch
	var longURLs []string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, v := range input {
		longURLs = append(longURLs, v.URL)
	}
	shortURLs, err := h.Service.GenerateShortURLBatchWithContext(r.Context(), longURLs, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i, shortURL := range shortURLs {
		if shortURL != "" {
			result = append(result, ResultBatch{ID: input[i].ID, URL: h.Host + shortURL})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// Ping that service is still alive.
func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.Service.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// Main structure for pair mapping LongURL <-> ShortURL
type UserURL struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

// Get all urls saved by user.
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

// Delete urls saved by user.
func (h *ShortenerHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user_id")
	var urls []string
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := h.Service.DeleteUserURLs(r.Context(), userID, urls...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

// Delete urls saved by user.
func (h ShortenerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Service.GetStats(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// Defines all handlers.
func ShortenerRouter(handler ShortenerHandler, isProduction bool) chi.Router {
	r := chi.NewRouter()
	if !isProduction {
		r.Mount("/debug", middleware.Profiler())
	}
	r.With(handler.IPChecker.CheckFromInnerNetwork).Get("/api/internal/stats", handler.GetStats)

	r.Group(func(r chi.Router) {
		r.Use(logger.RequestLoggerMiddleware)
		r.Use(gzip.GzipMiddleware)
		r.Route("/", func(r chi.Router) {
			r.Use(handler.Auth.CreateUserIfNeeded)
			r.Post("/", handler.Generate)
			r.Post("/api/shorten", handler.GenerateJSON)
			r.Post("/api/shorten/batch", handler.GenerateBatch)
			r.Get("/{url}", handler.Redirect)
			r.Get("/ping", handler.Ping)
			r.Get("/api/user/urls", handler.GetUserURLs)
		})

		r.With(handler.Auth.OnlyWithAuth).Delete("/api/user/urls", handler.DeleteUserURLs)
	})

	return r
}
