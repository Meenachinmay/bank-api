-- name: CreateReferralCode :one
INSERT INTO referral_codes (referral_code, referrer_user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetReferralCode :one
SELECT * FROM referral_codes
WHERE referral_code = $1
LIMIT 1;

-- name: MarkReferralCodeUsed :one
UPDATE referral_codes
SET is_used = true, used_at = $2
WHERE referral_code = $1
RETURNING *;

-- name: CreateReferralHistory :one
INSERT INTO referral_history (referrer_user_id, referred_user_id, referral_code_id, referral_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetReferralHistory :many
SELECT * FROM referral_history
WHERE referrer_user_id = $1
ORDER BY referral_date;

-- name: GetReferralHistoryByDate :many
SELECT * FROM referral_history
WHERE referrer_user_id = $1
  AND referral_date >= $2 AND referral_date <= $3
ORDER BY referral_date;