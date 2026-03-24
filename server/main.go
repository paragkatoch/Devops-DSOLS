package main

import (
	"flag"
	"log/slog"

	"github.com/paragkatoch/Devops-DSOLS/controllers"
	"github.com/paragkatoch/Devops-DSOLS/services"
)

var registry = map[string]map[string]func(){
	"service": {
		"order": services.OrderService,
		"user":  services.UserService,
	},
	"controller": {
		"order": controllers.OrderController,
		"user":  controllers.UserController,
	},
}

func main() {
	// parse flags
	slog.Info("checking server type and component")
	componentType := flag.String("type", "", "type of the server component (service/controller)")
	component := flag.String("component", "", "server component (order, user)")
	flag.Parse()

	if compType, ok := registry[*componentType]; ok {
		if fn, ok := compType[*component]; ok {
			fn()
		} else {
			slog.Error("Invalid component")
		}
	} else {
		slog.Error("Invalid type")
	}
}
