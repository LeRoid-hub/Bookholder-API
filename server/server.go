package server

import (
	"database/sql"
	"net/http"

	"github.com/LeRoid-hub/Bookholder-API/database"
	"github.com/gin-gonic/gin"
)

/*
	GET /Account/:id
	GET /Transaction/:TransactionID
	GET /Transactions/:AccountID/:year
	Get /Transactions/:AccountID/:year/:month
	GET /User/:UserID
	POST /NewAccount
	POST /NewTransaction
	POST /NewUser
	PUT /UpdateAccount/:AccountID
	PUT /UpdateTransaction/:TransactionID
	PUT /UpdateUser/:UserID
	DELETE /DeleteAccount/:AccountID
	DELETE /DeleteTransaction/:TransactionID
	DELETE /DeleteUser/:UserID


*/

var (
	Database *sql.DB
	Env      map[string]string
)

func Run(env map[string]string, db *database.DB) {
	dbase, err := database.New()
	if err != nil {
		panic(err)
	}
	Database = dbase

	r := gin.Default()

	// Index
	r.GET("/", welcome)

	v1 := r.Group("/v1")
	{
		//Welcome
		v1.GET("/", welcome)

		//Account
		v1.GET("/Account/:AccountID", checkAuth, getAccount)
		v1.POST("/NewAccount", checkAuth, newAccount)
		v1.PUT("/UpdateAccount", checkAuth, updateAccount)
		v1.DELETE("/DeleteAccount/:AccountID", checkAuth, deleteAccount)

		//Transaction
		v1.GET("/Transaction/:TransactionID", checkAuth, getTransaction)
		v1.GET("/Transactions/:AccountID/:year", checkAuth, getTransactions)
		v1.GET("/Transactions/:AccountID/:year/:month", checkAuth, getTransactions)
		v1.POST("/NewTransaction", checkAuth, newTransaction)
		v1.PUT("/UpdateTransaction/:TransactionID", checkAuth, updateTransaction)
		v1.DELETE("/DeleteTransaction/:TransactionID", checkAuth, deleteTransaction)

		//User
		v1.GET("/User/", checkAuth, getUserProfile)
		v1.POST("/NewUser", createUser)
		v1.POST("/AuthenticateUser", authenticateUser)
		v1.PUT("/UpdateUser/:UserID", checkAuth)
		v1.DELETE("/DeleteUser/:UserID", checkAuth)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	if port, ok := env["PORT"]; ok {
		r.Run(":" + port)
	} else {
		r.Run(":8080")
	}
}

func welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Bookholder API",
	})
}
