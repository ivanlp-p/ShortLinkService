package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ivanlp-p/ShortLinkService/cmd/config"
	"github.com/ivanlp-p/ShortLinkService/internal/compress"
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

func handler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		shortID := utils.ShortenURL(originalURL)

		storage.Set(shortID, originalURL)

		shortURL := fmt.Sprintf(config.BaseURL+"%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(fileStorage *storage.FileStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ID: ", chi.URLParam(r, "id"))

		id := chi.URLParam(r, "id")

		originalURL, err := fileStorage.GetOriginalURL(id)

		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		_, err = io.ReadAll(r.Body)
		if err != nil || originalURL == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		fmt.Println("originalURL: ", originalURL)

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func PostShortenRequest(fileStorage *storage.FileStorage) http.HandlerFunc {
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
		shortLink := models.ShortLink{UUID: uuid.NewString(),
			ShortURL:    shortID,
			OriginalURL: url,
		}

		fileStorage.SaveShortLink(shortLink)
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

	if err := run(); err != nil {
		log.Fatal(err)
	}

	store := storage.NewMapStorage()
	fileStorage := storage.NewFileStorage(config.FileStorage, store)

	err := fileStorage.LoadFromFile()
	if err != nil {
		logger.Log.Error("Store not load")
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", logger.RequestLogger(compress.GzipCompress(handler(store))))
		r.Get("/{id}", logger.RequestLogger(compress.GzipCompress(handlerGet(fileStorage))))
		r.Route("/api/", func(r chi.Router) {
			r.Post("/shorten", logger.RequestLogger(PostShortenRequest(fileStorage)))
		})
	})

	log.Fatal(http.ListenAndServe(config.Address, r))
}

func run() error {
	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}
	logger.Log.Info("Running server on", zap.String("Address", config.Address))
	return nil
}
