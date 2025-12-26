package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Printf("Logger service starting...")

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "orders-app-kafka:9092"
	}
	log.Printf("Using Kafka broker: %s", kafkaBroker)

	file, err := os.OpenFile("/var/log/orders.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	// ИСПРАВЛЕННАЯ КОНФИГУРАЦИЯ KAFKA READER
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		Topic:          "orders",
		GroupID:        "logger-group",
		MinBytes:       10e3,             // 10KB
		MaxBytes:       10e6,             // 10MB
		MaxWait:        3 * time.Second,  // Ждать сообщения до 3 секунд
		CommitInterval: time.Second,      // Частота коммитов
		StartOffset:    kafka.LastOffset, // Читать с последнего сообщения
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second, // Таймаут подключения
			DualStack: true,
		},
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 1 * time.Second,
	})
	defer r.Close()

	log.Printf("Kafka reader created with improved configuration")

	ctx := context.Background()
	messageCount := 0

	for {
		// Читаем сообщение с контекстом и таймаутом
		readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		m, err := r.ReadMessage(readCtx)
		cancel()

		if err != nil {
			// Проверяем, это таймаут или реальная ошибка
			if err == context.DeadlineExceeded {
				// Просто нет новых сообщений - это нормально
				log.Printf("No new messages in 30 seconds, continuing...")
				continue
			}

			log.Printf("ERROR reading from Kafka: %v (type: %T)", err, err)
			time.Sleep(2 * time.Second) // Увеличиваем задержку
			continue
		}

		messageCount++
		message := string(m.Value)

		// Пишем в файл
		if _, err := file.WriteString(fmt.Sprintf("[%s] %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			message)); err != nil {
			log.Printf("ERROR writing to file: %v", err)
		}
		file.Sync()

		// Пишем в лог
		log.Printf("SUCCESS: Received message #%d (offset %d): %.100s...",
			messageCount, m.Offset, message)
	}
}
