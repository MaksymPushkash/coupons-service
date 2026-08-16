CREATE TABLE coupons (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL,
    discount_type TEXT NOT NULL,
    value BIGINT NOT NULL,
    currency VARCHAR(3),
    max_redemptions INTEGER NOT NULL,
    redeemed_count INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT coupons_discount_type_check
        CHECK (discount_type IN ('percentage', 'fixed')),

    CONSTRAINT coupons_value_check
        CHECK (value > 0),

    CONSTRAINT coupons_percentage_value_check
        CHECK (
            discount_type <> 'percentage'
            OR value BETWEEN 1 AND 100
        ),

    CONSTRAINT coupons_currency_check
        CHECK (
            (discount_type = 'fixed' AND currency IS NOT NULL)
            OR
            (discount_type = 'percentage' AND currency IS NULL)
        ),

    CONSTRAINT coupons_max_redemptions_check
        CHECK (max_redemptions > 0),

    CONSTRAINT coupons_redeemed_count_check
        CHECK (
            redeemed_count >= 0
            AND redeemed_count <= max_redemptions
        )
);

CREATE UNIQUE INDEX coupons_code_unique_ci
    ON coupons (UPPER(code));

CREATE TABLE coupon_redemptions (
    id UUID PRIMARY KEY,
    coupon_id UUID NOT NULL
        REFERENCES coupons(id)
        ON DELETE CASCADE,

    invoice_id TEXT NOT NULL,
    original_amount BIGINT NOT NULL,
    discount_amount BIGINT NOT NULL,
    final_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT coupon_redemptions_amount_check
        CHECK (
            original_amount > 0
            AND discount_amount >= 0
            AND final_amount >= 0
        ),

    CONSTRAINT coupon_redemptions_unique_invoice
        UNIQUE (coupon_id, invoice_id)
);

CREATE INDEX coupon_redemptions_coupon_id_index
    ON coupon_redemptions (coupon_id);