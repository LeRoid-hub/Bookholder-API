package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "getUser",
	})
}

func newUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "newUser",
	})
}

func updateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "updateUser",
	})
}

func deleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "deleteUser",
	})
}
