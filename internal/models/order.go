package models

import (
	"time"
)

type Order struct {
	ID        string    `json:"id" bson:"_id"`
	Product   string    `json:"product" bson:"product"`
	Quantity  int       `json:"quantity" bson:"quantity"`
	Price     float64   `json:"price" bson:"price"`
	Status    string    `json:"status" bson:"status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type CreateOrderRequest struct {
	Product  string  `json:"product" binding:"required"`
	Quantity int     `json:"quantity" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
}
