package server

import (
	handler "apigateway/internal/handlers"
	router "apigateway/internal/routes"
	service "apigateway/internal/services"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start() {
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:3001"
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}

	service := service.NewService(authServiceURL, httpClient)
	handler := handler.NewHandler(service)
	r := router.SetupRouter(handler)

	server := &http.Server{
		Addr: ":9090",
		Handler: r,
	}

	go func() {
		log.Println("Starting server on :9090")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
