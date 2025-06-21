package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
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
	"log/slog"
	"net/http"
	"os"
	"strings"
)

var cfg config.Config

func handler(storage storage.Storage) http.HandlerFunc {
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

		shortURL := fmt.Sprintf("http://localhost:8080/"+"%s", shortID)

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
		shortURL := "http://localhost:8080/" + shortID

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
	logFormat := httplog.SchemaECS.Concise(true)

	loggerHttp := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", "example-app"),
		slog.String("version", "v1.0.0-a1fa420"),
		slog.String("env", "production"),
	)
	cfg := config.Init()

	if err := initLogger(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	logger.Log.Info("Running server on", zap.String("Address", cfg.Address))

	//strg := initStorage(cfg)
	store := storage.NewMapStorage()
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(loggerHttp, &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: httplog.SchemaECS,

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,

		// Optionally, filter out some request logs.
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},

		// Optionally, log selected request/response headers explicitly.
		LogRequestHeaders:  []string{"Origin"},
		LogResponseHeaders: []string{},

		// Optionally, enable logging of request/response body based on custom conditions.
		// Useful for debugging payload issues in development.
		LogRequestBody:  isDebugHeaderSet,
		LogResponseBody: isDebugHeaderSet,
	}))

	// Set request log attribute from within middleware.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			httplog.SetAttrs(ctx, slog.String("user", "user1"))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r.Get("/{id}", logger.RequestLogger(compress.GzipCompress(handlerGet(store))))
	r.Post("/", logger.RequestLogger(compress.GzipCompress(handler(store))))
	//r.Route("/", func(r chi.Router) {
	//	r.Route("/api/", func(r chi.Router) {
	//		r.Post("/shorten", logger.RequestLogger(PostShortenRequest(strg)))
	//	})
	//	r.Get("/ping", logger.RequestLogger(compress.GzipCompress(HandlerPing(strg))))
	//})

	chi.Walk(r, func(method, route string, h http.Handler, m ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-6s %s\n", method, route)
		return nil
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func initStorage(cfg config.Config) storage.Storage {
	var strg storage.Storage
	//var err error

	store := storage.NewMapStorage()

	//if cfg.DB != "" {
	//	strg, err = storage.NewPostgresStorage(context.Background(), cfg.DB)
	//	if err != nil {
	//		logger.Log.Error("Failed to initialize PostgreSQL storage: %v. Falling back to file storage", zap.Error(err))
	//	}
	//}

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

func isDebugHeaderSet(r *http.Request) bool {
	return r.Header.Get("Debug") == "reveal-body-logs"
}
