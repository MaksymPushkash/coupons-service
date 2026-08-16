package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			started := time.Now()
			wrappedWriter := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)

			next.ServeHTTP(wrappedWriter, request)

			logger.Info(
				"request completed",
				"request_id", middleware.GetReqID(request.Context()),
				"method", request.Method,
				"path", request.URL.Path,
				"status", wrappedWriter.Status(),
				"bytes", wrappedWriter.BytesWritten(),
				"duration_ms", time.Since(started).Milliseconds(),
				"remote_ip", request.RemoteAddr,
			)
		})
	}
}
