package storage

import "github.com/paragkatoch/Devops-DSOLS/types"

type Storage interface {
	// Order
	// CreateOrder(order types.Order) (string, error)
	// GetOrderStatus(orderId string) (types.OrderStatus, error)
	// GetOrder(orderId string) (types.Order, error)

	// product
	CreateProduct(product types.Product) error
	GetProduct(productId string) (types.Product, error)
	GetProducts() ([]types.Product, error)
	UpdateProductQuantity(productId string, quantity int) error // add/remove product from stock

	// user
	CreateUser(user types.User) error
	GetUser(userId string) (types.User, error)
	GetUsers() ([]types.User, error)
	GetUserOrders(userid string) ([]types.Order, error)
}
