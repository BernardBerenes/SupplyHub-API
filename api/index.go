package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"github.com/BernardBerenes/SupplyHub-API/internal/app"
	"github.com/BernardBerenes/SupplyHub-API/internal/config"
)

var (
	once         sync.Once
	proxyHandler http.HandlerFunc
	initErr      error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		fiberApp, err := app.New(config.Load())
		if err != nil {
			initErr = err
			return
		}

		proxyHandler = adaptor.FiberApp(fiberApp)
	})

	if initErr != nil {
		log.Printf("failed to initialize application: %v", initErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	r.RequestURI = r.URL.RequestURI()

	proxyHandler(w, r)
}
