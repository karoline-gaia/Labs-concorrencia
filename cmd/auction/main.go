package main

import (
	"context"
	"log"

	"github.com/auction-goexpert/configuration/database/mongodb"
	"github.com/auction-goexpert/internal/infra/api/web/controller/auction_controller"
	"github.com/auction-goexpert/internal/infra/api/web/controller/bid_controller"
	"github.com/auction-goexpert/internal/infra/database/auction"
	"github.com/auction-goexpert/internal/infra/database/bid"
	"github.com/auction-goexpert/internal/infra/database/user"
	"github.com/auction-goexpert/internal/usecase/auction_usecase"
	"github.com/auction-goexpert/internal/usecase/bid_usecase"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	// Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Conecta ao MongoDB
	database, err := mongodb.NewMongoDBConnection(ctx)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Inicializa repositories
	auctionRepo := auction.NewAuctionRepository(database)
	userRepo := user.NewUserRepository(database)
	bidRepo := bid.NewBidRepository(database, auctionRepo)

	// Inicializa use cases
	createAuctionUseCase := auction_usecase.NewCreateAuctionUseCase(auctionRepo)
	findAuctionUseCase := auction_usecase.NewFindAuctionUseCase(auctionRepo)
	createBidUseCase := bid_usecase.NewCreateBidUseCase(bidRepo)
	findBidUseCase := bid_usecase.NewFindBidUseCase(bidRepo)

	// Inicializa controllers
	auctionController := auction_controller.NewAuctionController(createAuctionUseCase, findAuctionUseCase)
	bidController := bid_controller.NewBidController(createBidUseCase, findBidUseCase)

	// Configura rotas
	router := gin.Default()

	// Rotas de leilão
	router.POST("/auction", auctionController.CreateAuction)
	router.GET("/auction/:auctionId", auctionController.FindAuctionById)
	router.GET("/auction", auctionController.FindAuctions)

	// Rotas de lance
	router.POST("/bid", bidController.CreateBid)
	router.GET("/bid/auction/:auctionId", bidController.FindBidByAuctionId)
	router.GET("/bid/auction/:auctionId/winner", bidController.FindWinningBidByAuctionId)

	log.Println("Server starting on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// Mantém o userRepo para evitar warning de variável não utilizada
	_ = userRepo
}
