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
)

var (
	Address     string
	BaseURL     string
	LogLevel    string
	FileStorage string
)

func Init() {
	flag.StringVar(&Address, hostFlagName, defaultPort, hostFlagUsage)
	flag.StringVar(&BaseURL, baseURLFlagName, defaultEndpoint, baseURLFlagUsage)
	flag.StringVar(&LogLevel, logLevelFlagName, defaultLogLevel, logLevelFlagUsage)
	flag.StringVar(&FileStorage, fileStorageFlagName, defaultFileStoragePath, fileStorageFlagUsage)

	flag.Parse()

	if envRunHostAddr := os.Getenv("HOST_ADDRESS"); envRunHostAddr != "" {
		Address = envRunHostAddr
	}
	if envRunBaseURL := os.Getenv("BASE_URL"); envRunBaseURL != "" {
		BaseURL = envRunBaseURL
	}
	if envRunLogLevel := os.Getenv("LOG_LEVEL"); envRunLogLevel != "" {
		LogLevel = envRunLogLevel
	}
	if envRunFileStorage := os.Getenv("FILE_STORAGE_PATH"); envRunFileStorage != "" {
		FileStorage = envRunFileStorage
	}

	// Убедиться, что baseURL заканчивается на /
	if !strings.HasSuffix(BaseURL, "/") {
		BaseURL += "/"
	}

}
