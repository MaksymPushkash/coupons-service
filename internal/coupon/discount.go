package coupon

import (
	"fmt"
	"time"
)

func CalculateDiscount(coupon Coupon, input ApplyInput, now time.Time) (ApplyResult, error) {
	input = normalizeApplyInput(input)
	if err := validateApplyInput(input); err != nil {
		return ApplyResult{}, err
	}

	if !coupon.Active {
		return ApplyResult{}, ErrCouponInactive
	}

	if !coupon.ExpiresAt.After(now) {
		return ApplyResult{}, ErrCouponExpired
	}

	if coupon.MaxRedemptions <= 0 {
		return ApplyResult{}, fmt.Errorf("%w: invalid redemption limit", ErrInvalidCoupon)
	}

	if coupon.RedeemedCount >= coupon.MaxRedemptions {
		return ApplyResult{}, ErrRedemptionLimitReached
	}

	var discount int64

	switch coupon.Type {
	case DiscountPercentage:
		if coupon.Value < 1 || coupon.Value > 100 {
			return ApplyResult{}, fmt.Errorf("%w: percentage must be between 1 and 100", ErrInvalidCoupon)
		}

		discount = (input.Amount/100)*coupon.Value + (input.Amount%100)*coupon.Value/100

	case DiscountFixed:
		if coupon.Value <= 0 {
			return ApplyResult{}, fmt.Errorf("%w: fixed discount must be positive", ErrInvalidCoupon)
		}

		if coupon.Currency == nil {
			return ApplyResult{}, fmt.Errorf("%w: fixed coupon requires currency", ErrInvalidCoupon)
		}

		if *coupon.Currency != input.Currency {
			return ApplyResult{}, ErrCurrencyMismatch
		}

		discount = coupon.Value

		if discount > input.Amount {
			discount = input.Amount
		}

	default:
		return ApplyResult{}, fmt.Errorf("%w: unsupported discount type", ErrInvalidCoupon)
	}

	return ApplyResult{
		InvoiceID:      input.InvoiceID,
		OriginalAmount: input.Amount,
		DiscountAmount: discount,
		FinalAmount:    input.Amount - discount,
		Currency:       input.Currency,
	}, nil
}
