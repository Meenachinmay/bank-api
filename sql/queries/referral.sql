-- name: CreateReferralCode :one
INSERT INTO referral_codes (referral_code, referrer_account_id, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetReferralCode :one
SELECT * FROM referral_codes
WHERE referral_code = $1
LIMIT 1;

-- name: GetReferralCodesForReferrerAccount :many
SELECT * FROM referral_codes
WHERE referrer_account_id = $1
LIMIT 10;

-- name: HasUnUsedCodeForReferrerAccount :one
SELECT EXISTS (
    SELECT 1
    FROM referral_codes
    WHERE referrer_account_id = $1
      AND is_used = false
);

-- name: GetReferralsByDateRange :one
SELECT COUNT(*) FROM referral_codes
    WHERE referrer_account_id = $1
    AND is_used = true
    AND created_at >= $2 AND created_at <= $3;

-- name: GetUnusedReferralCodes :many
SELECT * FROM referral_codes
WHERE is_used = true
  AND referrer_account_id = $1
  AND created_at >= $2 AND created_at <= $3;

-- name: MarkReferralCodeUsed :one
UPDATE referral_codes
SET is_used = true, used_at = $2
WHERE referral_code = $1
RETURNING *;

-- name: CreateReferralHistory :one
INSERT INTO referral_history (referrer_account_id, referred_account_id, referral_code_id, referral_date, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetReferralHistory :many
SELECT * FROM referral_history
WHERE referrer_account_id = $1
ORDER BY referral_date;

-- name: GetReferralHistoryByDate :many
SELECT * FROM referral_history
WHERE referrer_account_id = $1
  AND referral_date >= $2 AND referral_date <= $3
ORDER BY referral_date;