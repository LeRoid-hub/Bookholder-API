package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/LeRoid-hub/Bookholder-API/database"
	"github.com/gin-gonic/gin"
)

func getAccount(c *gin.Context) {
	id := c.Param("AccountID")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be an integer",
		})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be an integer",
		})
		return
	}

	if idInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be greater than 0",
		})
		return
	}

	acc, err := database.GetAccount(Database, idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	if acc.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "account not found",
		})
		return
	}

	jsondata, err := json.Marshal(acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account": jsondata,
	})
}

func newAccount(c *gin.Context) {
	var acc database.Account
	err := c.BindJSON(&acc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid json",
		})
		return
	}

	if acc.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be greater than 0",
		})
		return
	}

	err = database.NewAccount(Database, acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "account created",
	})
}

func updateAccount(c *gin.Context) {
	var acc database.Account
	err := c.BindJSON(&acc)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid json",
		})
		return
	}

	acc.ID = uint(acc.ID)

	if acc.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be greater than 0",
		})
		return
	}

	err = database.UpdateAccount(Database, acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account updated",
	})
}

func deleteAccount(c *gin.Context) {
	id := c.Param("AccountID")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be an integer",
		})
		return
	}

	if idInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id; must be greater than 0",
		})
		return
	}

	err = database.DeleteAccount(Database, idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account deleted",
	})
}
