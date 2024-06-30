package api

import (
	"bank-api/db/sqlc"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type generateReferralRequest struct {
	ID int64 `uri:"account" binding:"required,min=1"`
}

type generateReferralResponse struct {
	ReferralCode string `json:"referral_code"`
}

func (server *Server) createReferral(ctx *gin.Context) {
	var req generateReferralRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := sqlc.CreateReferralCodeParams{
		ReferralCode:      "uniqueCodeA",
		ReferrerAccountID: req.ID,
		CreatedAt:         utils.ConvertToTokyoTime(),
	}

	referralCode, err := server.store.CreateReferralCode(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, generateReferralResponse{ReferralCode: referralCode.ReferralCode})
	return
}

type useReferralRequest struct {
	ReferralCode      string `uri:"code" binding:"required,min=1"`
	ReferredAccountID int64  `uri:"account" binding:"required,min=1"`
}

func (server *Server) useReferralCode(ctx *gin.Context) {
	var req useReferralRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//// fetch referral code details
	//referral_code, err := server.store.GetReferralCode(ctx, req.ReferralCode)
	//if err != nil {
	//	if err == sql.ErrNoRows {
	//		ctx.JSON(http.StatusNotFound, errorResponse(err))
	//		return
	//	}
	//	ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	//	return
	//}
	//
	//// check if code is already used
	//if referral_code.IsUsed {
	//	ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("referral code is already used")))
	//	return
	//}
	//
	//// Start a transaction

}
