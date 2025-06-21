package main

import (
	"context"
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

var cfg config.Config

func handler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		shortID := utils.ShortenURL(originalURL)
		shortLink := models.ShortLink{
			UUID:        uuid.NewString(),
			ShortURL:    shortID,
			OriginalURL: originalURL,
		}

		storage.PutOriginalURL(context.Background(), shortLink)

		shortURL := fmt.Sprintf(cfg.BaseURL+"%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ID: ", chi.URLParam(r, "id"))

		id := chi.URLParam(r, "id")

		originalURL, err := storage.GetOriginalURL(context.Background(), id)

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

func PostShortenRequest(storage storage.Storage) http.HandlerFunc {
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

		storage.PutOriginalURL(context.Background(), shortLink)
		shortURL := cfg.BaseURL + shortID

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

func HandlerPing(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := storage.Ping(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	cfg := config.Init()

	if err := initLogger(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	logger.Log.Info("Running server on", zap.String("Address", cfg.Address))

	strg := initStorage(cfg)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", logger.RequestLogger(compress.GzipCompress(handler(strg))))
		r.Get("{id}", logger.RequestLogger(compress.GzipCompress(handlerGet(strg))))
		r.Route("/api/", func(r chi.Router) {
			r.Post("/shorten", logger.RequestLogger(PostShortenRequest(strg)))
		})
		r.Get("/ping", logger.RequestLogger(compress.GzipCompress(HandlerPing(strg))))
	})

	log.Fatal(http.ListenAndServe(cfg.Address, r))
}

func initStorage(cfg config.Config) storage.Storage {
	var strg storage.Storage
	var err error

	store := storage.NewMapStorage()

	if cfg.DB != "" {
		strg, err = storage.NewPostgresStorage(context.Background(), cfg.DB)
		if err != nil {
			logger.Log.Error("Failed to initialize PostgreSQL storage: %v. Falling back to file storage", zap.Error(err))
		}
	}

	if strg == nil && cfg.FileStorage != "" {
		strg = storage.NewFileStorage(cfg.FileStorage, store)
		err := strg.LoadFromFile()
		if err != nil {
			logger.Log.Error("Store not load")
		}
	}

	if strg == nil {
		strg = store
	}

	//conn, err := pgx.Connect(context.Background(), cfg.DB)
	//if err != nil {
	//	logger.Log.Error("Database not initialize")
	//}
	//defer conn.Close(context.Background())

	return strg
}

func initLogger(logLevel string) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	return nil
}
