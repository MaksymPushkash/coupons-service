package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MaksymPushkash/coupons-service/internal/coupon"
)

const (
	couponCodeUniqueConstraint = "coupons_code_unique_ci"
	redemptionUniqueConstraint = "coupon_redemptions_unique_invoice"
)

type CouponRepository struct {
	pool *pgxpool.Pool
}

func NewCouponRepository(pool *pgxpool.Pool) *CouponRepository {
	return &CouponRepository{pool: pool}
}

func (r *CouponRepository) Create(ctx context.Context, value coupon.Coupon) error {
	const query = `
		INSERT INTO coupons (
			id,
			code,
			discount_type,
			value,
			currency,
			max_redemptions,
			redeemed_count,
			expires_at,
			active,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		value.ID,
		value.Code,
		value.Type,
		value.Value,
		nullableString(value.Currency),
		value.MaxRedemptions,
		value.RedeemedCount,
		value.ExpiresAt,
		value.Active,
		value.CreatedAt,
	)

	if isUniqueViolation(err, couponCodeUniqueConstraint) {
		return coupon.ErrCouponCodeExists
	}

	if err != nil {
		return fmt.Errorf("insert coupon: %w", err)
	}

	return nil
}

func (r *CouponRepository) GetByCode(ctx context.Context, code string) (coupon.Coupon, error) {
	const query = `
		SELECT
			id::text,
			code,
			discount_type,
			value,
			currency,
			max_redemptions,
			redeemed_count,
			expires_at,
			active,
			created_at
		FROM coupons
		WHERE UPPER(code) = UPPER($1)
	`

	value, err := scanCoupon(r.pool.QueryRow(ctx, query, code))

	if errors.Is(err, pgx.ErrNoRows) {
		return coupon.Coupon{}, coupon.ErrCouponNotFound
	}

	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("select coupon: %w", err)
	}

	return value, nil
}

func (r *CouponRepository) Apply(
	ctx context.Context,
	input coupon.ApplyInput,
	now time.Time,
) (coupon.ApplyResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return coupon.ApplyResult{}, fmt.Errorf("begin apply transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	value, err := getCouponForUpdate(ctx, tx, input.CouponCode)

	if errors.Is(err, pgx.ErrNoRows) {
		return coupon.ApplyResult{}, coupon.ErrCouponNotFound
	}

	if err != nil {
		return coupon.ApplyResult{}, fmt.Errorf("select coupon for update: %w", err)
	}

	result, err := coupon.CalculateDiscount(value, input, now)
	if err != nil {
		return coupon.ApplyResult{}, err
	}

	if err := insertRedemption(ctx, tx, value.ID, result, now); err != nil {
		return coupon.ApplyResult{}, err
	}

	if err := incrementRedemptionCount(ctx, tx, value.ID); err != nil {
		return coupon.ApplyResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return coupon.ApplyResult{}, fmt.Errorf("commit apply transaction: %w", err)
	}

	return result, nil
}

func (r *CouponRepository) Deactivate(ctx context.Context, code string) error {
	const query = `
		UPDATE coupons
		SET active = FALSE
		WHERE UPPER(code) = UPPER($1)
	`

	result, err := r.pool.Exec(ctx, query, code)
	if err != nil {
		return fmt.Errorf("deactivate coupon: %w", err)
	}

	if result.RowsAffected() == 0 {
		return coupon.ErrCouponNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCoupon(row rowScanner) (coupon.Coupon, error) {
	var value coupon.Coupon
	var currency sql.NullString

	err := row.Scan(
		&value.ID,
		&value.Code,
		&value.Type,
		&value.Value,
		&currency,
		&value.MaxRedemptions,
		&value.RedeemedCount,
		&value.ExpiresAt,
		&value.Active,
		&value.CreatedAt,
	)
	if err != nil {
		return coupon.Coupon{}, err
	}

	if currency.Valid {
		value.Currency = &currency.String
	}

	return value, nil
}

func getCouponForUpdate(ctx context.Context, tx pgx.Tx, code string) (coupon.Coupon, error) {
	const query = `
		SELECT
			id::text,
			code,
			discount_type,
			value,
			currency,
			max_redemptions,
			redeemed_count,
			expires_at,
			active,
			created_at
		FROM coupons
		WHERE UPPER(code) = UPPER($1)
		FOR UPDATE
	`

	return scanCoupon(tx.QueryRow(ctx, query, code))
}

func insertRedemption(
	ctx context.Context,
	tx pgx.Tx,
	couponID string,
	result coupon.ApplyResult,
	createdAt time.Time,
) error {
	const query = `
		INSERT INTO coupon_redemptions (
			id,
			coupon_id,
			invoice_id,
			original_amount,
			discount_amount,
			final_amount,
			currency,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := tx.Exec(
		ctx,
		query,
		uuid.NewString(),
		couponID,
		result.InvoiceID,
		result.OriginalAmount,
		result.DiscountAmount,
		result.FinalAmount,
		result.Currency,
		createdAt,
	)
	if isUniqueViolation(err, redemptionUniqueConstraint) {
		return coupon.ErrAlreadyApplied
	}
	if err != nil {
		return fmt.Errorf("insert redemption: %w", err)
	}

	return nil
}

func incrementRedemptionCount(ctx context.Context, tx pgx.Tx, couponID string) error {
	const query = `
		UPDATE coupons
		SET redeemed_count = redeemed_count + 1
		WHERE id = $1
	`

	if _, err := tx.Exec(ctx, query, couponID); err != nil {
		return fmt.Errorf("update redemption count: %w", err)
	}
	return nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func isUniqueViolation(err error, constraint string) bool {
	var postgresError *pgconn.PgError

	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == constraint
}
