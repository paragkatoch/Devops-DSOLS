package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	config "github.com/paragkatoch/Devops-DSOLS/internal"
	errhandler "github.com/paragkatoch/Devops-DSOLS/util/errHandler"
)

type Postgres struct {
	Db *pgx.Conn
}

func New(cfg *config.Config) *Postgres {
	slog.Info("[+] Connecting with the DB")

	conn, err := pgx.Connect(context.Background(),
		cfg.Storage_path,
	)
	errhandler.FailOnError(err, "Unable to connect ot DB")

	return &Postgres{
		Db: conn,
	}
}
