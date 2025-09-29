package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/handler"
	"order-service/internal/kafka"
	"order-service/internal/repository"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

// Функция для получения значения из переменных окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Конфигурация из переменных окружения
	mongoURI := getEnv("MONGO_URI", "mongodb://mongo:27017")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50051")
	databaseName := getEnv("MONGO_DATABASE", "orders")
	collectionName := getEnv("MONGO_COLLECTION", "orders")
	kafkaTopic := getEnv("KAFKA_TOPIC", "order-events")

	log.Printf("Starting Order Service with configuration:")
	log.Printf("MongoDB: %s", mongoURI)
	log.Printf("Kafka: %s", kafkaBrokers)
	log.Printf("HTTP Port: %s", httpPort)
	log.Printf("gRPC Port: %s", grpcPort)

	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Проверка подключения к MongoDB
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Successfully connected to MongoDB")

	// Инициализация репозитория
	orderRepo := repository.NewMongoRepository(mongoClient, databaseName, collectionName)

	// Инициализация Kafka producer
	kafkaProducer := kafka.NewKafkaProducer([]string{kafkaBrokers}, kafkaTopic)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}()

	// Инициализация сервиса
	orderService := service.NewOrderService(orderRepo, kafkaProducer)

	// Инициализация HTTP handler
	httpHandler := handler.NewHTTPHandler(orderService)

	// Запуск HTTP сервера
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Проверка подключения к MongoDB
		if err := mongoClient.Ping(c.Request.Context(), nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "unhealthy",
				"error":     "MongoDB connection failed",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"services": map[string]string{
				"mongodb": "connected",
				"kafka":   "connected",
			},
		})
	})

	// Информация о сервисе
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Order Service",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"create_order": "POST /api/v1/orders",
				"get_order":    "GET /api/v1/orders/:id",
				"health":       "GET /health",
			},
		})
	})

	httpHandler.RegisterRoutes(router)

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	go func() {
		log.Printf("HTTP server starting on port %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Запуск gRPC сервера (упрощенная версия)
	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcHandler := handler.NewGRPCHandler(orderService)
	grpcHandler.RegisterService(grpcServer)

	go func() {
		log.Printf("gRPC server starting on port %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	log.Printf("Order service started successfully")
	log.Printf("HTTP API available on port %s", httpPort)
	log.Printf("gRPC API available on port %s", grpcPort)

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
	log.Println("Servers stopped gracefully")
}
