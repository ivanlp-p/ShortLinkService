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

	baseUrlFlagName  = "b"
	defaultEndpoint  = "http://localhost:8080/"
	baseUrlFlagUsage = "Base URL for shortened links"
)

var (
	Address string
	BaseURL string
)

func Init() {
	flag.StringVar(&Address, hostFlagName, defaultPort, hostFlagUsage)
	flag.StringVar(&BaseURL, baseUrlFlagName, defaultEndpoint, baseUrlFlagUsage)

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
