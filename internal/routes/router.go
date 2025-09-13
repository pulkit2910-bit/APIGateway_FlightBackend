package router

import (
	handler "apigateway/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(handler handler.HandlerInterface) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Service is running",
		})
	})

	return router
}