package main

import (
	"flag"
	"log/slog"

	"github.com/paragkatoch/Devops-DSOLS/controllers"
	config "github.com/paragkatoch/Devops-DSOLS/internal"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage/postgres"
	"github.com/paragkatoch/Devops-DSOLS/services"
	setupdb "github.com/paragkatoch/Devops-DSOLS/util/DB"
)

var registry = map[string]map[string]func(storage.Storage, *config.Config){
	"service": {
		"order":   services.OrderService,
		"user":    services.UserService,
		"product": services.ProductService,
	},
	"controller": {
		"order":   controllers.OrderController,
		"user":    controllers.UserController,
		"product": controllers.ProductController,
	},
}

func main() {

	// parse flags
	slog.Info("checking server type and component")
	componentType := flag.String("type", "", "type of the server component (service/controller)")
	component := flag.String("component", "", "server component (order, user)")
	// load config
	cfg := config.MustLoad()
	flag.Parse()

	st := postgres.New(cfg)

	if *componentType == "init" {
		setupdb.Init(st.Db)
		return
	}

	if compType, ok := registry[*componentType]; ok {
		if fn, ok := compType[*component]; ok {
			fn(st, cfg)
		} else {
			slog.Error("Invalid component")
		}
	} else {
		slog.Error("Invalid type")
	}
}
