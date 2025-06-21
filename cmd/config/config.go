package config

import (
	"flag"
	"os"
	"strings"
)

const (
	hostFlagName  = "a"
	defaultPort   = ":8080"
	hostFlagUsage = "Address to launch the HTTP server"

	baseURLFlagName  = "b"
	defaultEndpoint  = "http://localhost:8080/"
	baseURLFlagUsage = "Base URL for shortened links"

	logLevelFlagName  = "l"
	defaultLogLevel   = "info"
	logLevelFlagUsage = "Log level"

	fileStorageFlagName    = "f"
	defaultFileStoragePath = "" // /tmp/short-url-db.json
	fileStorageFlagUsage

	databaseFlagName    = "d"
	defaultDatabasePath = "" //postgres://postgres:GfdGjc964@localhost:5432/videos
	databaseFlagUsage   = "Database path"
)

type Config struct {
	Address     string
	BaseURL     string
	LogLevel    string
	FileStorage string
	DB          string
}

func Init() Config {
	var cfg Config

	flag.StringVar(&cfg.Address, hostFlagName, defaultPort, hostFlagUsage)
	flag.StringVar(&cfg.BaseURL, baseURLFlagName, defaultEndpoint, baseURLFlagUsage)
	flag.StringVar(&cfg.LogLevel, logLevelFlagName, defaultLogLevel, logLevelFlagUsage)
	flag.StringVar(&cfg.FileStorage, fileStorageFlagName, defaultFileStoragePath, fileStorageFlagUsage)
	flag.StringVar(&cfg.DB, databaseFlagName, defaultDatabasePath, databaseFlagUsage)

	flag.Parse()

	if envRunHostAddr := os.Getenv("HOST_ADDRESS"); envRunHostAddr != "" {
		cfg.Address = envRunHostAddr
	}
	if envRunBaseURL := os.Getenv("BASE_URL"); envRunBaseURL != "" {
		cfg.BaseURL = envRunBaseURL
	}
	if envRunLogLevel := os.Getenv("LOG_LEVEL"); envRunLogLevel != "" {
		cfg.LogLevel = envRunLogLevel
	}
	if envRunFileStorage := os.Getenv("FILE_STORAGE_PATH"); envRunFileStorage != "" {
		cfg.FileStorage = envRunFileStorage
	}
	if envRunDatabase := os.Getenv("DATABASE_DSN"); envRunDatabase != "" {
		cfg.DB = envRunDatabase
	}

	// Убедиться, что baseURL заканчивается на /
	if !strings.HasSuffix(cfg.BaseURL, "/") {
		cfg.BaseURL += "/"
	}

	return cfg
}
