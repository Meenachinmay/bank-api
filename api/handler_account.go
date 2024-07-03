package api

import (
	"bank-api/db/sqlc"
	"database/sql"
	"errors"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type createAccountRequest struct {
	Owner        string    `json:"owner" binding:"required"`
	Currency     string    `json:"currency" binding:"required,oneof=YEN EUR USD"`
	Email        string    `json:"email" binding:"required,email"`
	ReferralCode string    `json:"referral_code"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	var newBalance int64
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.ReferralCode != "" {
		// TODO: mark code as used (this will referred_account extra interest of 1% for the following month
		_, err := server.markCode(ctx, req.ReferralCode)
		if err != nil {
			if err.Error() == "referral code is already used" {
				ctx.JSON(http.StatusConflict, gin.H{"error": "This referral code is already used."})
				return
			}
			if err.Error() == "invalid referral code" {
				ctx.JSON(http.StatusConflict, gin.H{"error": "Referral code is invalid."})
				return
			}
			if err.Error() == "error with referral code" {
				ctx.JSON(http.StatusConflict, gin.H{"error": "Referral code is invalid."})
				return
			}
			// TODO: log as CRUCIAL error as referralCode couldn't be used
		}
		// TODO: update balance as gift for new user joined with code
		newBalance = 1000
		// TODO: send email to referrer_account as notification so referrer cannot know about it.
	}

	arg := sqlc.CreateAccountParams{
		Owner:     req.Owner,
		Currency:  req.Currency,
		Email:     req.Email,
		Balance:   newBalance,
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

func (server *Server) markCode(ctx *gin.Context, code string) (sqlc.ReferralCode, error) {
	// TODO: fetch the code to check if used or not
	referralCodeToCheck, err := server.store.GetReferralCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.ReferralCode{}, errors.New("invalid referral code")
		}
		return sqlc.ReferralCode{}, errors.New("error with referral code")
	}

	if referralCodeToCheck.IsUsed {
		return referralCodeToCheck, errors.New("referral code is already used")
	}

	args := sqlc.MarkReferralCodeUsedParams{
		ReferralCode: code,
		UsedAt:       sql.NullTime{Time: utils.ConvertToTokyoTime(), Valid: true},
	}
	referralCode, err := server.store.MarkReferralCodeUsed(ctx, args)
	if err != nil {
		return sqlc.ReferralCode{}, err
	}

	return referralCode, nil
}

type loginAccountRequest struct {
	Email string `json:"email" binding:"required"`
}

func (server *Server) loginAccount(ctx *gin.Context) {
	var req loginAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if email exists
	account, err := server.store.GetAccountWithEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("email does not exist")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) getAccounts(ctx *gin.Context) {
	var req getAccountsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := sqlc.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
