package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

type Postgres struct {
	Db *pgxpool.Pool
}

func New(dbUrl string) *Postgres {
	slog.Info("[+] Connecting with the DB")

	pool, err := pgxpool.New(context.Background(),
		dbUrl,
	)
	errhandler.FailOnError(err, "Unable to connect ot DB")

	return &Postgres{
		Db: pool,
	}
}
