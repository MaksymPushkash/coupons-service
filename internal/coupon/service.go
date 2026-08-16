package coupon

import "context"


type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Apply(ctx context.Context, input ApplyInput) (ApplyResult, error) {
	coupon, err := s.repository.Apply(ctx, input)

	if err != nil {
		return ApplyResult{}, err
	}

	discount := calculateDiscount(coupon, input.Amount)

	return ApplyResult{
		InvoiceID: input.InvoiceID,
		OriginalAmount: input.Amount,
		DiscountAmount: discount,
		FinalAmount: input.Amount - discount,
		Currency: input.Currency,
	}, nil
}

