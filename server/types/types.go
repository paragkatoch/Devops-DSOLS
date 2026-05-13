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
	ProductCreate          RabbitEventType = "product.create"
	ProductQuantity        RabbitEventType = "product.quantity"
	ProductQuantityReserve RabbitEventType = "product.quantity.reserve"
	UserCreate             RabbitEventType = "user.create"
	OrderCreate            RabbitEventType = "order.create"
	OrderStatusUpdate      RabbitEventType = "order.status.update"
)

type OrderHttp struct {
	UserID      string `json:"user_id" validate:"required"`
	TotalAmount int64  `json:"total_amount" validate:"required"`
	Currency    string `json:"currency" validate:"required"`

	Items []OrderItemHttp `json:"items" validate:"required"`
}

type Order struct {
	Id     string `json:"id"`
	UserID string `json:"user_id" validate:"required"`

	TotalAmount int64  `json:"total_amount" validate:"required"`
	Currency    string `json:"currency" validate:"required"`

	Status OrderStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderItemHttp struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
}

type OrderItem struct {
	OrderID   string `json:"order_id" validate:"required"`
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
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

type OrderCreateEvent struct {
	Order Order       `json:"order"`
	Items []OrderItem `json:"items"`
}

type ProductReserveEvent struct {
	OrderID string      `json:"order_id"`
	Items   []OrderItem `json:"items"`
}
