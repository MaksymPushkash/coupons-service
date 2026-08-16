package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MaksymPushkash/coupons-service/internal/coupon"
)

const maximumRequestBodySize = 1 << 20

type CouponService interface {
	Create(ctx context.Context, input coupon.CreateInput) (coupon.Coupon, error)
	Get(ctx context.Context, code string) (coupon.Coupon, error)
	Apply(ctx context.Context, input coupon.ApplyInput) (coupon.ApplyResult, error)
	Deactivate(ctx context.Context, code string) error
}

type CouponHandler struct {
	service CouponService
	logger  *slog.Logger
}

func NewCouponHandler(service CouponService, logger *slog.Logger) *CouponHandler {
	return &CouponHandler{
		service: service,
		logger:  logger,
	}
}

type createCouponRequest struct {
	Code           string              `json:"code"`
	Type           coupon.DiscountType `json:"type"`
	Value          int64               `json:"value"`
	Currency       *string             `json:"currency"`
	MaxRedemptions int                 `json:"max_redemptions"`
	ExpiresAt      time.Time           `json:"expires_at"`
}

type applyCouponRequest struct {
	InvoiceID string `json:"invoice_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type couponResponse struct {
	ID             string              `json:"id"`
	Code           string              `json:"code"`
	Type           coupon.DiscountType `json:"type"`
	Value          int64               `json:"value"`
	Currency       *string             `json:"currency,omitempty"`
	MaxRedemptions int                 `json:"max_redemptions"`
	RedeemedCount  int                 `json:"redeemed_count"`
	ExpiresAt      time.Time           `json:"expires_at"`
	Active         bool                `json:"active"`
	CreatedAt      time.Time           `json:"created_at"`
}

type applyCouponResponse struct {
	InvoiceID      string `json:"invoice_id"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	FinalAmount    int64  `json:"final_amount"`
	Currency       string `json:"currency"`
}

func (h *CouponHandler) Create(writer http.ResponseWriter, request *http.Request) {
	var body createCouponRequest

	if err := decodeJSON(writer, request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	created, err := h.service.Create(request.Context(), coupon.CreateInput{
		Code:           body.Code,
		Type:           body.Type,
		Value:          body.Value,
		Currency:       body.Currency,
		MaxRedemptions: body.MaxRedemptions,
		ExpiresAt:      body.ExpiresAt,
	})
	if err != nil {
		h.handleServiceError(writer, request, err)
		return
	}

	writer.Header().Set("Location", "/coupons/"+created.Code)
	writeJSON(writer, http.StatusCreated, toCouponResponse(created))
}

func (h *CouponHandler) Get(writer http.ResponseWriter, request *http.Request) {
	code := chi.URLParam(request, "code")

	value, err := h.service.Get(request.Context(), code)

	if err != nil {
		h.handleServiceError(writer, request, err)
		return
	}

	writeJSON(writer, http.StatusOK, toCouponResponse(value))
}

func (h *CouponHandler) Apply(writer http.ResponseWriter, request *http.Request) {
	var body applyCouponRequest

	if err := decodeJSON(writer, request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	result, err := h.service.Apply(request.Context(), coupon.ApplyInput{
		CouponCode: chi.URLParam(request, "code"),
		InvoiceID:  body.InvoiceID,
		Amount:     body.Amount,
		Currency:   body.Currency,
	})
	if err != nil {
		h.handleServiceError(writer, request, err)
		return
	}

	writeJSON(writer, http.StatusOK, applyCouponResponse{
		InvoiceID:      result.InvoiceID,
		OriginalAmount: result.OriginalAmount,
		DiscountAmount: result.DiscountAmount,
		FinalAmount:    result.FinalAmount,
		Currency:       result.Currency,
	})
}

func (h *CouponHandler) Deactivate(writer http.ResponseWriter, request *http.Request) {
	code := chi.URLParam(request, "code")

	if err := h.service.Deactivate(request.Context(), code); err != nil {
		h.handleServiceError(writer, request, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func toCouponResponse(value coupon.Coupon) couponResponse {
	return couponResponse{
		ID:             value.ID,
		Code:           value.Code,
		Type:           value.Type,
		Value:          value.Value,
		Currency:       value.Currency,
		MaxRedemptions: value.MaxRedemptions,
		RedeemedCount:  value.RedeemedCount,
		ExpiresAt:      value.ExpiresAt,
		Active:         value.Active,
		CreatedAt:      value.CreatedAt,
	}
}
