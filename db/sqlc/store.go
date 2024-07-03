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
	ReferrerAccountID int64 `json:"referrer_account_id"`
}

type UseReferralCodeTxResult struct {
	ReferrerAccountUpdate Account `json:"referrer_account"`
}

// UseReferralCodeTx Calculate Interest for the following month
func (store *Store) UseReferralCodeTx(ctx context.Context, arg UseReferralCodeTxParams) (UseReferralCodeTxResult, error) {
	var result UseReferralCodeTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// TODO: perform logic to give benefit to referrer_account_id
		result.ReferrerAccountUpdate, err = q.GetAccountForUpdate(context.Background(), arg.ReferrerAccountID)
		if err != nil {
			return err
		}

		var currentExtraInterest, newExtraInterest float64
		if result.ReferrerAccountUpdate.ExtraInterest.Valid {
			currentExtraInterest = result.ReferrerAccountUpdate.ExtraInterest.Float64
		}

		loc, err := tz.LoadLocation("Asia/Tokyo")
		if err != nil {
			return err
		}

		currentDate := utils.ConvertToTokyoTime()
		year, month, _ := currentDate.Date()

		// Determine the date range for the previous month
		var startDate, endDate time.Time
		if month == time.January {
			year--
			startDate = time.Date(year, time.December, 21, 0, 0, 0, 0, loc)
			endDate = time.Date(year, time.January, 20, 23, 59, 59, 0, loc)
		} else {
			startDate = time.Date(year, month-1, 21, 0, 0, 0, 0, loc)
			endDate = time.Date(year, month, 20, 23, 59, 59, 0, loc)
		}

		args := GetReferralsByDateRangeParams{
			ReferrerAccountID: result.ReferrerAccountUpdate.ID,
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
		if referralCount > 0 {
			if currentExtraInterest >= 10.0 {
				newExtraInterest = currentExtraInterest
			} else {
				if referralCount == int64(currentExtraInterest) {
					newExtraInterest = currentExtraInterest
				} else {
					newExtraInterest = float64(referralCount)
					if newExtraInterest > 10.0 {
						newExtraInterest = 10.0
					}
				}
			}

			extraInterestStartDate := getFirstDayOfNextMonth(currentDate)

			if newExtraInterest > 0 {
				// update the new interest here
				updateInterestArgs := UpdateAccountInterestParams{
					ID:                     result.ReferrerAccountUpdate.ID,
					ExtraInterest:          sql.NullFloat64{Float64: newExtraInterest, Valid: true},
					ExtraInterestStartDate: sql.NullTime{Time: extraInterestStartDate, Valid: true},
					ExtraInterestDuration:  9,
				}

				result.ReferrerAccountUpdate, err = q.UpdateAccountInterest(ctx, updateInterestArgs)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return result, err
}

func getFirstDayOfNextMonth(currentDate time.Time) time.Time {
	year, month, _ := currentDate.Date()
	if month == time.December {
		year++
		month = time.January
	} else {
		month++
	}
	loc, err := tz.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}
	}

	return time.Date(year, month, 1, 0, 0, 0, 0, loc)
}
