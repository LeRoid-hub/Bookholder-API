package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LeRoid-hub/Bookholder-API/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func createUser(c *gin.Context) {
	var authInput AuthInput

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user database.User
	user, err := database.GetUserByName(Database, authInput.Username)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if user.ID != 0 {
		c.JSON(400, gin.H{"error": "User already exists"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(authInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user = database.User{
		Name:     authInput.Username,
		Password: string(passwordHash),
	}

	err = database.NewUser(Database, user)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User created"})
}

func authenticateUser(c *gin.Context) {
	var authInput AuthInput

	if err := c.ShouldBindJSON(&authInput); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var userFound database.User
	userFound, err := database.GetUserByName(Database, authInput.Username)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if userFound.ID == 0 {
		c.JSON(400, gin.H{"error": "User not found"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFound.Password), []byte(authInput.Password))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid password"})
		return
	}

	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userFound.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := generateToken.SignedString([]byte(Env["SECRET"]))
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"token": tokenString})
}

func getUserProfile(c *gin.Context) {
	user, _ := c.Get("currentUser")
	c.JSON(200, gin.H{"user": user})
}

func checkAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	authToken := strings.Split(authHeader, " ")
	if len(authToken) != 2 || authToken[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tokenString := authToken[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(Env["SECRET"]), nil
	})

	if err != nil || token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, err := database.GetUser(Database, int(claims["id"].(float64)))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("currentUser", user)

	c.Next()
}
