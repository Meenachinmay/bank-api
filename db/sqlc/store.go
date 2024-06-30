package sqlc

import (
	"4d63.com/tz"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Meenachinmay/microservice-shared/utils"
	"time"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID1 int64, amount1 int64, accountID2 int64, amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}

type UseReferralCodeTxParams struct {
	ReferralCode      string `json:"referral_code"`
	ReferredAccountID int64  `json:"referred_account_id"`
}

type UseReferralCodeTxResult struct {
	ReferralCode    ReferralCode    `json:"referral_code"`
	ReferralHistory ReferralHistory `json:"referral_history"`
}

func (store *Store) UseReferralCodeTx(ctx context.Context, arg UseReferralCodeTxParams) (UseReferralCodeTxResult, error) {
	var result UseReferralCodeTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.ReferralCode, err = q.GetReferralCode(ctx, arg.ReferralCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("referral code not found")
			}
			return err
		}

		if result.ReferralCode.IsUsed {
			return fmt.Errorf("referral code already used")
		}

		result.ReferralCode, err = q.MarkReferralCodeUsed(ctx, MarkReferralCodeUsedParams{
			ReferralCode: result.ReferralCode.ReferralCode,
			UsedAt:       sql.NullTime{Time: utils.ConvertToTokyoTime(), Valid: true},
		})
		if err != nil {
			return err
		}

		// TODO: update the referral history table
		result.ReferralHistory, err = q.CreateReferralHistory(ctx, CreateReferralHistoryParams{
			ReferrerAccountID: result.ReferralCode.ReferrerAccountID,
			ReferredAccountID: arg.ReferredAccountID,
			ReferralCodeID:    result.ReferralCode.ID,
			ReferralDate:      result.ReferralCode.CreatedAt,
		})
		if err != nil {
			return err
		}

		// TODO: perform logic to give benefit to referrer_account_id
		err = store.updateExtraInterest(ctx, q, result.ReferralCode.ReferrerAccountID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (store *Store) updateExtraInterest(ctx context.Context, q *Queries, referrerAccountID int64) error {
	account, err := q.GetAccountForUpdate(context.Background(), referrerAccountID)
	if err != nil {
		return err
	}

	var currentExtraInterest, newExtraInterest float64
	if account.ExtraInterest.Valid {
		currentExtraInterest = account.ExtraInterest.Float64
	}

	currentDate := utils.ConvertToTokyoTime()
	loc, err := tz.LoadLocation("Asia/Tokyo")
	if err != nil {
		return err
	}
	startDate := time.Date(currentDate.Year(), currentDate.Month()-1, 21, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
	endDate := time.Date(currentDate.Year(), currentDate.Month(), 20, 23, 59, 59, 0, loc)

	args := GetReferralsByDateRangeParams{
		ReferrerAccountID: referrerAccountID,
		CreatedAt:         startDate,
		CreatedAt_2:       endDate,
	}

	referralCount, err := q.GetReferralsByDateRange(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("referral code not found")
		}
		return err
	}

	// updating the new extra interest
	if currentExtraInterest >= 10.0 {
		newExtraInterest = currentExtraInterest
	} else {
		newExtraInterest = float64(referralCount)
	}

	// update the new interest here
	updateInterestArgs := UpdateAccountInterestParams{
		ID:            referrerAccountID,
		ExtraInterest: sql.NullFloat64{Float64: newExtraInterest, Valid: true},
	}
	account, err = q.UpdateAccountInterest(ctx, updateInterestArgs)

	return nil
}
