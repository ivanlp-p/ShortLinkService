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

func handler(storage storage.Storage, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("This is handler")
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

		shortURL := fmt.Sprintf(conf.BaseURL+"%s", shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}

func handlerGet(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("This is handlerGet")
		fmt.Println("ID: ", chi.URLParam(r, "id"))
		logger.Log.Info("ID: ", zap.String("id - ", chi.URLParam(r, "id")))
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

func PostShortenRequest(storage storage.Storage, conf *config.Config) http.HandlerFunc {
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
		shortURL := conf.BaseURL + shortID

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
	conf := config.Init()
	if err := initLogger(config.Config{}.LogLevel); err != nil {
		log.Fatal(err)
	}

	logger.Log.Info("Running server on", zap.String("Address", conf.Address))

	strg := initStorage(conf)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", logger.RequestLogger(compress.GzipCompress(handlerGet(strg))))
		r.Post("/", logger.RequestLogger(compress.GzipCompress(handler(strg, conf))))
		r.Route("/api/", func(r chi.Router) {
			r.Post("/shorten", logger.RequestLogger(PostShortenRequest(strg, conf)))
		})
		r.Get("/ping", logger.RequestLogger(compress.GzipCompress(HandlerPing(strg))))
	})

	chi.Walk(r, func(method, route string, h http.Handler, m ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-6s %s\n", method, route)
		return nil
	})

	log.Fatal(http.ListenAndServe(conf.Address, r))
}

func initStorage(conf *config.Config) storage.Storage {
	var strg storage.Storage
	var err error

	store := storage.NewMapStorage()

	if conf.DB != "" {
		strg, err = storage.NewPostgresStorage(context.Background(), conf.DB)
		if err != nil {
			logger.Log.Error("Failed to initialize PostgreSQL storage: %v. Falling back to file storage", zap.Error(err))
		}
	}

	if strg == nil && conf.FileStorage != "" {
		strg = storage.NewFileStorage(conf.FileStorage, store)
		err := strg.LoadFromFile()
		if err != nil {
			logger.Log.Error("Store not load")
		}
	}

	if strg == nil {
		strg = store
	}

	return strg
}

func initLogger(logLevel string) error {
	if err := logger.Initialize(logLevel); err != nil {
		return err
	}

	return nil
}
