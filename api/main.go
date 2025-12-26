package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Order struct {
	ID       string  `json:"id" bson:"_id,omitempty"`
	Customer string  `json:"customer" bson:"customer"`
	Item     string  `json:"item" bson:"item"`
	Amount   float64 `json:"amount" bson:"amount"`
}

var (
	ordersCollection *mongo.Collection
	kafkaWriter      *kafka.Writer
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://orders-app-mongodb:27017"))

	// СТАЛО (правильный вариант):
	mongoURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:27017/%s?authSource=admin",
		os.Getenv("MONGO_USERNAME"),
		os.Getenv("MONGO_PASSWORD"),
		os.Getenv("MONGO_HOST"), // Используем отдельную переменную для хоста
		os.Getenv("MONGO_DATABASE"),
	)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	//
	if err != nil {
		log.Fatal(err)
	}
	ordersCollection = client.Database("ordersdb").Collection("orders")

	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("orders-app-kafka:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}

	r := mux.NewRouter()
	r.HandleFunc("/orders", createOrder).Methods("POST")
	r.HandleFunc("/orders", listOrders).Methods("GET")

	log.Println("API server running on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Сохраняем в MongoDB
	result, err := ordersCollection.InsertOne(ctx, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем ID из MongoDB
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		order.ID = oid.Hex()
	}

	// Отправляем в Kafka С ОБРАБОТКОЙ ОШИБОК
	msg, _ := json.Marshal(order)
	kafkaMsg := kafka.Message{
		Key:   []byte(order.ID),
		Value: msg,
		Time:  time.Now(),
	}

	// Логируем попытку отправки
	log.Printf("Sending to Kafka: %s", string(msg))

	// Пытаемся отправить с таймаутом
	kafkaCtx, kafkaCancel := context.WithTimeout(ctx, 3*time.Second)
	defer kafkaCancel()

	err = kafkaWriter.WriteMessages(kafkaCtx, kafkaMsg)
	if err != nil {
		// Логируем ошибку, но не прерываем запрос
		log.Printf("WARNING: Failed to send to Kafka: %v", err)
		// Можно продолжать, так как заказ уже сохранён в MongoDB
	} else {
		log.Printf("SUCCESS: Sent to Kafka: %s", order.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func listOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := ordersCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var orders []Order
	for cur.Next(ctx) {
		var order Order
		cur.Decode(&order)
		orders = append(orders, order)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
