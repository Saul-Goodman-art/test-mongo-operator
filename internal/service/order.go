package service

import (
	"context"
	"fmt"
	"order-service/internal/kafka"
	"order-service/internal/models"
	"order-service/internal/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error)
	GetOrder(ctx context.Context, id string) (*models.Order, error)
}

type orderService struct {
	repo     repository.OrderRepository
	producer kafka.Producer
}

func NewOrderService(repo repository.OrderRepository, producer kafka.Producer) OrderService {
	return &orderService{
		repo:     repo,
		producer: producer,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	order := &models.Order{
		Product:  req.Product,
		Quantity: req.Quantity,
		Price:    req.Price,
		Status:   "created",
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order in repository: %w", err)
	}

	if err := s.producer.SendOrderEvent(ctx, order); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		fmt.Printf("Failed to send order event to Kafka: %v\n", err)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	return s.repo.GetOrder(ctx, id)
}
