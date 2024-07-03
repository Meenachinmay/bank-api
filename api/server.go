package api

import (
	"bank-api/db/sqlc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

type Server struct {
	store  *sqlc.Store
	router *gin.Engine
}

func NewServer(store *sqlc.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost", "https://*", "http://*"}, // Specify the exact origin of your Next.js app
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // Important: Must be true when credentials are included
		MaxAge:           12 * time.Hour,
	}))

	// account related routes (login, signup, fetch)
	router.POST("/accounts", server.createAccount)      // create a account (email, name, referral_code?)
	router.POST("/accounts/login", server.loginAccount) // login using email
	router.GET("/accounts/:id", server.getAccount)      // get account detail for a user
	router.GET("/accounts", server.getAccounts)

	// referral_Code feature routes
	router.POST("/referral/account/:account", server.createReferral)    // create a new referral code
	router.POST("/referral/code/:code", server.useReferralCode)         //
	router.GET("referral/calculate/:account", server.calculateInterest) // refresh extra interest for the following month
	router.GET("/referral-codes", server.getReferralCodesForAccount)    // get all the referrals code for a user

	server.router = router
	return server
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
