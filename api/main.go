package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://orders-app-mongodb:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ordersCollection = client.Database("ordersdb").Collection("orders")

	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("kafka:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}

	r := mux.NewRouter()
	r.HandleFunc("/orders", createOrder).Methods("POST")
	r.HandleFunc("/orders", listOrders).Methods("GET")

	log.Println("API server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ordersCollection.InsertOne(ctx, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg, _ := json.Marshal(order)
	kafkaWriter.WriteMessages(ctx, kafka.Message{Value: msg})

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
