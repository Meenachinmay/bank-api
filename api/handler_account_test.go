package api

import (
	"bank-api/db/sqlc"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, store *sqlc.Store) *Server {
	server := NewServer(store)
	return server
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
