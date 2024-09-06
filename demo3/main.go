package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var secretKey = []byte("secret")

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(2 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("Token is invalid")
	}

	return nil
}

func main() {
	r := gin.Default()

	r.POST("/admin/login", func(c *gin.Context) {
		var reqBody LoginRequest
		if err := c.BindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if reqBody.Username != "username" || reqBody.Password != "password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		token, _ := generateToken(reqBody.Username)
		resp := LoginResponse{Token: token}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/resource/listall", func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Token validation error: %v", err)})
			return
		}

		data := []Resource{
			{Name: "Resource 1"},
			{Name: "Resource 2"},
			{Name: "Resource 3"},
		}

		c.JSON(http.StatusOK, data)
	})

	fmt.Println("Starting server on :8081")
	r.Run(":8081")
}

type Resource struct {
	Name string `json:"name"`
}
