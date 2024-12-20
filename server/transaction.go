package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getTransaction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "getTransaction",
	})
}

func getTransactions(c *gin.Context) {
	year := c.Param("year")
	month := c.Param("month")

	message := "getTransactions " + year + " " + month
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

func newTransaction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "newTransaction",
	})
}

func updateTransaction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "updateTransaction",
	})
}

func deleteTransaction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "deleteTransaction",
	})
}
