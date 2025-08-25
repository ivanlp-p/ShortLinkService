package middleware

import (
	"context"
	"github.com/ivanlp-p/ShortLinkService/internal/cookie"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userID, ok := cookie.ValidateAuthCookie(r); ok {
			// если кука валидна, добавить userID в контекст
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// если куки нет — сгенерировать и установить
		userID := cookie.GenerateUserID()
		http.SetCookie(w, cookie.MakeSignedCookie(userID))

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(userIDKey).(string)
	return userID, ok
}
