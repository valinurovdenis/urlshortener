package main

import (
	"fmt"

	"github.com/valinurovdenis/urlshortener/cmd/shortener/runner"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := runner.Run(); err != nil {
		panic(err)
	}
}
