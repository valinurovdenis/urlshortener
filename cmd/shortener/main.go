package main

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

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	config := new(Config)
	parseFlags(config)
	config.updateFromEnv()

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	var storage urlstorage.URLStorage
	var userStorage userstorage.UserStorage
	if config.Database != "" {
		db, err := sql.Open("pgx", config.Database)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		storage = urlstorage.NewDatabaseStorage(db)
		userStorage = userstorage.NewDatabaseUserStorage(db)
	} else {
		storage = urlstorage.NewSimpleMapLockStorage()
		userStorage = userstorage.NewSimpleUserStorage()
		if config.FileStorage != "" {
			fileStorageWrapper, err := urlstorage.NewFileDumpWrapper(
				config.FileStorage, storage)
			if err != nil {
				return err
			}
			fileStorageWrapper.RestoreFromDump()
			storage = fileStorageWrapper
		}
	}

	generator := shortcutgenerator.NewRandBase64Generator(config.ShortLength)
	service := service.NewShortenerService(storage, generator)
	auth := auth.NewAuthenticator(config.SecretKey, userStorage)
	handler := handlers.NewShortenerHandler(*service, *auth, config.BaseURL+"/")
	return http.ListenAndServe(config.LocalURL, handlers.ShortenerRouter(*handler))
}
