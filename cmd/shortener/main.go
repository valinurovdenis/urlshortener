package main

import "github.com/valinurovdenis/urlshortener/cmd/shortener/runner"

func main() {
	if err := runner.Run(); err != nil {
		panic(err)
	}
}
