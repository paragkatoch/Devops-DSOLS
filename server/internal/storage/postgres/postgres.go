package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	config "github.com/paragkatoch/Devops-DSOLS/internal"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

type Postgres struct {
	Db *pgxpool.Pool
}

func New(cfg *config.Config) *Postgres {
	slog.Info("[+] Connecting with the DB")

	pool, err := pgxpool.New(context.Background(),
		cfg.Storage_path,
	)
	errhandler.FailOnError(err, "Unable to connect ot DB")

	return &Postgres{
		Db: pool,
	}
}
