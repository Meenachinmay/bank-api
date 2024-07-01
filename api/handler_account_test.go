package api

import (
	"4d63.com/tz"
	"bank-api/db/sqlc"
	"bank-api/util"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store *sqlc.Store) *Server {
	server := NewServer(store)
	return server
}

var (
	generatedAccounts = make(map[int64]struct{})
	generatedCodes    = make(map[int64]struct{})
	mu                sync.Mutex
)

func CreateUniqueRandomAccount(t *testing.T) sqlc.Account {
	mu.Lock()
	defer mu.Unlock()
	var account sqlc.Account
	for {
		account = CreateRandomAccount(t)
		if _, exists := generatedAccounts[account.ID]; !exists {
			generatedAccounts[account.ID] = struct{}{}
			break
		}
	}
	return account
}

func CreateUniqueRandomReferralCode(t *testing.T, referrerAccountID int64) sqlc.ReferralCode {
	mu.Lock()
	defer mu.Unlock()
	var referralCode sqlc.ReferralCode
	for {
		referralCode = CreateRandomReferralCode(t, referrerAccountID)
		if _, exists := generatedCodes[referralCode.ID]; !exists {
			generatedCodes[referralCode.ID] = struct{}{}
			break
		}
	}
	return referralCode
}

func CreateRandomReferralCode(t *testing.T, referrerAccountID int64) sqlc.ReferralCode {
	loc, err := tz.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	currentDate := utils.ConvertToTokyoTime()
	startDate := time.Date(currentDate.Year(), currentDate.Month(), 21, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
	endDate := time.Date(currentDate.Year(), currentDate.Month(), 20, 23, 59, 59, 0, loc)

	randomDate := randomTimeBetween(t, startDate, endDate)

	arg := sqlc.CreateReferralCodeParams{
		ReferralCode:      util.RandomString(10),
		ReferrerAccountID: referrerAccountID,
		CreatedAt:         randomDate,
	}

	referralCode, err := testStore.CreateReferralCode(context.Background(), arg)
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

func CreateRandomAccount(t *testing.T) sqlc.Account {
	args := sqlc.CreateAccountParams{
		Owner:     util.RandomOwner(),
		Balance:   util.RandomMoney(),
		Email:     util.RandomEmail(),
		Currency:  util.RandomCurrency(),
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	account, err := testStore.CreateAccount(context.Background(), args)
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
	server := newTestServer(t, testStore)

	account := sqlc.Account{
		Owner:     "John Doe",
		Balance:   0,
		Email:     "johndoe@gmail.com",
		Currency:  "YEN",
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	data, err := json.Marshal(createAccountRequest{
		Owner:     account.Owner,
		Email:     account.Email,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
	})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/accounts", bytes.NewReader(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var createdAccount sqlc.Account
	err = json.Unmarshal(recorder.Body.Bytes(), &createdAccount)
	require.NoError(t, err)

	require.NotZero(t, createdAccount.ID)
	require.Equal(t, account.Owner, createdAccount.Owner)
	require.Equal(t, account.Email, createdAccount.Email)
	require.Equal(t, account.Currency, createdAccount.Currency)
	require.Equal(t, account.Balance, createdAccount.Balance)
	require.NotZero(t, createdAccount.CreatedAt)
}

func TestCreateAccountWithReferralCode(t *testing.T) {
	account := CreateUniqueRandomAccount(t)
	referralCode := CreateUniqueRandomReferralCode(t, account.ID)

	server := newTestServer(t, testStore)

	recorder := httptest.NewRecorder()
	url := "/accounts"

	// Define the request body
	reqBody := createAccountRequest{
		Owner:        "John Doe",
		Currency:     "YEN",
		Email:        "johndoe@example.com",
		ReferralCode: referralCode.ReferralCode,
	}
	jsonReq, err := json.Marshal(reqBody)
	require.NoError(t, err)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var createdAccount sqlc.Account
	err = json.Unmarshal(recorder.Body.Bytes(), &createdAccount)
	require.NoError(t, err)

	require.NotZero(t, createdAccount.ID)
	require.Equal(t, "John Doe", createdAccount.Owner)
	require.Equal(t, "YEN", createdAccount.Currency)
	require.Equal(t, "johndoe@example.com", createdAccount.Email)
	require.Equal(t, int64(1000), createdAccount.Balance)

	// Check if the referral code is marked as used
	usedReferralCode, err := server.store.GetReferralCode(context.Background(), referralCode.ReferralCode)
	require.NoError(t, err)
	require.True(t, usedReferralCode.IsUsed)
	require.NotZero(t, usedReferralCode.UsedAt)
}

func TestGetAccount(t *testing.T) {
	server := newTestServer(t, testStore)

	// Create an account to retrieve
	account := sqlc.Account{
		Owner:     "John Doe",
		Balance:   0,
		Currency:  "YEN",
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	// Create account using the store directly to ensure it exists
	arg := sqlc.CreateAccountParams{
		Owner:     account.Owner,
		Currency:  account.Currency,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt,
	}
	createdAccount, err := testStore.CreateAccount(context.Background(), arg)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	url := fmt.Sprintf("/accounts/%d", createdAccount.ID)
	request, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var retrievedAccount sqlc.Account
	err = json.Unmarshal(recorder.Body.Bytes(), &retrievedAccount)
	require.NoError(t, err)

	require.Equal(t, createdAccount.ID, retrievedAccount.ID)
	require.Equal(t, createdAccount.Owner, retrievedAccount.Owner)
	require.Equal(t, createdAccount.Currency, retrievedAccount.Currency)
	require.Equal(t, createdAccount.Balance, retrievedAccount.Balance)
	require.NotZero(t, retrievedAccount.CreatedAt)
}

func TestCreateReferral(t *testing.T) {
	account := CreateUniqueRandomAccount(t)

	server := newTestServer(t, testStore)

	recorder := httptest.NewRecorder()
	url := fmt.Sprintf("/referral/account/%d", account.ID)

	request, err := http.NewRequest("POST", url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var referralCode sqlc.ReferralCode
	err = json.Unmarshal(recorder.Body.Bytes(), &referralCode)
	require.NoError(t, err)
	require.NotEmpty(t, referralCode.ReferralCode)
	require.Equal(t, account.ID, referralCode.ReferrerAccountID)
	require.Equal(t, false, referralCode.IsUsed)
	require.NotZero(t, referralCode.ReferralCode)
}

func TestUseReferralCode(t *testing.T) {
	referrerAccount := CreateUniqueRandomAccount(t)
	referredAccount := CreateUniqueRandomAccount(t)

	referralCode := CreateUniqueRandomReferralCode(t, referrerAccount.ID)

	server := newTestServer(t, testStore)

	recorder := httptest.NewRecorder()
	url := fmt.Sprintf("/referral/code/%s", referralCode.ReferralCode)
	jsonReq := fmt.Sprintf(`{"referred_account_id": %d}`, referredAccount.ID)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonReq)))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var result sqlc.ReferralCode
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Equal(t, referrerAccount.ID, result.ReferrerAccountID)
	require.Equal(t, referrerAccount.ID, result.ReferrerAccountID)
	require.NotZero(t, result.UsedAt.Time)

}

func TestLoginAccount(t *testing.T) {
	account := CreateUniqueRandomAccount(t)

	log.Printf(">> account: %+v", account)

	server := newTestServer(t, testStore)
	loginReq := loginAccountRequest{
		Email: account.Email,
	}

	recorder := httptest.NewRecorder()
	url := fmt.Sprintf("/accounts/login")
	jsonReq, err := json.Marshal(loginReq)
	require.NoError(t, err)

	log.Printf(">> json request: %+v", string(jsonReq))

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var result sqlc.Account
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Equal(t, account.ID, result.ID)
	require.Equal(t, account.Email, result.Email)
	require.Equal(t, account.Currency, result.Currency)
}

func TestGetReferralCodesForAccount(t *testing.T) {
	account := CreateUniqueRandomAccount(t)

	referralCodes := make([]sqlc.ReferralCode, 4)
	for i := 0; i < 4; i++ { // Changed <= to < to avoid index out of range error
		referralCodes[i] = CreateUniqueRandomReferralCode(t, account.ID)
	}

	log.Printf(">> referral codes: %+v\n", referralCodes)

	server := newTestServer(t, testStore)
	recorder := httptest.NewRecorder()
	url := fmt.Sprintf("/referral-codes?account=%d", account.ID)

	request, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var gotReferralCodes []sqlc.ReferralCode
	err = json.Unmarshal(recorder.Body.Bytes(), &gotReferralCodes)
	require.NoError(t, err)

	log.Printf("got referral codes: %+v\n", gotReferralCodes)

	require.Equal(t, len(referralCodes), len(gotReferralCodes))
	for i := range referralCodes {
		require.Equal(t, referralCodes[i].ReferralCode, gotReferralCodes[i].ReferralCode)
		require.Equal(t, referralCodes[i].ReferrerAccountID, gotReferralCodes[i].ReferrerAccountID)
		require.Equal(t, referralCodes[i].IsUsed, gotReferralCodes[i].IsUsed)
		require.WithinDuration(t, referralCodes[i].CreatedAt, gotReferralCodes[i].CreatedAt, time.Second)
	}
}
