package middleware

import (
	"context"
	"fmt"

	"net/http"
	"redditclone/pkg/helpers"
	"time"

	"go.uber.org/zap"
)

type logrKey string

const (
	loggerKey logrKey = "logger"
)

func LogMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)

			GetLogger(r.Context()).Info("Request",
				zap.String("url", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Duration("time", time.Since(start)),
			)
		})
	}
}

func SetupLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = helpers.RandBytesHex(16)
				r.Header.Set("X-Request-ID", requestID)
				r.Header.Set("trace-id", requestID)
				w.Header().Set("trace-id", requestID)
				w.Header().Set("X-Request-ID", requestID)
			}
			ctxlogger := logger.With(
				zap.String("logger", "ctxlog"),
				zap.String("trace-id", requestID),
			).WithOptions(
				zap.AddCaller(),
				zap.AddStacktrace(zap.ErrorLevel),
			)

			ctx := context.WithValue(r.Context(), loggerKey, ctxlogger)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		fmt.Println("No logger in req context")
		return zap.L()
	}
	return logger
}
