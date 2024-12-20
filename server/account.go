package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "getAccount",
	})
}

func newAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "newAccount",
	})
}

func updateAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "updateAccount",
	})
}

func deleteAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "deleteAccount",
	})
}
