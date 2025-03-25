package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/pprof"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
)

var (
	dbConnection string
	maxSize      int
	isBase       bool
)

func main() {
	flag.StringVar(&dbConnection, "d", "", "db connection")
	flag.IntVar(&maxSize, "s", 50000, "number of requests")
	flag.BoolVar(&isBase, "b", false, "base or result pprof")
	flag.Parse()

	var urlStorage urlstorage.URLStorage
	var userURLStorage urlstorage.UserURLStorage
	var userStorage userstorage.UserStorage
	if dbConnection != "" {
		db, err := sql.Open("pgx", dbConnection)
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
	}

	generator := shortcutgenerator.NewRandBase64Generator(8)
	service := service.NewShortenerService(urlStorage, userURLStorage, generator)
	auth := auth.NewAuthenticator("secret_benchmark", userStorage)
	handler := handlers.NewShortenerHandler(*service, *auth, "/")
	defer service.Stop()
	ts := httptest.NewServer(handlers.ShortenerRouter(*handler, false))

	c := make(chan struct{})
	go benchmarkGenerateAPI(ts, c)
	<-c
}

type inputURL struct {
	URL string `json:"url"`
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func benchmarkGenerateAPI(ts *httptest.Server, c chan struct{}) {
	path := "/api/shorten"
	generateErrors := 0
	answerErrors := 0

	for i := 0; i < maxSize; i++ {
		var input bytes.Buffer
		randomURL := randStringRunes(10) + ".com"
		json.NewEncoder(&input).Encode(inputURL{randomURL})
		req, _ := http.NewRequest(http.MethodPost, ts.URL+path, &input)
		resp, err := ts.Client().Do(req)
		if err != nil {
			answerErrors += 1
		}
		resp.Body.Close()
	}

	fmt.Printf("generateErrors: %d\nanswerErrors: %d", generateErrors, answerErrors)
	saveProfileToFile()
	c <- struct{}{}
}

func saveProfileToFile() {
	var fileName string
	if isBase {
		fileName = "base.pprof"
	} else {
		fileName = "result.pprof"
	}
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()

	pprof.Lookup("heap").WriteTo(f, 0)
}
