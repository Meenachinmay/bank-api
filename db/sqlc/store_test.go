//go:build storetest
// +build storetest

package sqlc

import (
	"bank-api/util"
	"context"
	"database/sql"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestTransferTx(t *testing.T) {
	store := testStore
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)

	n := 5
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}
	// check result
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check account's balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	//
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := testStore
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)

	n := 10
	amount := int64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}
	// check result
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	//
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}

func TestUseReferralCodeTx(t *testing.T) {
	store := testStore

	n := 10

	referrerAccounts := make([]Account, n)
	for i := range referrerAccounts {
		referrerAccounts[i] = CreateUniqueRandomAccount(t)
	}

	// create referral codes for each referrer account
	for _, referrerAccount := range referrerAccounts {
		for i := 0; i < 5; i++ {
			referralCode := createUniqueRandomReferralCode(t, referrerAccount.ID)
			//Randomly set some referral codes as used
			if rand.Intn(2) == 0 { // 50% chance to set the code as used
				_, err := store.MarkReferralCodeUsed(context.Background(), MarkReferralCodeUsedParams{
					ReferralCode: referralCode.ReferralCode,
					UsedAt:       sql.NullTime{Time: utils.ConvertToTokyoTime(), Valid: true},
				})
				require.NoError(t, err)
			}
		}
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(referrerAccounts))
	results := make(chan UseReferralCodeTxResult, len(referrerAccounts))

	// Run the transaction for each referrer account
	for _, referrerAccount := range referrerAccounts {
		wg.Add(1)
		go func(account Account) {
			defer wg.Done()

			result, err := store.UseReferralCodeTx(context.Background(), UseReferralCodeTxParams{
				ReferrerAccountID: account.ID,
			})

			errs <- err
			results <- result
		}(referrerAccount)
	}

	wg.Wait()
	close(errs)
	close(results)

	for err := range errs {
		require.NoError(t, err)
	}

	// Check results for each referrer account
	for result := range results {
		require.NotEmpty(t, result)

		// Verify updated extra interest
		account, err := store.GetAccount(context.Background(), result.ReferrerAccountUpdate.ID)
		require.NoError(t, err)
		require.NotEmpty(t, account)

		// Fetch used referral count for this account
		startDate, endDate := getReferralDateRange()
		referralCount, err := store.GetReferralsByDateRange(context.Background(), GetReferralsByDateRangeParams{
			ReferrerAccountID: account.ID,
			CreatedAt:         startDate,
			CreatedAt_2:       endDate,
		})
		require.NoError(t, err)

		expectedExtraInterest := float64(referralCount)
		if expectedExtraInterest > 10.0 {
			expectedExtraInterest = 10.0
		}

		require.Equal(t, expectedExtraInterest, account.ExtraInterest.Float64)
		require.NotZero(t, account.ExtraInterestStartDate)
		require.Equal(t, int32(9), account.ExtraInterestDuration)
	}
}

// createUniqueRandomReferralCode creates a unique random referral code for the given referrer account.
func createUniqueRandomReferralCode(t *testing.T, referrerAccountID int64) ReferralCode {
	arg := CreateReferralCodeParams{
		ReferralCode:      util.RandomString(10),
		ReferrerAccountID: referrerAccountID,
		CreatedAt:         randomDateBetween(t, "2024-06-21", "2024-07-01"),
	}

	referralCode, err := testQueries.CreateReferralCode(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, referralCode)

	return referralCode
}

// randomDateBetween generates a random date between two dates
func randomDateBetween(t *testing.T, startDateStr, endDateStr string) time.Time {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	require.NoError(t, err)
	endDate, err := time.Parse(layout, endDateStr)
	require.NoError(t, err)

	delta := endDate.Sub(startDate)
	sec := rand.Int63n(int64(delta.Seconds()))
	return startDate.Add(time.Duration(sec) * time.Second)
}

// getReferralDateRange returns the start and end date for the referral code date range
func getReferralDateRange() (time.Time, time.Time) {
	currentDate := utils.ConvertToTokyoTime()
	loc, _ := time.LoadLocation("Asia/Tokyo")
	year, month, _ := currentDate.Date()

	var startDate, endDate time.Time
	if month == time.January {
		year--
		startDate = time.Date(year, time.December, 21, 0, 0, 0, 0, loc)
		endDate = time.Date(year, time.January, 20, 23, 59, 59, 0, loc)
	} else {
		startDate = time.Date(year, month-1, 21, 0, 0, 0, 0, loc)
		endDate = time.Date(year, month, 20, 23, 59, 59, 0, loc)
	}

	return startDate, endDate
}

func TestUseReferralCodeTxWithEdgeCases(t *testing.T) {
	store := testStore

	// Create a single referrer account
	referrerAccount := CreateUniqueRandomAccount(t)

	// Edge case 1: No referral codes created
	// Run the transaction
	_, err := store.UseReferralCodeTx(context.Background(), UseReferralCodeTxParams{
		ReferrerAccountID: referrerAccount.ID,
	})
	require.NoError(t, err)

	// Verify no extra interest
	account, err := store.GetAccount(context.Background(), referrerAccount.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Zero(t, account.ExtraInterest.Float64)

	// Edge case 2: Create referral codes on boundary dates
	referralCodeOnStartBoundary := createReferralCodeWithDate(t, referrerAccount.ID, "2024-06-21T00:00:00Z")
	referralCodeOnEndBoundary := createReferralCodeWithDate(t, referrerAccount.ID, "2024-07-20T23:59:59Z")

	// Set referral codes as used
	_, err = store.MarkReferralCodeUsed(context.Background(), MarkReferralCodeUsedParams{
		ReferralCode: referralCodeOnStartBoundary.ReferralCode,
		UsedAt:       sql.NullTime{Time: utils.ConvertToTokyoTime(), Valid: true},
	})
	require.NoError(t, err)

	_, err = store.MarkReferralCodeUsed(context.Background(), MarkReferralCodeUsedParams{
		ReferralCode: referralCodeOnEndBoundary.ReferralCode,
		UsedAt:       sql.NullTime{Time: utils.ConvertToTokyoTime(), Valid: true},
	})
	require.NoError(t, err)

	// Run the transaction
	_, err = store.UseReferralCodeTx(context.Background(), UseReferralCodeTxParams{
		ReferrerAccountID: referrerAccount.ID,
	})
	require.NoError(t, err)

	// Verify updated extra interest
	account, err = store.GetAccount(context.Background(), referrerAccount.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	expectedExtraInterest := 2.0
	require.Equal(t, expectedExtraInterest, account.ExtraInterest.Float64)
	require.NotZero(t, account.ExtraInterestStartDate)
	require.Equal(t, int32(9), account.ExtraInterestDuration)
}

// createReferralCodeWithDate creates a referral code with a specific creation date.
func createReferralCodeWithDate(t *testing.T, referrerAccountID int64, createdAt string) ReferralCode {
	createdAtTime, err := time.Parse(time.RFC3339, createdAt)
	require.NoError(t, err)

	arg := CreateReferralCodeParams{
		ReferralCode:      util.RandomString(10),
		ReferrerAccountID: referrerAccountID,
		CreatedAt:         createdAtTime,
	}

	referralCode, err := testQueries.CreateReferralCode(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, referralCode)

	return referralCode
}
