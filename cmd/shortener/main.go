package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ivanlp-p/ShortLinkService/cmd/config"
	"github.com/ivanlp-p/ShortLinkService/internal/logger"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"github.com/ivanlp-p/ShortLinkService/internal/storage"
	"github.com/ivanlp-p/ShortLinkService/internal/utils"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"strings"
)

func handler(store *storage.MapStorage) http.HandlerFunc {
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
		shortID := utils.ShortenURL(originalURL)

		store.Set(shortID, originalURL)

		shortURL := fmt.Sprintf(config.BaseURL+"%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(store *storage.MapStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ID: ", chi.URLParam(r, "id"))
		if r.Method != http.MethodGet {
			http.Error(w, "Bad Request", http.StatusMethodNotAllowed)
			return
		}

		id := chi.URLParam(r, "id")

		originalURL, ok := store.Get(id)

		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		_, err := io.ReadAll(r.Body)
		if err != nil || originalURL == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		fmt.Println("originalURL: ", originalURL)

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostShortenRequest(store *storage.MapStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var originURL models.OriginalURL

		if r.Method != http.MethodPost {
			http.Error(w, "Bad Request", http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&originURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := originURL.URL
		shortID := utils.ShortenURL(url)
		store.Set(shortID, url)
		shortURL := config.BaseURL + shortID

		resp := models.ShortURL{
			Result: shortURL,
		}

		response, err := json.MarshalIndent(resp, "", "   ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func main() {
	config.Init()

	run()

	store := storage.NewMapStorage()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", logger.RequestLogger(handler(store)))
		r.Get("/{id}", logger.RequestLogger(handlerGet(store)))
		r.Route("/api/", func(r chi.Router) {
			r.Post("/shorten", logger.RequestLogger(PostShortenRequest(store)))
		})
	})

	log.Fatal(http.ListenAndServe(config.Address, r))
}

func run() {
	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Log.Error(err.Error())
	}

	logger.Log.Info("Running server on", zap.String("Address", config.Address))
}
