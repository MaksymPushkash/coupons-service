//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MaksymPushkash/coupons-service/internal/coupon"
)

func TestCouponRepositoryLifecycle(t *testing.T) {
	ctx := t.Context()
	pool := openIntegrationPool(t)
	repository := NewCouponRepository(pool)
	now := time.Now().UTC()
	value := testCoupon("WELCOME20", 2, now)

	if err := repository.Create(ctx, value); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	duplicate := value
	duplicate.ID = uuid.NewString()
	if err := repository.Create(ctx, duplicate); !errors.Is(err, coupon.ErrCouponCodeExists) {
		t.Fatalf("duplicate Create() error = %v, want ErrCouponCodeExists", err)
	}

	stored, err := repository.GetByCode(ctx, "welcome20")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if stored.Code != value.Code {
		t.Fatalf("GetByCode() code = %q, want %q", stored.Code, value.Code)
	}

	input := coupon.ApplyInput{
		CouponCode: value.Code,
		InvoiceID:  "inv-1",
		Amount:     10000,
		Currency:   "USD",
	}
	result, err := repository.Apply(ctx, input, now)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if result.DiscountAmount != 2000 || result.FinalAmount != 8000 {
		t.Fatalf("Apply() result = %+v", result)
	}

	_, err = repository.Apply(ctx, input, now)
	if !errors.Is(err, coupon.ErrAlreadyApplied) {
		t.Fatalf("duplicate Apply() error = %v, want ErrAlreadyApplied", err)
	}

	stored, err = repository.GetByCode(ctx, value.Code)
	if err != nil {
		t.Fatalf("GetByCode() after apply error = %v", err)
	}
	if stored.RedeemedCount != 1 {
		t.Fatalf("RedeemedCount = %d, want 1", stored.RedeemedCount)
	}

	if err := repository.Deactivate(ctx, value.Code); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}

	input.InvoiceID = "inv-2"
	_, err = repository.Apply(ctx, input, now)
	if !errors.Is(err, coupon.ErrCouponInactive) {
		t.Fatalf("inactive Apply() error = %v, want ErrCouponInactive", err)
	}
}

func TestCouponRepositoryConcurrentRedemptionLimit(t *testing.T) {
	ctx := t.Context()
	pool := openIntegrationPool(t)
	repository := NewCouponRepository(pool)
	now := time.Now().UTC()
	value := testCoupon("LIMIT5", 5, now)

	if err := repository.Create(ctx, value); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	const requests = 20
	results := make(chan error, requests)
	var waitGroup sync.WaitGroup

	for index := 0; index < requests; index++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			_, err := repository.Apply(ctx, coupon.ApplyInput{
				CouponCode: value.Code,
				InvoiceID:  "inv-" + uuid.NewString(),
				Amount:     10000 + int64(index),
				Currency:   "USD",
			}, now)
			results <- err
		}(index)
	}

	waitGroup.Wait()
	close(results)

	successes := 0
	limitErrors := 0

	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, coupon.ErrRedemptionLimitReached):
			limitErrors++
		default:
			t.Fatalf("Apply() unexpected error = %v", err)
		}
	}

	if successes != 5 || limitErrors != requests-5 {
		t.Fatalf("successes = %d, limit errors = %d", successes, limitErrors)
	}

	stored, err := repository.GetByCode(ctx, value.Code)
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if stored.RedeemedCount != 5 {
		t.Fatalf("RedeemedCount = %d, want 5", stored.RedeemedCount)
	}
}

func openIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := t.Context()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create admin pool: %v", err)
	}

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		adminPool.Close()
		t.Fatalf("create test schema: %v", err)
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database config: %v", err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	config.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
	})

	for _, name := range []string{
		"000001_create_coupons.up.sql",
		"000002_strengthen_constraints.up.sql",
	} {
		migrationPath := filepath.Join("..", "..", "..", "migrations", name)
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}

	return pool
}

func testCoupon(code string, maxRedemptions int, now time.Time) coupon.Coupon {
	return coupon.Coupon{
		ID:             uuid.NewString(),
		Code:           code,
		Type:           coupon.DiscountPercentage,
		Value:          20,
		MaxRedemptions: maxRedemptions,
		ExpiresAt:      now.Add(time.Hour),
		Active:         true,
		CreatedAt:      now,
	}
}
