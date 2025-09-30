package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	file, err := os.OpenFile("/var/log/orders.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"kafka:9092"},
		Topic:     "orders",
		GroupID:   "logger-group",
		Partition: 0,
	})

	ctx := context.Background()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("error reading message: %v", err)
			time.Sleep(time.Second)
			continue
		}
		file.WriteString(string(m.Value) + "\n")
		file.Sync()
		log.Printf("logged order: %s", string(m.Value))
	}
}
