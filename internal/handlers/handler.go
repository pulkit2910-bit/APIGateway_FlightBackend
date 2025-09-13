package handler

import (
	service "apigateway/internal/services"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.ServiceInterface
}

type HandlerInterface interface {
	SignUpHandler(ctx *gin.Context)
}

func (h *Handler) SignUpHandler(ctx *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding request body: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		fmt.Printf("Invalid request: email and password are required\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	// Marshal request to JSON
	payload, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}

	resp, err := h.service.SignUp(payload)
	if err != nil {
		fmt.Printf("Error calling auth service: %+v\n", err)
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "auth service error"})
		return
	}
	defer resp.Body.Close()

	ctx.Status(resp.StatusCode)
	ctx.Writer.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	if _, copyErr := io.Copy(ctx.Writer, resp.Body); copyErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read auth service response"})
	}
}

func NewHandler(service service.ServiceInterface) HandlerInterface {
	return &Handler{
		service: service,
	}
}
