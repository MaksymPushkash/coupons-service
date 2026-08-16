package coupon

import "errors"

var (
	ErrCouponNotFound         = errors.New("coupon not found")
	ErrCouponCodeExists       = errors.New("coupon code already exists")
	ErrCouponExpired          = errors.New("coupon expired")
	ErrCouponInactive         = errors.New("coupon inactive")
	ErrRedemptionLimitReached = errors.New("redemption limit reached")
	ErrAlreadyApplied         = errors.New("coupon already applied to invoice")
	ErrCurrencyMismatch       = errors.New("currency mismatch")
	ErrInvalidCoupon          = errors.New("invalid coupon")
	ErrInvalidApplyInput      = errors.New("invalid apply input")
)
