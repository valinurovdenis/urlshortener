package runner

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
)

// Struct contains all service settings.
//
// Settings read both from env and from args.
type Config struct {
	LocalURL      string `env:"SERVER_ADDRESS" json:"server_address"`
	LocalGrpcURL  string `env:"SERVER_GRPC_ADDRESS" json:"server_grpc_address"`
	BaseURL       string `env:"BASE_URL" json:"base_url"`
	LogLevel      string `env:"LOG_LEVEL"`
	FileStorage   string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Database      string `env:"DATABASE_DSN" json:"database_dsn"`
	SecretKey     string `env:"SECRET_KEY"`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	EnableGRPC    bool   `env:"ENABLE_GRPC" json:"enable_grpc"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	ShortLength   int
	IsProduction  bool
}

// Default config values.
var defaultConfig = Config{
	LocalURL:     "localhost:8080",
	LocalGrpcURL: "localhost:8081",
	BaseURL:      "http://localhost:8080",
	LogLevel:     "info",
	FileStorage:  "",
	Database:     "",
	SecretKey:    "SECRET_KEY",
	EnableHTTPS:  false,
	EnableGRPC:   true,
	ShortLength:  8,
	IsProduction: false,
}

// Parse command line flags.
func parseFlags(config *Config) {
	flag.StringVar(&config.LocalURL, "a", defaultConfig.LocalURL, "address and port to run server")
	flag.StringVar(&config.LocalGrpcURL, "c", defaultConfig.LocalGrpcURL, "address and port to run grpc server")
	flag.StringVar(&config.BaseURL, "b", defaultConfig.BaseURL, "address to return before short url")
	flag.IntVar(&config.ShortLength, "e", defaultConfig.ShortLength, "length of short url")
	flag.StringVar(&config.LogLevel, "l", defaultConfig.LogLevel, "log level")
	flag.StringVar(&config.FileStorage, "f", defaultConfig.FileStorage, "file storage path")
	flag.StringVar(&config.Database, "d", defaultConfig.Database, "database address")
	flag.StringVar(&config.SecretKey, "k", defaultConfig.SecretKey, "secret key")
	flag.BoolVar(&config.IsProduction, "p", defaultConfig.IsProduction, "is production")
	flag.BoolVar(&config.EnableHTTPS, "s", defaultConfig.EnableHTTPS, "is https enabled")
	flag.BoolVar(&config.EnableGRPC, "g", defaultConfig.EnableGRPC, "is grpc enabled")
	flag.StringVar(&config.TrustedSubnet, "t", defaultConfig.TrustedSubnet, "is https enabled")
	flag.Parse()
}

// Get config from env.
func updateFromEnv(config *Config) {
	v := reflect.Indirect(reflect.ValueOf(config))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var envName string
		if envName = field.Tag.Get("env"); envName == "" {
			continue
		}
		if envVal := os.Getenv(envName); envVal != "" {
			v.Field(i).SetString(envVal)
		}
	}
}

// Get config from file.
func updateDefaultFromConfigFile(configFile string) {
	if configFile == "" {
		return
	}
	file, err := os.ReadFile(configFile)
	if err != nil {
		log.Println("Cannot read config file")
		return
	}
	err = json.Unmarshal(file, &defaultConfig)
	if err != nil {
		log.Println("Wrong json config")
		return
	}
}

// Singleton variable.
var shortnerConfig *Config = nil

// Get overall config.
func GetConfig() Config {
	if shortnerConfig == nil {
		var configFile string
		if configFile = os.Getenv("CONFIG"); configFile == "" {
			flagSetJSON := flag.NewFlagSet("json", flag.ContinueOnError)
			flagSetJSON.StringVar(&configFile, "c", "", "Json config")
			flagSetJSON.Parse(os.Args[1:])
		}
		updateDefaultFromConfigFile(configFile)
		var config Config
		parseFlags(&config)
		updateFromEnv(&config)
		shortnerConfig = &config
	}
	return *shortnerConfig
}
