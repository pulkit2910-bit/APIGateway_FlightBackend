package middlewares

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("khjfkasdhfkaioqwerfmckjshfkajwkef")
var rateLimiterServiceURL = "http://localhost:8080"

func JWTAuthenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.Request.Header.Get("x-access-token")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}
         
        verifiedToken, err := verifyToken(token)
		if err != nil {
			fmt.Printf("Token verification failed: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		fmt.Printf("Token verified successfully: %s\n", token)
		c.Set("userId", verifiedToken.Claims.(jwt.MapClaims)["userId"])
		c.Next()
    }
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	var token *jwt.Token
	claims := jwt.MapClaims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return secretKey, nil
    })
    if err != nil {
        return token, err
    }
  
    if !token.Valid {
        return token, fmt.Errorf("invalid token")
    }
  
    return token, nil
}

func RateLimitCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        userId, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			c.Abort()
			return
		}

		if userId == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID"})
			c.Abort()
			return
		}

		api := c.Param("api")
        url := fmt.Sprintf("%s/check?%v&api=%s", rateLimiterServiceURL, userId, api)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	    if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to rate limiter service"})
			c.Abort()
	    }

		httpClient := &http.Client{Timeout: 10 * time.Second}

	    req.Header.Set("Content-Type", "application/json")
	    resp, err := httpClient.Do(req)
		if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": "Rate limiter service error"})
			c.Abort()
		}

		defer resp.Body.Close()
		
		fmt.Printf("Rate limiter response: %+v\n", resp)
	}
}