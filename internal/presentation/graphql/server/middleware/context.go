package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "logger"

func WithLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loggerKey, log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggerFromCtx(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return log
	}
	return zap.NewNop()
}
