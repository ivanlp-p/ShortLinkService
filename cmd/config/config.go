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
)

var (
	Address string
	BaseURL string
)

func Init() {
	flag.StringVar(&Address, hostFlagName, defaultPort, hostFlagUsage)
	flag.StringVar(&BaseURL, baseURLFlagName, defaultEndpoint, baseURLFlagUsage)

	flag.Parse()

	if envRunHostAddr := os.Getenv("HOST_ADDRESS"); envRunHostAddr != "" {
		Address = envRunHostAddr
	}
	if envRunBaseURL := os.Getenv("BASE_URL"); envRunBaseURL != "" {
		BaseURL = envRunBaseURL
	}

	// Убедиться, что baseURL заканчивается на /
	if !strings.HasSuffix(BaseURL, "/") {
		BaseURL += "/"
	}

}
