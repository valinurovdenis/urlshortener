// Package runner for running service with given config.
package runner

import (
	"database/sql"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
)

// Runs shortener service with given config.
func Run() error {
	config := new(Config)
	parseFlags(config)
	config.updateFromEnv()

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
	handler := handlers.NewShortenerHandler(*service, *auth, config.BaseURL+"/")
	defer service.Stop()
	return http.ListenAndServe(config.LocalURL, handlers.ShortenerRouter(*handler, config.IsProduction))
}
