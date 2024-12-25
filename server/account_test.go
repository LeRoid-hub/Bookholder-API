package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetAccountStringAsAccountID(t *testing.T) {
	r := gin.Default()
	r.GET("/Account/:AccountID", getAccount)

	req, _ := http.NewRequest("GET", "/Account/a", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetAccountInvalidID(t *testing.T) {
	r := gin.Default()
	r.GET("/Account/:AccountID", getAccount)

	req, _ := http.NewRequest("GET", "/Account/0", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetAccountNegativID(t *testing.T) {
	r := gin.Default()
	r.GET("/Account/:AccountID", getAccount)

	req, _ := http.NewRequest("GET", "/Account/-1", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TODO: TestGetAccountValidID
// TODO: TestGetAccountInternalError

func TestNewAccountNegativeID(t *testing.T) {
	r := gin.Default()
	r.POST("/NewAccount", newAccount)

	accountJson := `{
		"ID": -1,
		"Name": "Test Account",
		"Kind": "1000.00"
	}`

	req, _ := http.NewRequest("POST", "/NewAccount", strings.NewReader(accountJson))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestNewAccountWrongFormat(t *testing.T) {
	r := gin.Default()
	r.POST("/NewAccount", newAccount)

	accountJson := `{
		"ID": 1,
		"Stuff": "Test Account",
	}`

	req, _ := http.NewRequest("POST", "/NewAccount", strings.NewReader(accountJson))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TODO: TestNewAccountInternalError
// TODO: TestNewAccountValid

func TestUpdateAccountNegativeID(t *testing.T) {
	r := gin.Default()
	r.PUT("/UpdateAccount", updateAccount)

	accountJson := `{
		"ID": -1,
		"Name": "Test Account",
		"Kind": "1000.00"
	}`

	req, _ := http.NewRequest("PUT", "/UpdateAccount", strings.NewReader(accountJson))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestUpdateAccountWrongFormat(t *testing.T) {
	r := gin.Default()
	r.PUT("/UpdateAccount", updateAccount)

	accountJson := `{
		"ID": 1,
		"Stuff": "Test Account",
	}`

	req, _ := http.NewRequest("PUT", "/UpdateAccount", strings.NewReader(accountJson))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TODO: TestUpdateAccountInternalError
// TODO: TestUpdateAccountValid

func TestDeleteAccountNegativeID(t *testing.T) {
	r := gin.Default()
	r.DELETE("/DeleteAccount/:AccountID", deleteAccount)

	req, _ := http.NewRequest("DELETE", "/DeleteAccount/-1", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestDeleteAccountString(t *testing.T) {
	r := gin.Default()
	r.DELETE("/DeleteAccount/:AccountID", deleteAccount)

	req, _ := http.NewRequest("DELETE", "/DeleteAccount/a", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	fmt.Println(resp.Body.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TODO: TestDeleteAccountInternalError
// TODO: TestDeleteAccountValid
