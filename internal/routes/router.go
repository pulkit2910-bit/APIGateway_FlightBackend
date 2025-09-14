package routers

import (
	handler "apigateway/internal/handlers"
	"apigateway/internal/middlewares"
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

	// Forward request to Authentication Service
	router.POST("/signup", handler.SignUpHandler)
	router.POST("/signin", handler.SignInHandler)

	protected := router.Group("/")
    protected.Use(middlewares.JWTAuthenticate(), middlewares.RateLimitCheck())
    {
        // Catch-all route for all methods (GET, POST, PUT, DELETE, PATCH, etc.)
        protected.Any("/service/*proxyPath", handler.ProxyRequestHandler)
    }

	return router
}