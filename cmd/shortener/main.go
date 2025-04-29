package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	urlStore = make(map[string]string)
	mu       sync.Mutex

	address string
	baseURL string
)

func init() {
	flag.StringVar(&address, "a", ":8080", "Address to launch the HTTP server")
	flag.StringVar(&baseURL, "b", "http://localhost:8080/", "Base URL for shortened links")

}

func shortenURL(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	bs := h.Sum(nil)
	encoded := base64.URLEncoding.EncodeToString(bs)
	// Используем первые 8 символов для короткого URL
	short := encoded[:8]
	return short
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		shortID := shortenURL(originalURL)

		mu.Lock()
		urlStore[shortID] = originalURL
		mu.Unlock()

		shortURL := fmt.Sprintf(baseURL+"%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(urlStore map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ID: ", chi.URLParam(r, "id"))
		if r.Method != http.MethodGet {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		id := chi.URLParam(r, "id")
		originalUrl := urlStore[id]

		_, err := io.ReadAll(r.Body)
		if err != nil || originalUrl == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		fmt.Println("originalUrl: ", originalUrl)

		w.Header().Set("Location", originalUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func main() {
	flag.Parse()

	if envRunHostAddr := os.Getenv("HOST_ADDRESS"); envRunHostAddr != "" {
		address = envRunHostAddr
	}
	if envRunBaseUrl := os.Getenv("BASE_URL"); envRunBaseUrl != "" {
		baseURL = envRunBaseUrl
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", handler())
		r.Get("/{id}", handlerGet(urlStore))
	})

	log.Fatal(http.ListenAndServe(address, r))
}
