package postgres

import (
	"context"
	"fmt"

	"github.com/paragkatoch/Devops-DSOLS/types"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

// Product
func (p *Postgres) CreateProduct(product types.Product) error {
	result, err := p.Db.Exec(context.Background(),
		`INSERT INTO products (id, name, price, quantity) VALUES ($1, $2, $3, $4);`,
		product.Id, product.Name, product.Price, product.Quantity,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Couldn't create new product")
	}

	return nil

}

func (p *Postgres) GetProduct(productId string) (types.Product, error) {
	var product types.Product

	err := p.Db.QueryRow(context.Background(),
		`SELECT id, name, price, quantity FROM products WHERE id=$1`, productId,
	).Scan(&product.Id, &product.Name, &product.Price, &product.Quantity)

	return product, err
}

func (p *Postgres) GetProducts() ([]types.Product, error) {
	var products []types.Product

	rows, err := p.Db.Query(context.Background(),
		`SELECT id, name, price, quantity FROM products`,
	)

	if err != nil {
		errhandler.LogOnError(err, "")
		return []types.Product{}, fmt.Errorf("something went wrong")
	}

	defer rows.Close()

	for rows.Next() {
		var product types.Product
		err = rows.Scan(&product.Id, &product.Name, &product.Price, &product.Quantity)

		if err != nil {
			errhandler.LogOnError(err, "")
			return []types.Product{}, fmt.Errorf("something went wrong")
		}

		products = append(products, product)
	}

	return products, nil
}

func (p *Postgres) UpdateProductQuantity(productId string, quantity int) error {
	result, err := p.Db.Exec(context.Background(),
		`UPDATE products SET quantity = quantity + $1 WHERE id=$2 AND quantity + $1 >= 0`, quantity, productId,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficent stock or product not found")
	}

	return nil
}

func (p *Postgres) UpdateProductQuantityTransaction(products []types.OrderItem) error {
	ctx := context.Background()

	tx, err := p.Db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	for _, product := range products {

		result, err := tx.Exec(ctx,
			`
			UPDATE products
			SET quantity = quantity + $1
			WHERE id = $2
			AND quantity + $1 >= 0
			`,
			product.Quantity,
			product.ProductID,
		)

		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf(
				"insufficient stock for product %s",
				product.ProductID,
			)
		}
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}
