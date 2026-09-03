package main

import (
	"gorm.io/gorm"

	"github.com/BernardBerenes/SupplyHub-API/internal/config"
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
