package api

import (
	"bank-api/db/sqlc"
	"bytes"
	"encoding/json"
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
		Currency:  "YEN",
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	recorder := httptest.NewRecorder()
	data, err := json.Marshal(createAccountRequest{
		Owner:     account.Owner,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
	})
	require.NoError(t, err)

	request, err := http.NewRequest("POST", "/accounts", bytes.NewReader(data))
	require.NoError(t, err)

	server.router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var createdAccount sqlc.Account
	err = json.Unmarshal(recorder.Body.Bytes(), &createdAccount)
	require.NoError(t, err)

	require.NotZero(t, createdAccount.ID)
	require.Equal(t, account.Owner, createdAccount.Owner)
	require.Equal(t, account.Currency, createdAccount.Currency)
	require.Equal(t, account.Balance, createdAccount.Balance)
	require.NotZero(t, createdAccount.CreatedAt)
}
