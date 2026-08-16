package coupon

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Coupon, error) {
	now := s.now().UTC()
	input = normalizeCreateInput(input)

	if err := validateCreateInput(input, now); err != nil {
		return Coupon{}, err
	}

	coupon := Coupon{
		ID:             uuid.NewString(),
		Code:           input.Code,
		Type:           input.Type,
		Value:          input.Value,
		Currency:       input.Currency,
		MaxRedemptions: input.MaxRedemptions,
		RedeemedCount:  0,
		ExpiresAt:      input.ExpiresAt,
		Active:         true,
		CreatedAt:      now,
	}

	if err := s.repository.Create(ctx, coupon); err != nil {
		return Coupon{}, err
	}

	return coupon, nil
}

func (s *Service) Get(ctx context.Context, code string) (Coupon, error) {
	code = normalizeCode(code)

	if err := validateCode(code, ErrInvalidCoupon); err != nil {
		return Coupon{}, err
	}

	return s.repository.GetByCode(ctx, code)
}

func (s *Service) Apply(ctx context.Context, input ApplyInput) (ApplyResult, error) {
	input = normalizeApplyInput(input)
	if err := validateApplyInput(input); err != nil {
		return ApplyResult{}, err
	}

	return s.repository.Apply(ctx, input, s.now().UTC())
}

func (s *Service) Deactivate(ctx context.Context, code string) error {
	code = normalizeCode(code)

	if err := validateCode(code, ErrInvalidCoupon); err != nil {
		return err
	}

	return s.repository.Deactivate(ctx, code)
}
