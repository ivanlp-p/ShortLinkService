package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	urlStore = make(map[string]string)
	mu       sync.Mutex
)

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

		shortURL := fmt.Sprintf("%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(urlStore map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ID: ", r.PathValue("id"))
		if r.Method != http.MethodGet {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		id := r.PathValue("id")
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler())
	mux.HandleFunc("/{id}", handlerGet(urlStore))

	fmt.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
