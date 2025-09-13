package handler

import service "apigateway/internal/services"

type Handler struct {
	service service.Service
}

type HandlerInterface interface {
	// Define handler methods here
}

func NewHandler(service service.ServiceInterface) HandlerInterface {
	return &Handler{
		service: service,
	}
}
