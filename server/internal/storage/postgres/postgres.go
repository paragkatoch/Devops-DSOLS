package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	setupdb "github.com/paragkatoch/Devops-DSOLS/util/DB"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

type Postgres struct {
	Db *pgx.Conn
}

func New() *Postgres {
	slog.Info("[+] Connecting with the DB")

	conn, err := pgx.Connect(context.Background(),
		"postgres://admin:admin@localhost:5432/retail?sslmode=disable",
	)
	errhandler.FailOnError(err, "Unable to connect ot DB")
	setupdb.Init(conn)

	return &Postgres{
		Db: conn,
	}
}
