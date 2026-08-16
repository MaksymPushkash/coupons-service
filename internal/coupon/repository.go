package coupon

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, coupon Coupon) error
	GetByCode(ctx context.Context, code string) (Coupon, error)
	Apply(ctx context.Context, input ApplyInput, now time.Time) (ApplyResult, error)
	Deactivate(ctx context.Context, code string) error
}
