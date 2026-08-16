package coupon

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

const maxInvoiceIDLength = 128

var (
	couponCodePattern = regexp.MustCompile(`^[A-Z0-9_-]{3,64}$`)
	currencyPattern   = regexp.MustCompile(`^[A-Z]{3}$`)
)

func normalizeCreateInput(input CreateInput) CreateInput {
	input.Code = normalizeCode(input.Code)
	input.Currency = normalizeCurrencyPointer(input.Currency)
	input.ExpiresAt = input.ExpiresAt.UTC()
	return input
}

func normalizeApplyInput(input ApplyInput) ApplyInput {
	input.CouponCode = normalizeCode(input.CouponCode)
	input.InvoiceID = strings.TrimSpace(input.InvoiceID)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	return input
}

func validateApplyInput(input ApplyInput) error {
	if err := validateCode(input.CouponCode, ErrInvalidApplyInput); err != nil {
		return err
	}
	if input.InvoiceID == "" || len(input.InvoiceID) > maxInvoiceIDLength {
		return fmt.Errorf(
			"%w: invoice ID must contain 1-%d characters",
			ErrInvalidApplyInput,
			maxInvoiceIDLength,
		)
	}
	if !currencyPattern.MatchString(input.Currency) {
		return fmt.Errorf("%w: currency must contain three uppercase letters", ErrInvalidApplyInput)
	}
	if input.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidApplyInput)
	}
	return nil
}

func validateCreateInput(input CreateInput, now time.Time) error {
	if err := validateCode(input.Code, ErrInvalidCoupon); err != nil {
		return err
	}
	if input.Value <= 0 {
		return fmt.Errorf("%w: value must be positive", ErrInvalidCoupon)
	}
	if input.MaxRedemptions <= 0 || input.MaxRedemptions > math.MaxInt32 {
		return fmt.Errorf(
			"%w: max redemptions must be between 1 and %d",
			ErrInvalidCoupon,
			math.MaxInt32,
		)
	}
	if input.ExpiresAt.IsZero() || !input.ExpiresAt.After(now) {
		return fmt.Errorf("%w: expiration must be in the future", ErrInvalidCoupon)
	}

	switch input.Type {
	case DiscountPercentage:
		if input.Value > 100 {
			return fmt.Errorf("%w: percentage must not exceed 100", ErrInvalidCoupon)
		}
		if input.Currency != nil {
			return fmt.Errorf("%w: percentage coupon must not have currency", ErrInvalidCoupon)
		}
	case DiscountFixed:
		if input.Currency == nil {
			return fmt.Errorf("%w: fixed coupon requires currency", ErrInvalidCoupon)
		}
		if !currencyPattern.MatchString(*input.Currency) {
			return fmt.Errorf("%w: currency must contain three uppercase letters", ErrInvalidCoupon)
		}
	default:
		return fmt.Errorf("%w: unsupported discount type", ErrInvalidCoupon)
	}

	return nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func validateCode(code string, target error) error {
	if couponCodePattern.MatchString(code) {
		return nil
	}
	return fmt.Errorf(
		"%w: code must contain 3-64 uppercase letters, digits, underscores, or hyphens",
		target,
	)
}

func normalizeCurrencyPointer(currency *string) *string {
	if currency == nil {
		return nil
	}
	value := strings.ToUpper(strings.TrimSpace(*currency))
	if value == "" {
		return nil
	}
	return &value
}
