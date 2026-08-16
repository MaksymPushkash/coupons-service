package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MaksymPushkash/coupons-service/internal/coupon"
)

type couponServiceStub struct {
	createResult  coupon.Coupon
	createErr     error
	getResult     coupon.Coupon
	getErr        error
	applyResult   coupon.ApplyResult
	applyErr      error
	deactivateErr error
}

func (s *couponServiceStub) Create(_ context.Context, _ coupon.CreateInput) (coupon.Coupon, error) {
	return s.createResult, s.createErr
}

func (s *couponServiceStub) Get(_ context.Context, _ string) (coupon.Coupon, error) {
	return s.getResult, s.getErr
}

func (s *couponServiceStub) Apply(_ context.Context, _ coupon.ApplyInput) (coupon.ApplyResult, error) {
	return s.applyResult, s.applyErr
}

func (s *couponServiceStub) Deactivate(_ context.Context, _ string) error {
	return s.deactivateErr
}

type healthCheckerStub struct {
	err error
}

func (s healthCheckerStub) Ping(_ context.Context) error {
	return s.err
}

func testRouter(service CouponService, checker HealthChecker) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(NewCouponHandler(service, logger), checker, logger)
}

func TestCreateCoupon(t *testing.T) {
	expiresAt := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	service := &couponServiceStub{
		createResult: coupon.Coupon{
			ID:             "coupon-id",
			Code:           "WELCOME20",
			Type:           coupon.DiscountPercentage,
			Value:          20,
			MaxRedemptions: 100,
			ExpiresAt:      expiresAt,
			Active:         true,
			CreatedAt:      time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC),
		},
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/coupons",
		strings.NewReader(`{
			"code":"WELCOME20",
			"type":"percentage",
			"value":20,
			"max_redemptions":100,
			"expires_at":"2027-01-01T00:00:00Z"
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	testRouter(service, healthCheckerStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	if location := response.Header().Get("Location"); location != "/coupons/WELCOME20" {
		t.Fatalf("Location = %q, want /coupons/WELCOME20", location)
	}
}

func TestCreateCouponRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{"missing content type", "", `{}`},
		{"unknown field", "application/json", `{"unknown":true}`},
		{"multiple objects", "application/json", `{} {}`},
		{"empty body", "application/json", ``},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/coupons", strings.NewReader(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			response := httptest.NewRecorder()

			testRouter(&couponServiceStub{}, healthCheckerStub{}).ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestApplyCouponErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", coupon.ErrCouponNotFound, http.StatusNotFound},
		{"already applied", coupon.ErrAlreadyApplied, http.StatusConflict},
		{"expired", coupon.ErrCouponExpired, http.StatusUnprocessableEntity},
		{"invalid input", coupon.ErrInvalidApplyInput, http.StatusBadRequest},
		{"internal error", errors.New("database unavailable"), http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &couponServiceStub{applyErr: test.err}
			request := httptest.NewRequest(
				http.MethodPost,
				"/coupons/WELCOME20/apply",
				strings.NewReader(`{"invoice_id":"inv-1","amount":10000,"currency":"USD"}`),
			)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			testRouter(service, healthCheckerStub{}).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.wantStatus, response.Body.String())
			}

			var body errorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body.Error.Code == "" {
				t.Fatal("error response has an empty code")
			}
		})
	}
}

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		checker    HealthChecker
		wantStatus int
	}{
		{"live", "/health/live", healthCheckerStub{err: errors.New("ignored")}, http.StatusOK},
		{"ready", "/health/ready", healthCheckerStub{}, http.StatusOK},
		{
			"not ready",
			"/health/ready",
			healthCheckerStub{err: errors.New("database unavailable")},
			http.StatusServiceUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			testRouter(&couponServiceStub{}, test.checker).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
