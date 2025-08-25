package handler

import (
	"context"
	"encoding/json"
	"github.com/ivanlp-p/ShortLinkService/cmd/config"
	"github.com/ivanlp-p/ShortLinkService/internal/logger"
	"github.com/ivanlp-p/ShortLinkService/internal/storage"
	"net/http"
)

func GetUrlsByUserID(storage storage.Storage, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		if userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		links, err := storage.GetUrlsByUserID(context.Background(), userID.(string))
		if err != nil {
			logger.Log.Error("Error when reading content")
		}

		if len(links) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		type resp struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}

		out := make([]resp, 0, len(links))
		for _, l := range links {
			out = append(out, resp{
				ShortURL:    conf.BaseURL + l.ShortURL,
				OriginalURL: l.OriginalURL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
