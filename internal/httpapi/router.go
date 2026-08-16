package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)


func NewRouter(handler *CouponHandler) http.Handler {
	router := chi.NewRouter()

	router.Get("/health", healthHandler)

	router.Route("/coupons", func(router chi.Router) {
		router.Post("/", handler.Create)
		router.Get("/{code}", handler.Get)
		router.Post("/{code}/apply", handler.Apply)
		router.Post("/{code}/deactivate", handler.Deactivate)
	})
	return router
}

