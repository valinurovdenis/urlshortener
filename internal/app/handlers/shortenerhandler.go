package handlers

import (
	"io"
	"net/http"

	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

type ShortenerHandler struct {
	Storage urlstorage.URLStorage
	Host    string
}

func (h *ShortenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.generate(w, r)
	} else if r.Method == http.MethodGet {
		h.redirect(w, r)
	} else {
		http.Error(w, "Not supported method", http.StatusBadRequest)
	}
}

func (h *ShortenerHandler) redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	url, err := h.Storage.Get(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *ShortenerHandler) generate(w http.ResponseWriter, r *http.Request) {
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
