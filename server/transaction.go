package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/LeRoid-hub/Bookholder-API/database"
	"github.com/gin-gonic/gin"
)

func getTransaction(c *gin.Context) {
	id := c.Param("TransactionID")

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

	transaction, err := database.GetTransaction(Database, idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	if transaction.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "transaction not found",
		})
		return
	}

	jasondata, err := json.Marshal(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction": jasondata,
	})
}

func getTransactions(c *gin.Context) {
	year := c.Param("year")
	month := c.Param("month")

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid year; must be an integer",
		})
		return
	}

	if yearInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid year; must be greater than 0",
		})
		return
	}

	monthInt, err := strconv.Atoi(month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid month; must be an integer",
		})
		return
	}

	if monthInt < 1 || monthInt > 12 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid month; must be between 1 and 12",
		})
		return
	}

	account := c.Param("AccountID")
	accountInt, err := strconv.Atoi(account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid account; must be an integer",
		})
		return
	}

	if accountInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid account; must be greater than 0",
		})
		return
	}

	transactions, err := database.GetTransactions(Database, accountInt, yearInt, monthInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	jasondata, err := json.Marshal(transactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": jasondata,
	})
}

func newTransaction(c *gin.Context) {
	var transaction database.Transaction
	err := c.BindJSON(&transaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid json",
		})
		return
	}

	if transaction.OffsetAccount == transaction.Account {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid offset account; must be different from account",
		})
		return
	}

	err = database.NewTransaction(Database, transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "transaction created",
	})
}

func updateTransaction(c *gin.Context) {
	var transaction database.Transaction
	err := c.BindJSON(&transaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid json",
		})
		return
	}

	if transaction.OffsetAccount == transaction.Account {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid offset account; must be different from account",
		})
		return
	}

	err = database.UpdateTransaction(Database, transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "transaction updated",
	})
}

func deleteTransaction(c *gin.Context) {
	id := c.Param("TransactionID")

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

	err = database.DeleteTransaction(Database, idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "transaction deleted",
	})
}
