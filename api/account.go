package api

import (
	"bank-api/db/sqlc"
	"database/sql"
	"github.com/Meenachinmay/microservice-shared/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type createAccountRequest struct {
	Owner     string    `json:"owner" binding:"required"`
	Currency  string    `json:"currency" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := sqlc.CreateAccountParams{
		Owner:     req.Owner,
		Currency:  req.Currency,
		Balance:   0,
		CreatedAt: utils.ConvertToTokyoTime(),
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required, min=1"`
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
