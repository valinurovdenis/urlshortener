package runner_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/valinurovdenis/urlshortener/cmd/shortener/runner"
)

type inputURL struct {
	URL string `json:"url"`
}

type resultURL struct {
	URL string `json:"result"`
}

// Example of shortener service usage.
func Example() {
	// Run service on 8080 port and wait some time.
	go runner.Run()
	time.Sleep(100 * time.Millisecond)

	// Client for sending requests to service.
	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// Encode longURL.
	var input bytes.Buffer
	json.NewEncoder(&input).Encode(inputURL{"https://www.youtube.com"})
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", &input)
	if err != nil {
		panic(err)
	}

	// Get shortURL.
	genResp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer genResp.Body.Close()

	// Decode shortURL.
	var getShortURL resultURL
	respBody, err := io.ReadAll(genResp.Body)
	json.NewDecoder(strings.NewReader(string(respBody))).Decode(&getShortURL)
	fmt.Println("Get short URL: ", getShortURL.URL)

	// Get redirect to longURL from shortURL.
	req, err = http.NewRequest(http.MethodGet, getShortURL.URL, nil)
	redirectResp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer redirectResp.Body.Close()
	fmt.Println("Redirect code:", redirectResp.StatusCode)
	fmt.Println("Redirected to long URL: ", redirectResp.Header.Get("Location"))
}
