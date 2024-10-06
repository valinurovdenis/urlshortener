package main

import (
	"flag"
)

type Config struct {
	LocalURL    string
	BaseURL     string
	ShortLength int
}

func parseFlags(config *Config) {
	flag.StringVar(&config.LocalURL, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "address to return before short url")
	flag.IntVar(&config.ShortLength, "l", 8, "length of short url")
	flag.Parse()
}
