package setupdb

import (
	"context"
	"log/slog"

	"github.com/paragkatoch/Devops-DSOLS/internal/storage/postgres"
)

func Init(st *postgres.Postgres) {
	ctx := context.Background()

	slog.Info("[+] Creating tables if not already exists")

	st.Db.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		total_amount BIGINT NOT NULL,
		currency TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	);`)

	st.Db.Exec(ctx, `CREATE TABLE IF NOT EXISTS order_items (
		id SERIAL PRIMARY KEY,
		order_id TEXT REFERENCES orders(id) ON DELETE CASCADE,
		product_id TEXT NOT NULL,
		quantity INT NOT NULL,
		price BIGINT NOT NULL
	);`)

	st.Db.Exec(ctx, `CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		price INT NOT NULL,
		quantity INT NOT NULL
	);`)

	st.Db.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email text NOT NULL
	);`)

}
