package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/MaksymPushkash/coupons-service/internal/coupon"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *CouponHandler) handleServiceError(writer http.ResponseWriter, request *http.Request, err error) {
	status, code, message := mapServiceError(err)
	if status == http.StatusInternalServerError {
		h.logger.Error(
			"request failed",
			"request_id", middleware.GetReqID(request.Context()),
			"error", err,
			"method", request.Method,
			"path", request.URL.Path,
		)
	}
	writeError(writer, status, code, message)
}

func mapServiceError(err error) (int, string, string) {
	switch {
	case errors.Is(err, coupon.ErrCouponNotFound):
		return http.StatusNotFound, "coupon_not_found", "Coupon was not found"
	case errors.Is(err, coupon.ErrCouponCodeExists):
		return http.StatusConflict, "coupon_code_exists", "Coupon code already exists"
	case errors.Is(err, coupon.ErrAlreadyApplied):
		return http.StatusConflict, "coupon_already_applied", "Coupon was already applied to this invoice"
	case errors.Is(err, coupon.ErrCouponExpired):
		return http.StatusUnprocessableEntity, "coupon_expired", "Coupon has expired"
	case errors.Is(err, coupon.ErrCouponInactive):
		return http.StatusUnprocessableEntity, "coupon_inactive", "Coupon is inactive"
	case errors.Is(err, coupon.ErrRedemptionLimitReached):
		return http.StatusUnprocessableEntity, "redemption_limit_reached", "Coupon redemption limit has been reached"
	case errors.Is(err, coupon.ErrCurrencyMismatch):
		return http.StatusUnprocessableEntity, "currency_mismatch", "Coupon currency does not match invoice currency"
	case errors.Is(err, coupon.ErrInvalidCoupon), errors.Is(err, coupon.ErrInvalidApplyInput):
		return http.StatusBadRequest, "validation_error", err.Error()
	default:
		return http.StatusInternalServerError, "internal_error", "Internal server error"
	}
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, errorResponse{
		Error: errorBody{Code: code, Message: message},
	})
}
