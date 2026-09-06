package main

import (
	"log"

	"github.com/BernardBerenes/SupplyHub-API/internal/app"
	"github.com/BernardBerenes/SupplyHub-API/internal/config"
)

func main() {
	cfg := config.Load()

	fiberApp, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	log.Fatal(fiberApp.Listen(":" + cfg.Port))
}
