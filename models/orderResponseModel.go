package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderResponse struct {
	ID         primitive.ObjectID `bson:"_id"`
	Order_Date time.Time          `json:"order_date" validate:"required"`
	Food_Items []Food             `json:"food_items" validate:"required"`
	Amount     float64            `json:"amount"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Order_id   string             `json:"order_id"`
	Table_id   *string            `json:"table_id" validate:"required"`
}