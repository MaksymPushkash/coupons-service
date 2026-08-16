package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

func NewRouter(handler *CouponHandler, checker HealthChecker, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requestLogger(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(15 * time.Second))

	router.Get("/health/live", liveHandler)
	router.Get("/health/ready", readyHandler(checker))
	mountDocs(router)

	router.Post("/coupons", handler.Create)
	router.Get("/coupons/{code}", handler.Get)
	router.Post("/coupons/{code}/apply", handler.Apply)
	router.Post("/coupons/{code}/deactivate", handler.Deactivate)

	return router
}

func liveHandler(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func readyHandler(checker HealthChecker) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()

		if err := checker.Ping(ctx); err != nil {
			writeError(writer, http.StatusServiceUnavailable, "not_ready", "Service is not ready")
			return
		}

		writeJSON(writer, http.StatusOK, map[string]string{"status": "ready"})
	}
}
