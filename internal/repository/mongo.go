package repository

import (
	"context"
	"fmt"
	"order-service/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrder(ctx context.Context, id string) (*models.Order, error)
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, dbName, collectionName string) *MongoRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoRepository{
		collection: collection,
	}
}

func (r *MongoRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	order.ID = primitive.NewObjectID().Hex()
	order.CreatedAt = time.Now()
	order.Status = "created"

	_, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to create order: %v", err)
	}

	return nil
}

func (r *MongoRepository) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	return &order, nil
}
