package app

import (
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"

	"github.com/BernardBerenes/SupplyHub-API/internal/config"
	"github.com/BernardBerenes/SupplyHub-API/internal/database"
	"github.com/BernardBerenes/SupplyHub-API/internal/middleware"
	"github.com/BernardBerenes/SupplyHub-API/internal/products"
	"github.com/BernardBerenes/SupplyHub-API/internal/storage"
	"github.com/BernardBerenes/SupplyHub-API/internal/stores"
	"github.com/BernardBerenes/SupplyHub-API/internal/transactiondetails"
	"github.com/BernardBerenes/SupplyHub-API/internal/transactions"
	"github.com/BernardBerenes/SupplyHub-API/internal/users"
)

type Container struct {
	JWTSecret                string
	UserHandler              *users.Handler
	ProductHandler           *products.Handler
	StoreHandler             *stores.Handler
	TransactionHandler       *transactions.Handler
	TransactionDetailHandler *transactiondetails.Handler
}

func NewContainer(db *gorm.DB, cfg *config.Config) (*Container, error) {
	minioClient, err := storage.NewMinIO(cfg)
	if err != nil {
		return nil, err
	}

	userRepo := users.NewRepository(db)
	userUseCase := users.NewUseCase(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	userHandler := users.NewHandler(userUseCase)

	productRepo := products.NewRepository(db)
	productUseCase := products.NewUseCase(productRepo, minioClient)
	productHandler := products.NewHandler(productUseCase)

	storeRepo := stores.NewRepository(db)
	storeUseCase := stores.NewUseCase(storeRepo)
	storeHandler := stores.NewHandler(storeUseCase)

	transactionRepo := transactions.NewRepository(db)
	transactionUseCase := transactions.NewUseCase(transactionRepo, storeUseCase)

	transactionDetailRepo := transactiondetails.NewRepository(db)
	transactionDetailUseCase := transactiondetails.NewUseCase(transactionDetailRepo, productUseCase, transactionUseCase)
	transactionDetailHandler := transactiondetails.NewHandler(transactionDetailUseCase)

	transactionHandler := transactions.NewHandler(transactionUseCase, transactionDetailUseCase)

	return &Container{
		JWTSecret:                cfg.JWTSecret,
		UserHandler:              userHandler,
		ProductHandler:           productHandler,
		StoreHandler:             storeHandler,
		TransactionHandler:       transactionHandler,
		TransactionDetailHandler: transactionDetailHandler,
	}, nil
}

func RegisterRoutes(app *fiber.App, c *Container) {
	api := app.Group("/api/v1")

	api.Post("/login", c.UserHandler.Login)

	auth := middleware.JWTAuth(c.JWTSecret)

	products := api.Group("/products", auth)
	products.Post("", c.ProductHandler.Create)
	products.Get("", c.ProductHandler.List)
	products.Get("/:uuid", c.ProductHandler.Detail)
	products.Post("/paginate", c.ProductHandler.Paginate)
	products.Patch("/:uuid", c.ProductHandler.Update)
	products.Delete("/:uuid", c.ProductHandler.Delete)

	stores := api.Group("/stores", auth)
	stores.Post("", c.StoreHandler.Create)
	stores.Get("", c.StoreHandler.List)
	stores.Post("/paginate", c.StoreHandler.Paginate)
	stores.Patch("/:uuid", c.StoreHandler.Update)
	stores.Delete("/:uuid", c.StoreHandler.Delete)

	transactions := api.Group("/transactions", auth)
	transactions.Post("", c.TransactionHandler.Create)
	transactions.Post("/paginate", c.TransactionHandler.Paginate)
	transactions.Post("/sync", c.TransactionHandler.Sync)
	transactions.Patch("/:uuid", c.TransactionHandler.Update)
	transactions.Delete("/:uuid", c.TransactionHandler.Delete)

	transactionDetails := transactions.Group("/:transaction_id/details")
	transactionDetails.Post("", c.TransactionDetailHandler.Create)
	transactionDetails.Post("/paginate", c.TransactionDetailHandler.Paginate)
	transactionDetails.Patch("/:uuid", c.TransactionDetailHandler.Update)
	transactionDetails.Delete("/:uuid", c.TransactionDetailHandler.Delete)
}

// New builds the Fiber app: connects the database, runs the schema
// migration, wires the container, and registers routes. Both the
// persistent server entrypoint (cmd/server) and the Vercel serverless
// entrypoint (api/index.go) share this so route wiring never drifts
// between the two.
func New(cfg *config.Config) (*fiber.App, error) {
	db, err := database.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db,
		&users.User{},
		&products.Product{},
		&stores.Store{},
		&transactions.Transaction{},
		&transactiondetails.TransactionDetail{},
	); err != nil {
		return nil, err
	}

	fiberApp := fiber.New()
	fiberApp.Use(middleware.Recovery)
	fiberApp.Use(middleware.CORS(cfg.CORSOrigins))

	container, err := NewContainer(db, cfg)
	if err != nil {
		return nil, err
	}

	RegisterRoutes(fiberApp, container)

	return fiberApp, nil
}
