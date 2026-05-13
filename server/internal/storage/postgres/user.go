package postgres

import (
	"context"
	"fmt"

	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

func (p *Postgres) CreateUser(user types.User) error {
	result, err := p.Db.Exec(context.Background(),
		`INSERT INTO users (id, email) VALUES ($1, $2);`,
		user.Id, user.Email,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Couldn't create new user")
	}

	return nil
}

func (p *Postgres) GetUser(userId string) (types.User, error) {
	var user types.User

	err := p.Db.QueryRow(context.Background(),
		`SELECT id, email FROM users WHERE id=$1`,
		userId,
	).Scan(&user.Id, &user.Email)

	return user, err
}

func (p *Postgres) GetUsers() ([]types.User, error) {
	var users []types.User

	rows, err := p.Db.Query(context.Background(),
		`SELECT id, email FROM users`,
	)

	if err != nil {
		errhandler.LogOnError(err, "")
		return []types.User{}, fmt.Errorf("something went wrong")
	}

	defer rows.Close()

	for rows.Next() {
		var user types.User
		err := rows.Scan(&user.Id, &user.Email)

		if err != nil {
			errhandler.LogOnError(err, "")
			return []types.User{}, fmt.Errorf("something went wrong")
		}
		users = append(users, user)
	}

	return users, nil
}

// func (p *Postgres) GetUserOrders(userid string) ([]types.Order, error) {
// 	var orders []types.Order

// 	rows, err := p.Db.Query(context.Background(),
// 		`SELECT id, user_id, total_amount, currency, status FROM orders WHERE user_id=$1`,
// 		userid,
// 	)

// 	if err != nil {
// 		errhandler.LogOnError(err, "")
// 		return []types.Order{}, fmt.Errorf("something went wrong")
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		var order types.Order
// 		err := rows.Scan(&order.Id, &order.UserID, &order.Items, &order.TotalAmount, &order.Currency)

// 		if err != nil {
// 			errhandler.LogOnError(err, "")
// 			return []types.Order{}, fmt.Errorf("something went wrong")
// 		}
// 		orders = append(orders, order)
// 	}

// 	return orders, nil
// }
