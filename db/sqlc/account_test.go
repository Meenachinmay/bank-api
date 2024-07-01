package sqlc

import (
	"4d63.com/tz"
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

var (
	generatedAccounts = make(map[int64]struct{})
	generatedCodes    = make(map[int64]struct{})
	mu                sync.Mutex
)

func CreateUniqueRandomAccount(t *testing.T) Account {
	mu.Lock()
	defer mu.Unlock()
	var account Account
	for {
		account = CreateRandomAccount(t)
		if _, exists := generatedAccounts[account.ID]; !exists {
			generatedAccounts[account.ID] = struct{}{}
			break
		}
	}
	return account
}

func CreateUniqueRandomReferralCode(t *testing.T, referrerAccountID int64) ReferralCode {
	mu.Lock()
	defer mu.Unlock()
	var referralCode ReferralCode
	for {
		referralCode = CreateRandomReferralCode(t, referrerAccountID)
		if _, exists := generatedCodes[referralCode.ID]; !exists {
			generatedCodes[referralCode.ID] = struct{}{}
			break
		}
	}
	return referralCode
}

func CreateRandomReferralCode(t *testing.T, referrerAccountID int64) ReferralCode {
	loc, err := tz.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	currentDate := utils.ConvertToTokyoTime()
	startDate := time.Date(currentDate.Year(), currentDate.Month(), 21, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
	endDate := time.Date(currentDate.Year(), currentDate.Month(), 20, 23, 59, 59, 0, loc)

	randomDate := randomTimeBetween(t, startDate, endDate)

	arg := CreateReferralCodeParams{
		ReferralCode:      util.RandomString(10),
		ReferrerAccountID: referrerAccountID,
		CreatedAt:         randomDate,
	}

	referralCode, err := testQueries.CreateReferralCode(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, referralCode)

	return referralCode
}

func randomTimeBetween(t *testing.T, start, end time.Time) time.Time {
	delta := end.Sub(start)
	sec := rand.Int63n(int64(delta.Seconds()))
	randomDate := start.Add(time.Duration(sec) * time.Second)
	return randomDate
}

func CreateRandomAccount(t *testing.T) Account {
	args := CreateAccountParams{
		Owner:     util.RandomOwner(),
		Balance:   util.RandomMoney(),
		Email:     util.RandomEmail(),
		Currency:  util.RandomCurrency(),
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	account, err := testQueries.CreateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, args.Owner, account.Owner)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Email, account.Email)
	require.Equal(t, args.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	CreateRandomAccount(t)
}

func TestCreateUniqueAccount(t *testing.T) {
	CreateUniqueRandomAccount(t)
}

func TestCreateReferralCode(t *testing.T) {
	account := CreateRandomAccount(t)
	CreateUniqueRandomReferralCode(t, account.ID)
}

func TestGetAccount(t *testing.T) {
	// create account
	account1 := CreateRandomAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	// create account
	account1 := CreateRandomAccount(t)

	arg := UpdateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomMoney(),
	}
	account2, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, arg.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	// create account
	account1 := CreateRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.Error(t, err)
	require.Equal(t, err, sql.ErrNoRows)
	require.Empty(t, account2)
}
