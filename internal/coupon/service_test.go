package coupon

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct {
	created     Coupon
	applied     ApplyInput
	createErr   error
	applyErr    error
	applyResult ApplyResult
}

func (r *repositoryStub) Create(_ context.Context, value Coupon) error {
	r.created = value
	return r.createErr
}

func (r *repositoryStub) GetByCode(_ context.Context, code string) (Coupon, error) {
	return Coupon{Code: code}, nil
}

func (r *repositoryStub) Apply(_ context.Context, input ApplyInput, _ time.Time) (ApplyResult, error) {
	r.applied = input
	return r.applyResult, r.applyErr
}

func (r *repositoryStub) Deactivate(_ context.Context, _ string) error {
	return nil
}

func TestServiceCreateNormalizesInput(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	currency := " usd "
	repository := &repositoryStub{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	created, err := service.Create(context.Background(), CreateInput{
		Code:           " fixed-10 ",
		Type:           DiscountFixed,
		Value:          1000,
		Currency:       &currency,
		MaxRedemptions: 10,
		ExpiresAt:      now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Code != "FIXED-10" {
		t.Fatalf("Code = %q, want FIXED-10", created.Code)
	}

	if created.Currency == nil || *created.Currency != "USD" {
		t.Fatalf("Currency = %v, want USD", created.Currency)
	}

	if repository.created.ID == "" {
		t.Fatal("repository received an empty coupon ID")
	}
}

func TestServiceCreateValidation(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	usd := "USD"
	validInput := CreateInput{
		Code:           "WELCOME20",
		Type:           DiscountPercentage,
		Value:          20,
		MaxRedemptions: 10,
		ExpiresAt:      now.Add(time.Hour),
	}

	tests := []struct {
		name   string
		change func(*CreateInput)
	}{
		{"invalid code", func(input *CreateInput) { input.Code = "bad code" }},
		{"zero value", func(input *CreateInput) { input.Value = 0 }},
		{"percentage above 100", func(input *CreateInput) { input.Value = 101 }},
		{"percentage with currency", func(input *CreateInput) { input.Currency = &usd }},
		{"fixed without currency", func(input *CreateInput) { input.Type = DiscountFixed }},
		{"expired", func(input *CreateInput) { input.ExpiresAt = now }},
		{"zero redemption limit", func(input *CreateInput) { input.MaxRedemptions = 0 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput
			test.change(&input)
			service := NewService(&repositoryStub{})
			service.now = func() time.Time { return now }

			_, err := service.Create(context.Background(), input)
			if !errors.Is(err, ErrInvalidCoupon) {
				t.Fatalf("Create() error = %v, want ErrInvalidCoupon", err)
			}
		})
	}
}

func TestServiceApplyNormalizesInput(t *testing.T) {
	repository := &repositoryStub{
		applyResult: ApplyResult{InvoiceID: "inv-1"},
	}
	service := NewService(repository)

	_, err := service.Apply(context.Background(), ApplyInput{
		CouponCode: " welcome20 ",
		InvoiceID:  " inv-1 ",
		Amount:     10000,
		Currency:   " usd ",
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if repository.applied.CouponCode != "WELCOME20" ||
		repository.applied.InvoiceID != "inv-1" ||
		repository.applied.Currency != "USD" {
		t.Fatalf("Apply() input was not normalized: %+v", repository.applied)
	}
}

func TestServiceApplyRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input ApplyInput
	}{
		{"invalid coupon code", ApplyInput{CouponCode: "A", InvoiceID: "inv-1", Amount: 100, Currency: "USD"}},
		{"missing invoice ID", ApplyInput{CouponCode: "VALID", InvoiceID: "", Amount: 100, Currency: "USD"}},
		{"invalid amount", ApplyInput{CouponCode: "VALID", InvoiceID: "inv-1", Amount: 0, Currency: "USD"}},
		{"invalid currency", ApplyInput{CouponCode: "VALID", InvoiceID: "inv-1", Amount: 100, Currency: "US"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&repositoryStub{})
			_, err := service.Apply(context.Background(), test.input)
			if !errors.Is(err, ErrInvalidApplyInput) {
				t.Fatalf("Apply() error = %v, want ErrInvalidApplyInput", err)
			}
		})
	}
}
