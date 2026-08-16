package coupon

import (
	"errors"
	"testing"
	"time"
)

func TestCalculateDiscount(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	usd := "USD"

	baseCoupon := Coupon{
		Type:           DiscountPercentage,
		Value:          20,
		MaxRedemptions: 10,
		RedeemedCount:  0,
		ExpiresAt:      now.Add(time.Hour),
		Active:         true,
	}

	baseInput := ApplyInput{
		CouponCode: "WELCOME20",
		InvoiceID:  "inv-1",
		Amount:     10000,
		Currency:   "USD",
	}

	tests := []struct {
		name         string
		coupon       Coupon
		input        ApplyInput
		wantDiscount int64
		wantFinal    int64
		wantError    error
	}{
		{
			name:         "percentage",
			coupon:       baseCoupon,
			input:        baseInput,
			wantDiscount: 2000,
			wantFinal:    8000,
		},
		{
			name: "fixed",
			coupon: Coupon{
				Type:           DiscountFixed,
				Value:          1500,
				Currency:       &usd,
				MaxRedemptions: 10,
				ExpiresAt:      now.Add(time.Hour),
				Active:         true,
			},
			input:        baseInput,
			wantDiscount: 1500,
			wantFinal:    8500,
		},
		{
			name: "fixed discount cannot make amount negative",
			coupon: Coupon{
				Type:           DiscountFixed,
				Value:          20000,
				Currency:       &usd,
				MaxRedemptions: 10,
				ExpiresAt:      now.Add(time.Hour),
				Active:         true,
			},
			input:        baseInput,
			wantDiscount: 10000,
			wantFinal:    0,
		},
		{
			name: "expired",
			coupon: Coupon{
				Type:           DiscountPercentage,
				Value:          20,
				MaxRedemptions: 10,
				ExpiresAt:      now.Add(-time.Second),
				Active:         true,
			},
			input:     baseInput,
			wantError: ErrCouponExpired,
		},
		{
			name: "inactive",
			coupon: Coupon{
				Type:           DiscountPercentage,
				Value:          20,
				MaxRedemptions: 10,
				ExpiresAt:      now.Add(time.Hour),
				Active:         false,
			},
			input:     baseInput,
			wantError: ErrCouponInactive,
		},
		{
			name: "limit reached",
			coupon: Coupon{
				Type:           DiscountPercentage,
				Value:          20,
				MaxRedemptions: 10,
				RedeemedCount:  10,
				ExpiresAt:      now.Add(time.Hour),
				Active:         true,
			},
			input:     baseInput,
			wantError: ErrRedemptionLimitReached,
		},
		{
			name: "currency mismatch",
			coupon: Coupon{
				Type:           DiscountFixed,
				Value:          1000,
				Currency:       &usd,
				MaxRedemptions: 10,
				ExpiresAt:      now.Add(time.Hour),
				Active:         true,
			},
			input: ApplyInput{
				CouponCode: "FIXED",
				InvoiceID:  "inv-1",
				Amount:     10000,
				Currency:   "EUR",
			},
			wantError: ErrCurrencyMismatch,
		},
		{
			name:   "invalid amount",
			coupon: baseCoupon,
			input: ApplyInput{
				CouponCode: "WELCOME20",
				InvoiceID:  "inv-1",
				Amount:     0,
				Currency:   "USD",
			},
			wantError: ErrInvalidApplyInput,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := CalculateDiscount(
				test.coupon,
				test.input,
				now,
			)

			if test.wantError != nil {
				if !errors.Is(err, test.wantError) {
					t.Fatalf(
						"got error %v, want %v",
						err,
						test.wantError,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.DiscountAmount != test.wantDiscount {
				t.Fatalf(
					"discount = %d, want %d",
					result.DiscountAmount,
					test.wantDiscount,
				)
			}

			if result.FinalAmount != test.wantFinal {
				t.Fatalf(
					"final amount = %d, want %d",
					result.FinalAmount,
					test.wantFinal,
				)
			}
		})
	}
}
