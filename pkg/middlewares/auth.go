package middlewares

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/auth"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/logger"
	"net/http"
)

func Auth(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHandler := auth.New(cfg)

			cookie, err := r.Cookie("Authorization")
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				logger.Logger.Error(err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := authHandler.ParseToken(cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if claims.UserID == uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
