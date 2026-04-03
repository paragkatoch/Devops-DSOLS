package main

import (
	"flag"
	"log/slog"

	"github.com/paragkatoch/Devops-DSOLS/controllers"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
	"github.com/paragkatoch/Devops-DSOLS/internal/storage/postgres"
	"github.com/paragkatoch/Devops-DSOLS/services"
)

var registry = map[string]map[string]func(storage.Storage){
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
	flag.Parse()

	st := postgres.New()

	if compType, ok := registry[*componentType]; ok {
		if fn, ok := compType[*component]; ok {
			fn(st)
		} else {
			slog.Error("Invalid component")
		}
	} else {
		slog.Error("Invalid type")
	}
}

// package main

// import (
// 	"flag"
// 	"log/slog"

// 	"github.com/paragkatoch/Devops-DSOLS/controllers"
// 	"github.com/paragkatoch/Devops-DSOLS/internal/storage"
// 	"github.com/paragkatoch/Devops-DSOLS/internal/storage/postgres"
// 	"github.com/paragkatoch/Devops-DSOLS/services"
// )

// var serviceRegistry = map[string]func(storage.Storage){
// 	"order":   services.OrderService,
// 	"user":    services.UserService,
// 	"product": services.ProductService,
// }

// var controllerRegistry = map[string]func(){
// 	"order":   controllers.OrderController,
// 	"user":    controllers.UserController,
// 	"product": controllers.ProductController,
// }

// func main() {
// 	// parse flags
// 	slog.Info("checking server type and component")
// 	componentType := flag.String("type", "", "type of the server component (service/controller)")
// 	component := flag.String("component", "", "server component (order, user)")
// 	flag.Parse()

// 	switch *componentType {

// 	case "service":
// 		fn, ok := serviceRegistry[*component]
// 		if !ok {
// 			slog.Error("Invalid service")
// 			return
// 		}
// 		st := postgres.New()
// 		fn(st)

// 	case "controller":
// 		fn, ok := controllerRegistry[*component]
// 		if !ok {
// 			slog.Error("Invalid controller")
// 			return
// 		}

// 		fn()

// 	default:
// 		slog.Error("Invalid type")
// 	}
// }
