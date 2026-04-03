package types

import (
	"encoding/json"
	"time"
)

type RabbitEventType string

type RabbitEvent struct {
	Type RabbitEventType `json:"type"`
	Data json.RawMessage `json:"data"`
}

const (
	ProductCreate   RabbitEventType = "product.create"
	ProductQuantity RabbitEventType = "product.quantity"

	UserCreate RabbitEventType = "user.create"
)

type Order struct {
	Id     string `json:"id"`
	UserID string `json:"user_id" validate:"required"`

	Items []OrderItem `json:"items" validate:"required"`

	TotalAmount int64  `json:"total_amount" validate:"required"`
	Currency    string `json:"currency" validate:"required"`

	Status OrderStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderItem struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
	Price     int64  `json:"price" validate:"required"`
}

type Product struct {
	Id       string `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Price    int64  `json:"price" validate:"required"`
	Quantity int    `json:"quantity" validate:"required"`
}

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

type OrderStatus string

const (
	OrderCreated    OrderStatus = "CREATED"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderCompleted  OrderStatus = "COMPLETED"
	OrderFailed     OrderStatus = "FAILED"
)
