ALTER TABLE coupon_redemptions
    DROP CONSTRAINT coupon_redemptions_coupon_id_fkey,
    ADD CONSTRAINT coupon_redemptions_coupon_id_fkey
        FOREIGN KEY (coupon_id)
        REFERENCES coupons(id)
        ON DELETE CASCADE;

ALTER TABLE coupon_redemptions
    DROP CONSTRAINT coupon_redemptions_total_check,
    DROP CONSTRAINT coupon_redemptions_discount_check,
    DROP CONSTRAINT coupon_redemptions_currency_format_check,
    DROP CONSTRAINT coupon_redemptions_invoice_id_check;

ALTER TABLE coupons
    DROP CONSTRAINT coupons_expiration_check,
    DROP CONSTRAINT coupons_currency_format_check,
    DROP CONSTRAINT coupons_code_format_check;
