package main

import (
	"fourth-exam/api_gateway_evrone/internal/app"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// config 
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// app 
	app, err := app.NewApp(*config)
	if err != nil {
		log.Fatal(err)
	}

	// run app 
	go func() {
		app.Logger.Info("Listen: ", zap.String("address", config.Server.Host + config.Server.Port))
		if err := app.Run(); err != nil {
			app.Logger.Error("app run", zap.Error(err))
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<- sigs

	// app stops
	app.Logger.Info("api gateway stops")
	app.Stop()
}