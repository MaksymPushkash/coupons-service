ALTER TABLE coupons
    ADD CONSTRAINT coupons_code_format_check
        CHECK (code ~ '^[A-Z0-9_-]{3,64}$'),
    ADD CONSTRAINT coupons_currency_format_check
        CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    ADD CONSTRAINT coupons_expiration_check
        CHECK (expires_at > created_at);

ALTER TABLE coupon_redemptions
    ADD CONSTRAINT coupon_redemptions_invoice_id_check
        CHECK (char_length(invoice_id) BETWEEN 1 AND 128),
    ADD CONSTRAINT coupon_redemptions_currency_format_check
        CHECK (currency ~ '^[A-Z]{3}$'),
    ADD CONSTRAINT coupon_redemptions_discount_check
        CHECK (discount_amount <= original_amount),
    ADD CONSTRAINT coupon_redemptions_total_check
        CHECK (original_amount = discount_amount + final_amount);

ALTER TABLE coupon_redemptions
    DROP CONSTRAINT coupon_redemptions_coupon_id_fkey,
    ADD CONSTRAINT coupon_redemptions_coupon_id_fkey
        FOREIGN KEY (coupon_id)
        REFERENCES coupons(id)
        ON DELETE RESTRICT;
