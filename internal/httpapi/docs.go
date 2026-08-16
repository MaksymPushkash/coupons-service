package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/swgui/v5emb"

	apidocs "github.com/MaksymPushkash/coupons-service/api"
)

func mountDocs(router chi.Router) {
	router.Get("/openapi.yaml", serveOpenAPISpec)
	router.Get("/docs", redirectToDocs)
	router.Handle("/docs/*", v5emb.New(
		"Coupons Service API",
		"/openapi.yaml",
		"/docs/",
	))
}

func serveOpenAPISpec(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/yaml")
	_, _ = writer.Write(apidocs.Spec)
}

func redirectToDocs(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/docs/", http.StatusPermanentRedirect)
}
