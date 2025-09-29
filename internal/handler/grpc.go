package handler

import (
	"order-service/internal/service"

	"google.golang.org/grpc"
)

// GRPCHandler - упрощенная версия без полноценной gRPC реализации
type GRPCHandler struct {
	orderService service.OrderService
}

func NewGRPCHandler(orderService service.OrderService) *GRPCHandler {
	return &GRPCHandler{
		orderService: orderService,
	}
}

func (h *GRPCHandler) RegisterService(server *grpc.Server) {
	// Пустая реализация - в данном примере фокусируемся на HTTP API
	// Для полноценной gRPC реализации нужно создать .proto файлы и сгенерировать код
}
