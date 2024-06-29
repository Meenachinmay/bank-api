package api

import (
	"bank-api/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *sqlc.Store
	router *gin.Engine
}

func NewServer(store *sqlc.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.getAccounts)

	server.router = router
	return server
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
