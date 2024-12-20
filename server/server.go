package server

import (
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

func Run(env map[string]string, db *database.DB) {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		//Account
		v1.GET("/Account/:AccountID", getAccount)
		v1.POST("/NewAccount", newAccount)
		v1.PUT("/UpdateAccount/:AccountID", updateAccount)
		v1.DELETE("/DeleteAccount/:AccountID", deleteAccount)

		//Transaction
		v1.GET("/Transaction/:TransactionID", getTransaction)
		v1.GET("/Transactions/:AccountID/:year", getTransactions)
		v1.GET("/Transactions/:AccountID/:year/:month", getTransactions)
		v1.POST("/NewTransaction", newTransaction)
		v1.PUT("/UpdateTransaction/:TransactionID", updateTransaction)
		v1.DELETE("/DeleteTransaction/:TransactionID", deleteTransaction)

		//User
		v1.GET("/User/:UserID", getUser)
		v1.POST("/NewUser", newUser)
		v1.PUT("/UpdateUser/:UserID", updateUser)
		v1.DELETE("/DeleteUser/:UserID", deleteUser)
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
