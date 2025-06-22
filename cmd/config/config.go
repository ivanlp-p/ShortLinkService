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
	defaultFileStoragePath = "/tmp/short-url-db.json"
	fileStorageFlagUsage

	databaseFlagName    = "d"
	defaultDatabasePath = "postgres://postgres:GfdGjc964@localhost:5432/videos" //postgres://postgres:GfdGjc964@localhost:5432/videos
	databaseFlagUsage   = "Database path"
)

type Config struct {
	Address     string
	BaseURL     string
	LogLevel    string
	FileStorage string
	DB          string
}

var config *Config

func Init() *Config {
	if config != nil {
		return config
	}

	parseFlags()
	return config
}

func parseFlags() *Config {

	if config == nil {
		config = &Config{}
	}
	flag.StringVar(&config.Address, hostFlagName, defaultPort, hostFlagUsage)
	flag.StringVar(&config.BaseURL, baseURLFlagName, defaultEndpoint, baseURLFlagUsage)
	flag.StringVar(&config.LogLevel, logLevelFlagName, defaultLogLevel, logLevelFlagUsage)
	flag.StringVar(&config.FileStorage, fileStorageFlagName, defaultFileStoragePath, fileStorageFlagUsage)
	flag.StringVar(&config.DB, databaseFlagName, defaultDatabasePath, databaseFlagUsage)

	flag.Parse()

	if envRunHostAddr := os.Getenv("HOST_ADDRESS"); envRunHostAddr != "" {
		config.Address = envRunHostAddr
	}
	if envRunBaseURL := os.Getenv("BASE_URL"); envRunBaseURL != "" {
		config.BaseURL = envRunBaseURL
	}
	if envRunLogLevel := os.Getenv("LOG_LEVEL"); envRunLogLevel != "" {
		config.LogLevel = envRunLogLevel
	}
	if envRunFileStorage := os.Getenv("FILE_STORAGE_PATH"); envRunFileStorage != "" {
		config.FileStorage = envRunFileStorage
	}
	if envRunDatabase := os.Getenv("DATABASE_DSN"); envRunDatabase != "" {
		config.DB = envRunDatabase
	}

	// Убедиться, что baseURL заканчивается на /
	if !strings.HasSuffix(config.BaseURL, "/") {
		config.BaseURL += "/"
	}

	return config
}
