package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

var rateLimiterServiceURL = "http://localhost:8080"
var authServiceURL = "http://localhost:3001"

func JWTAuthenticate(httpClient *http.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.Request.Header.Get("x-access-token")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}
         
        url := fmt.Sprintf("%s/isAuthenticated", authServiceURL)
		bodyBytes, err := io.ReadAll(c.Request.Body)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err, "message": "failed to read request body"})
            c.Abort()
            return
        }

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	    if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to create request to authentication service"})
			c.Abort()
	    }

	    req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-access-token", token)
	    resp, err := httpClient.Do(req)
		if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "Authentication service error"})
			c.Abort()
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
            c.JSON(resp.StatusCode, gin.H{"error": "authentication service error"})
            c.Abort()
            return
        }

        var authResp struct {
            Data    interface{} `json:"data"`
            Success bool        `json:"success"`
            Message string      `json:"message"`
            Err     interface{} `json:"err"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "invalid response from auth service"})
            c.Abort()
            return
        }

        if !authResp.Success {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed", "message": authResp.Message})
            c.Abort()
            return
        }

		userIdStr := fmt.Sprintf("%v", authResp.Data)
		if userIdStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in auth response"})
            c.Abort()
            return
        }

        c.Set("userId", userIdStr)
		c.Next()
    }
}

func RateLimitCheck(httpClient *http.Client) gin.HandlerFunc {
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
        url := fmt.Sprintf("%s/check?userId=%v&api=%s", rateLimiterServiceURL, userId, api)

		req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	    if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to create request to rate limiter service"})
			c.Abort()
	    }

	    req.Header.Set("Content-Type", "application/json")
	    resp, err := httpClient.Do(req)
		if err != nil {
            c.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "Rate limiter service error"})
			c.Abort()
		}

		defer resp.Body.Close()

		var rateLimiterResp struct {
			Allowed bool `json:"isAllowed"`
			Result interface{} `json:"result"`
		}
		
		err = json.NewDecoder(resp.Body).Decode(&rateLimiterResp)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err, "message": "invalid response from rate limiter service"})
			c.Abort()
			return
		}

		if !rateLimiterResp.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}