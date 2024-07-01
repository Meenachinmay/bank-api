package api

//
//import (
//	"4d63.com/tz"
//	"bank-api/db/sqlc"
//	"bank-api/util"
//	"bytes"
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/Meenachinmay/microservice-shared/utils"
//	"github.com/stretchr/testify/require"
//	"math/rand"
//	"net/http"
//	"net/http/httptest"
//	"sync"
//	"testing"
//	"time"
//)
//
//func newTestServer(t *testing.T, store *sqlc.Store) *Server {
//	server := NewServer(store)
//	return server
//}
//
//var (
//	generatedAccounts = make(map[int64]struct{})
//	generatedCodes    = make(map[int64]struct{})
//	mu                sync.Mutex
//)
//
//func CreateUniqueRandomAccount(t *testing.T) sqlc.Account {
//	mu.Lock()
//	defer mu.Unlock()
//	var account sqlc.Account
//	for {
//		account = CreateRandomAccount(t)
//		if _, exists := generatedAccounts[account.ID]; !exists {
//			generatedAccounts[account.ID] = struct{}{}
//			break
//		}
//	}
//	return account
//}
//
//func CreateUniqueRandomReferralCode(t *testing.T, referrerAccountID int64) sqlc.ReferralCode {
//	mu.Lock()
//	defer mu.Unlock()
//	var referralCode sqlc.ReferralCode
//	for {
//		referralCode = CreateRandomReferralCode(t, referrerAccountID)
//		if _, exists := generatedCodes[referralCode.ID]; !exists {
//			generatedCodes[referralCode.ID] = struct{}{}
//			break
//		}
//	}
//	return referralCode
//}
//
//func CreateRandomReferralCode(t *testing.T, referrerAccountID int64) sqlc.ReferralCode {
//	loc, err := tz.LoadLocation("Asia/Tokyo")
//	require.NoError(t, err)
//
//	currentDate := utils.ConvertToTokyoTime()
//	startDate := time.Date(currentDate.Year(), currentDate.Month(), 21, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
//	endDate := time.Date(currentDate.Year(), currentDate.Month(), 20, 23, 59, 59, 0, loc)
//
//	randomDate := randomTimeBetween(t, startDate, endDate)
//
//	arg := sqlc.CreateReferralCodeParams{
//		ReferralCode:      util.RandomString(10),
//		ReferrerAccountID: referrerAccountID,
//		CreatedAt:         randomDate,
//	}
//
//	referralCode, err := testStore.CreateReferralCode(context.Background(), arg)
//	require.NoError(t, err)
//	require.NotEmpty(t, referralCode)
//
//	return referralCode
//}
//
//func randomTimeBetween(t *testing.T, start, end time.Time) time.Time {
//	delta := end.Sub(start)
//	sec := rand.Int63n(int64(delta.Seconds()))
//	randomDate := start.Add(time.Duration(sec) * time.Second)
//	return randomDate
//}
//
//func CreateRandomAccount(t *testing.T) sqlc.Account {
//	args := sqlc.CreateAccountParams{
//		Owner:     util.RandomOwner(),
//		Balance:   util.RandomMoney(),
//		Email:     util.RandomEmail(),
//		Currency:  util.RandomCurrency(),
//		CreatedAt: utils.ConvertToTokyoTime(),
//	}
//
//	account, err := testStore.CreateAccount(context.Background(), args)
//	require.NoError(t, err)
//	require.NotEmpty(t, account)
//
//	require.Equal(t, args.Owner, account.Owner)
//	require.Equal(t, args.Balance, account.Balance)
//	require.Equal(t, args.Email, account.Email)
//	require.Equal(t, args.Currency, account.Currency)
//
//	require.NotZero(t, account.ID)
//	require.NotZero(t, account.CreatedAt)
//
//	return account
//}
//
//func TestCreateAccount(t *testing.T) {
//	server := newTestServer(t, testStore)
//
//	account := sqlc.Account{
//		Owner:     "John Doe",
//		Balance:   0,
//		Email:     "johndoe@gmail.com",
//		Currency:  "YEN",
//		CreatedAt: utils.ConvertToTokyoTime(),
//	}
//
//	data, err := json.Marshal(createAccountRequest{
//		Owner:     account.Owner,
//		Email:     account.Email,
//		Currency:  account.Currency,
//		CreatedAt: account.CreatedAt,
//	})
//	require.NoError(t, err)
//
//	recorder := httptest.NewRecorder()
//	request, err := http.NewRequest("POST", "/accounts", bytes.NewReader(data))
//	require.NoError(t, err)
//
//	server.router.ServeHTTP(recorder, request)
//	require.Equal(t, http.StatusOK, recorder.Code)
//
//	var createdAccount sqlc.Account
//	err = json.Unmarshal(recorder.Body.Bytes(), &createdAccount)
//	require.NoError(t, err)
//
//	require.NotZero(t, createdAccount.ID)
//	require.Equal(t, account.Owner, createdAccount.Owner)
//	require.Equal(t, account.Email, createdAccount.Email)
//	require.Equal(t, account.Currency, createdAccount.Currency)
//	require.Equal(t, account.Balance, createdAccount.Balance)
//	require.NotZero(t, createdAccount.CreatedAt)
//}
//
//func TestGetAccount(t *testing.T) {
//	server := newTestServer(t, testStore)
//
//	// Create an account to retrieve
//	account := sqlc.Account{
//		Owner:     "John Doe",
//		Balance:   0,
//		Currency:  "YEN",
//		CreatedAt: utils.ConvertToTokyoTime(),
//	}
//
//	// Create account using the store directly to ensure it exists
//	arg := sqlc.CreateAccountParams{
//		Owner:     account.Owner,
//		Currency:  account.Currency,
//		Balance:   account.Balance,
//		CreatedAt: account.CreatedAt,
//	}
//	createdAccount, err := testStore.CreateAccount(context.Background(), arg)
//	require.NoError(t, err)
//
//	recorder := httptest.NewRecorder()
//	url := fmt.Sprintf("/accounts/%d", createdAccount.ID)
//	request, err := http.NewRequest("GET", url, nil)
//	require.NoError(t, err)
//
//	server.router.ServeHTTP(recorder, request)
//	require.Equal(t, http.StatusOK, recorder.Code)
//
//	var retrievedAccount sqlc.Account
//	err = json.Unmarshal(recorder.Body.Bytes(), &retrievedAccount)
//	require.NoError(t, err)
//
//	require.Equal(t, createdAccount.ID, retrievedAccount.ID)
//	require.Equal(t, createdAccount.Owner, retrievedAccount.Owner)
//	require.Equal(t, createdAccount.Currency, retrievedAccount.Currency)
//	require.Equal(t, createdAccount.Balance, retrievedAccount.Balance)
//	require.NotZero(t, retrievedAccount.CreatedAt)
//}
//
//func TestCreateReferral(t *testing.T) {
//	account := CreateUniqueRandomAccount(t)
//
//	server := newTestServer(t, testStore)
//
//	recorder := httptest.NewRecorder()
//	url := fmt.Sprintf("/referral/account/%d", account.ID)
//
//	request, err := http.NewRequest("POST", url, nil)
//	require.NoError(t, err)
//
//	server.router.ServeHTTP(recorder, request)
//	require.Equal(t, http.StatusOK, recorder.Code)
//
//	var referralCode sqlc.ReferralCode
//	err = json.Unmarshal(recorder.Body.Bytes(), &referralCode)
//	require.NoError(t, err)
//	require.NotEmpty(t, referralCode)
//	require.Equal(t, account.ID, referralCode.ReferrerAccountID)
//	require.Equal(t, false, referralCode.IsUsed)
//	require.NotZero(t, referralCode.ReferralCode)
//}
//
//func TestUseReferralCode(t *testing.T) {
//	referrerAccount := CreateUniqueRandomAccount(t)
//	referredAccount := CreateUniqueRandomAccount(t)
//
//	referralCode := CreateUniqueRandomReferralCode(t, referrerAccount.ID)
//
//	server := newTestServer(t, testStore)
//
//	recorder := httptest.NewRecorder()
//	url := fmt.Sprintf("/referral/code/%s", referralCode.ReferralCode)
//	jsonReq := fmt.Sprintf(`{"referred_account_id": %d}`, referredAccount.ID)
//
//	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonReq)))
//	require.NoError(t, err)
//
//	server.router.ServeHTTP(recorder, request)
//	require.Equal(t, http.StatusOK, recorder.Code)
//
//	var result sqlc.ReferralCode
//	err = json.Unmarshal(recorder.Body.Bytes(), &result)
//	require.NoError(t, err)
//
//	require.Equal(t, referrerAccount.ID, result.ReferrerAccountID)
//	require.Equal(t, referrerAccount.ID, result.ReferrerAccountID)
//	require.NotZero(t, result.UsedAt.Time)
//
//}
