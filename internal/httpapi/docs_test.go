package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPISpec(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	response := httptest.NewRecorder()
	testRouter(&couponServiceStub{}, healthCheckerStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/yaml" {
		t.Fatalf("Content-Type = %q, want application/yaml", contentType)
	}
	if !strings.Contains(response.Body.String(), "openapi: 3.0.3") {
		t.Fatal("response does not contain the OpenAPI specification")
	}
}

func TestSwaggerUI(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	response := httptest.NewRecorder()
	testRouter(&couponServiceStub{}, healthCheckerStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "Swagger UI") {
		t.Fatal("response does not contain Swagger UI")
	}
}

func TestSwaggerRedirect(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/docs", nil)
	response := httptest.NewRecorder()
	testRouter(&couponServiceStub{}, healthCheckerStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusPermanentRedirect)
	}
	if location := response.Header().Get("Location"); location != "/docs/" {
		t.Fatalf("Location = %q, want /docs/", location)
	}
}
