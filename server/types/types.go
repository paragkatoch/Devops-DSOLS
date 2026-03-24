package types

import "time"

type Order struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id" validate:"required"`

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

type OrderStatus string

const (
	OrderCreated    OrderStatus = "CREATED"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderCompleted  OrderStatus = "COMPLETED"
	OrderFailed     OrderStatus = "FAILED"
)
