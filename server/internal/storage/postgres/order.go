package postgres

import (
	"context"
	"fmt"

	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

// Order
// func (p *Postgres) CreateOrder(order types.Order, orderItem []types.OrderItem) error {
// 	result, err := p.Db.Exec(context.Background(),
// 		`INSERT INTO products (id, name, price, quantity) VALUES ($1, $2, $3, $4);`,
// 		product.Id, product.Name, product.Price, product.Quantity,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	if result.RowsAffected() == 0 {
// 		return fmt.Errorf("Couldn't create new product")
// 	}

// 	return nil

// }

func (p *Postgres) CreateOrder(order types.Order, orderItems []types.OrderItem) error {
	ctx := context.Background()

	tx, err := p.Db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, user_id, total_amount, currency, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.Id, order.UserID, order.TotalAmount, order.Currency, order.Status, order.CreatedAt, order.UpdatedAt,
	)

	if err != nil {
		return err
	}

	for _, item := range orderItems {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			item.OrderID, item.ProductID, item.Quantity,
		)

		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetOrder(orderId string) (types.Order, error) {
	var order types.Order

	err := p.Db.QueryRow(context.Background(),
		`SELECT id, user_id, total_amount, currency, status, created_at, updated_at FROM orders WHERE id=$1`, orderId,
	).Scan(&order.Id, &order.UserID, &order.TotalAmount, &order.Currency, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	return order, err
}

func (p *Postgres) GetOrders() ([]types.Order, error) {
	var orders []types.Order

	rows, err := p.Db.Query(context.Background(),
		`SELECT id, user_id, total_amount, currency, status, created_at, updated_at FROM orders`,
	)

	if err != nil {
		errhandler.LogOnError(err, "")
		return []types.Order{}, fmt.Errorf("something went wrong")
	}

	defer rows.Close()

	for rows.Next() {
		var order types.Order
		err = rows.Scan(&order.Id, &order.UserID, &order.TotalAmount, &order.Currency, &order.Status, &order.CreatedAt, &order.UpdatedAt)

		if err != nil {
			errhandler.LogOnError(err, "")
			return []types.Order{}, fmt.Errorf("something went wrong")
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (p *Postgres) UpdateOrderStatus(orderId string, status types.OrderStatus) error {
	_, err := p.Db.Exec(context.Background(),
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id=$2`, status, orderId,
	)

	if err != nil {
		return err
	}

	return nil
}
