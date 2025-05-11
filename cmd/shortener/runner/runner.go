// Package runner for running service with given config.
package runner

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/acme/autocert"

	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/ipchecker"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
)

// Runs shortener service with given config.
func Run(ctx context.Context, stopped chan struct{}) error {
	config := GetConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	var urlStorage urlstorage.URLStorage
	var userURLStorage urlstorage.UserURLStorage
	var userStorage userstorage.UserStorage
	if config.Database != "" {
		db, err := sql.Open("pgx", config.Database)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		storage := urlstorage.NewDatabaseStorage(db)
		urlStorage = storage
		userURLStorage = storage
		userStorage = userstorage.NewDatabaseUserStorage(db)
	} else {
		storage := urlstorage.NewSimpleMapLockStorage()
		urlStorage = storage
		userURLStorage = storage
		userStorage = userstorage.NewSimpleUserStorage()
		if config.FileStorage != "" {
			fileStorageWrapper, err := urlstorage.NewFileDumpWrapper(
				config.FileStorage, storage)
			if err != nil {
				return err
			}
			fileStorageWrapper.RestoreFromDump()
			urlStorage = fileStorageWrapper
		}
	}

	generator := shortcutgenerator.NewRandBase64Generator(config.ShortLength)
	service := service.NewShortenerService(urlStorage, userURLStorage, generator)
	auth := auth.NewAuthenticator(config.SecretKey, userStorage)
	ipchecker := ipchecker.NewIPChecker(config.TrustedSubnet)
	handler := handlers.NewShortenerHandler(*service, *auth, config.BaseURL+"/", *ipchecker)

	router := handlers.ShortenerRouter(*handler, config.IsProduction)
	var srv *http.Server
	if config.EnableHTTPS {
		dir, _ := os.Getwd()
		mgr := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.BaseURL),
			Cache:      autocert.DirCache(dir),
		}
		srv = &http.Server{
			Addr:      config.LocalURL,
			Handler:   router,
			TLSConfig: mgr.TLSConfig(),
		}
	} else {
		srv = &http.Server{
			Addr:    config.LocalURL,
			Handler: router,
		}
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Error when shutdown server: %v", err)
		}
		service.Stop()
		<-service.Stopped
		close(stopped)
	}()

	if config.EnableHTTPS {
		return srv.ListenAndServeTLS("", "")
	}
	return srv.ListenAndServe()
}
