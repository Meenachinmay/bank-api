// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"database/sql"
	"time"
)

type Account struct {
	ID        int64     `json:"id"`
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type Entry struct {
	ID        int64 `json:"id"`
	AccountID int64 `json:"account_id"`
	// can be negative or positive
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type ReferralCode struct {
	ID             int64        `json:"id"`
	ReferralCode   string       `json:"referral_code"`
	ReferrerUserID int64        `json:"referrer_user_id"`
	IsUsed         sql.NullBool `json:"is_used"`
	CreatedAt      time.Time    `json:"created_at"`
	UsedAt         sql.NullTime `json:"used_at"`
}

type ReferralHistory struct {
	ID             int64     `json:"id"`
	ReferrerUserID int64     `json:"referrer_user_id"`
	ReferredUserID int64     `json:"referred_user_id"`
	ReferralCodeID int64     `json:"referral_code_id"`
	ReferralDate   time.Time `json:"referral_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type Transfer struct {
	ID            int64 `json:"id"`
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	// must be positive
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID            int64           `json:"id"`
	UserName      string          `json:"user_name"`
	Email         string          `json:"email"`
	ExtraInterest sql.NullFloat64 `json:"extra_interest"`
	CreatedAt     time.Time       `json:"created_at"`
}
