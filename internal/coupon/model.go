package coupon

import "time"

type DiscountType string

const (
	DiscountPercentage DiscountType = "percentage"
	DiscountFixed      DiscountType = "fixed"
)

type Coupon struct {
	ID             string
	Code           string
	Type           DiscountType
	Value          int64
	Currency       *string
	MaxRedemptions int
	RedeemedCount  int
	ExpiresAt      time.Time
	Active         bool
	CreatedAt      time.Time
}

type CreateInput struct {
	Code           string
	Type           DiscountType
	Value          int64
	Currency       *string
	MaxRedemptions int
	ExpiresAt      time.Time
}

type ApplyInput struct {
	CouponCode string
	InvoiceID  string
	Amount     int64
	Currency   string
}

type ApplyResult struct {
	InvoiceID      string
	OriginalAmount int64
	DiscountAmount int64
	FinalAmount    int64
	Currency       string
}
