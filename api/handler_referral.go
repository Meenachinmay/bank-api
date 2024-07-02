package api

import (
	"bank-api/db/sqlc"
	"bank-api/util"
	"database/sql"
	"errors"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type generateReferralRequest struct {
	ID int64 `uri:"account" binding:"required,min=1"`
}

type generateReferralResponse struct {
	ReferralCode sqlc.ReferralCode `json:"referral_code"`
}

func (server *Server) createReferral(ctx *gin.Context) {
	var req generateReferralRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//TODO: Check for any un-used code by this user.
	hasUnUsedCode, err := server.store.HasUnUsedCodeForReferrerAccount(ctx, req.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if hasUnUsedCode {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "cannot create a new code, before using it",
		})
		return
	}

	referral := util.RandomUUID()
	arg := sqlc.CreateReferralCodeParams{
		ReferralCode:      referral,
		ReferrerAccountID: req.ID,
		CreatedAt:         utils.ConvertToTokyoTime(),
	}

	referralCode, err := server.store.CreateReferralCode(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, referralCode)
	return
}

type useReferralRequestCode struct {
	ReferralCode string `uri:"code" binding:"required,min=1"`
}
type useReferralRequestAccountID struct {
	ReferredAccount int64 `json:"referred_account_id" binding:"required,min=1"`
}

func (server *Server) useReferralCode(ctx *gin.Context) {
	var req useReferralRequestCode
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var jsonReq useReferralRequestAccountID
	if err := ctx.ShouldBindJSON(&jsonReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	args := sqlc.MarkReferralCodeUsedParams{
		ReferralCode: req.ReferralCode,
		UsedAt:       sql.NullTime{Time: utils.ConvertToTokyoTime(), Valid: true},
	}

	// find the code in table and mark it true
	referralCode, err := server.store.MarkReferralCodeUsed(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create record in referral codes history table
	_, err = server.store.CreateReferralHistory(ctx, sqlc.CreateReferralHistoryParams{
		ReferrerAccountID: referralCode.ReferrerAccountID,
		ReferredAccountID: jsonReq.ReferredAccount,
		ReferralCodeID:    referralCode.ID,
		ReferralDate:      referralCode.CreatedAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	referrerAccount, err := server.store.GetAccount(ctx, referralCode.ReferrerAccountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	currentExtraInterest := 0.0
	if referrerAccount.ExtraInterest.Valid {
		currentExtraInterest = referrerAccount.ExtraInterest.Float64
	}

	expectedExtraInterest := currentExtraInterest + 1.0
	if expectedExtraInterest > 10.0 {
		expectedExtraInterest = 10.0
	}

	ctx.JSON(http.StatusOK, referralCode)

}

type calculateReferralRequest struct {
	ReferrerAccountID int64 `uri:"account" binding:"required,min=1"`
}

// CalculateInterest - This method runs on 21st of every month to calculate extra interest for
// following month
func (server *Server) calculateInterest(ctx *gin.Context) {
	var req calculateReferralRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	args := sqlc.UseReferralCodeTxParams{
		ReferrerAccountID: req.ReferrerAccountID,
	}

	txResult, err := server.store.UseReferralCodeTx(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, txResult.ReferrerAccountUpdate)

}

func (server *Server) getReferralCodesForAccount(ctx *gin.Context) {
	accountIDStr := ctx.Query("account")
	if accountIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "query params are missing",
		})
		return
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errors.New("invalid account id"))
		return
	}

	referralCodes, err := server.store.GetReferralCodesForReferrerAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "no referral codes are available",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, referralCodes)
}
