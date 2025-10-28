package auction

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/auction-goexpert/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionRepository struct {
	Collection *mongo.Collection
	mu         sync.RWMutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection: database.Collection("auctions"),
	}

	// Inicia a goroutine para verificar leilões expirados
	go repo.startAuctionExpirationChecker()

	return repo
}

// CreateAuction cria um novo leilão e calcula o tempo de expiração
func (ar *AuctionRepository) CreateAuction(ctx context.Context, auction *entity.Auction) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	// Calcula o tempo de duração do leilão baseado na variável de ambiente
	duration := calculateAuctionDuration()
	auction.ExpiresAt = time.Now().Add(duration)

	auctionEntityMongo := &entity.AuctionEntityMongo{
		Id:          auction.Id,
		ProductName: auction.ProductName,
		Category:    auction.Category,
		Description: auction.Description,
		Condition:   auction.Condition,
		Status:      auction.Status,
		Timestamp:   auction.Timestamp.Unix(),
		ExpiresAt:   auction.ExpiresAt.Unix(),
	}

	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		log.Printf("Error creating auction: %v", err)
		return err
	}

	log.Printf("Auction created successfully: %s, expires at: %s", auction.Id, auction.ExpiresAt.Format(time.RFC3339))
	return nil
}

// calculateAuctionDuration calcula a duração do leilão baseado na variável de ambiente AUCTION_DURATION
func calculateAuctionDuration() time.Duration {
	durationStr := os.Getenv("AUCTION_DURATION")
	if durationStr == "" {
		// Valor padrão: 5 minutos
		return 5 * time.Minute
	}

	durationSeconds, err := strconv.Atoi(durationStr)
	if err != nil {
		log.Printf("Invalid AUCTION_DURATION value, using default 5 minutes")
		return 5 * time.Minute
	}

	return time.Duration(durationSeconds) * time.Second
}

// getCheckInterval retorna o intervalo de verificação baseado na variável de ambiente
func getCheckInterval() time.Duration {
	intervalStr := os.Getenv("AUCTION_CHECK_INTERVAL")
	if intervalStr == "" {
		// Valor padrão: 10 segundos
		return 10 * time.Second
	}

	intervalSeconds, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("Invalid AUCTION_CHECK_INTERVAL value, using default 10 seconds")
		return 10 * time.Second
	}

	return time.Duration(intervalSeconds) * time.Second
}

// startAuctionExpirationChecker inicia uma goroutine que verifica periodicamente leilões expirados
func (ar *AuctionRepository) startAuctionExpirationChecker() {
	checkInterval := getCheckInterval()
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	log.Printf("Auction expiration checker started with interval: %v", checkInterval)

	for range ticker.C {
		ctx := context.Background()
		if err := ar.closeExpiredAuctions(ctx); err != nil {
			log.Printf("Error closing expired auctions: %v", err)
		}
	}
}

// closeExpiredAuctions busca e fecha todos os leilões que expiraram
func (ar *AuctionRepository) closeExpiredAuctions(ctx context.Context) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	now := time.Now().Unix()

	// Busca leilões ativos que já expiraram
	filter := bson.M{
		"status":     entity.Active,
		"expires_at": bson.M{"$lte": now},
	}

	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var expiredAuctions []entity.AuctionEntityMongo
	if err := cursor.All(ctx, &expiredAuctions); err != nil {
		return err
	}

	// Fecha cada leilão expirado
	for _, auction := range expiredAuctions {
		update := bson.M{
			"$set": bson.M{
				"status": entity.Completed,
			},
		}

		_, err := ar.Collection.UpdateOne(ctx, bson.M{"_id": auction.Id}, update)
		if err != nil {
			log.Printf("Error updating auction %s status: %v", auction.Id, err)
			continue
		}

		log.Printf("Auction %s closed automatically (expired at: %s)", 
			auction.Id, 
			time.Unix(auction.ExpiresAt, 0).Format(time.RFC3339))
	}

	if len(expiredAuctions) > 0 {
		log.Printf("Closed %d expired auction(s)", len(expiredAuctions))
	}

	return nil
}

// FindAuctionById busca um leilão pelo ID
func (ar *AuctionRepository) FindAuctionById(ctx context.Context, id string) (*entity.Auction, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	filter := bson.M{"_id": id}

	var auctionEntityMongo entity.AuctionEntityMongo
	err := ar.Collection.FindOne(ctx, filter).Decode(&auctionEntityMongo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &entity.Auction{
		Id:          auctionEntityMongo.Id,
		ProductName: auctionEntityMongo.ProductName,
		Category:    auctionEntityMongo.Category,
		Description: auctionEntityMongo.Description,
		Condition:   auctionEntityMongo.Condition,
		Status:      auctionEntityMongo.Status,
		Timestamp:   time.Unix(auctionEntityMongo.Timestamp, 0),
		ExpiresAt:   time.Unix(auctionEntityMongo.ExpiresAt, 0),
	}, nil
}

// FindAuctions busca leilões com filtros opcionais
func (ar *AuctionRepository) FindAuctions(ctx context.Context, status entity.AuctionStatus, category, productName string) ([]entity.Auction, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	filter := bson.M{}

	if status >= 0 {
		filter["status"] = status
	}

	if category != "" {
		filter["category"] = category
	}

	if productName != "" {
		filter["product_name"] = bson.M{"$regex": productName, "$options": "i"}
	}

	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var auctionEntitiesMongo []entity.AuctionEntityMongo
	if err := cursor.All(ctx, &auctionEntitiesMongo); err != nil {
		return nil, err
	}

	var auctions []entity.Auction
	for _, auctionMongo := range auctionEntitiesMongo {
		auctions = append(auctions, entity.Auction{
			Id:          auctionMongo.Id,
			ProductName: auctionMongo.ProductName,
			Category:    auctionMongo.Category,
			Description: auctionMongo.Description,
			Condition:   auctionMongo.Condition,
			Status:      auctionMongo.Status,
			Timestamp:   time.Unix(auctionMongo.Timestamp, 0),
			ExpiresAt:   time.Unix(auctionMongo.ExpiresAt, 0),
		})
	}

	return auctions, nil
}

// UpdateAuctionStatus atualiza o status de um leilão
func (ar *AuctionRepository) UpdateAuctionStatus(ctx context.Context, id string, status entity.AuctionStatus) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	return err
}

// FindExpiredAuctions busca leilões que expiraram
func (ar *AuctionRepository) FindExpiredAuctions(ctx context.Context) ([]entity.Auction, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	now := time.Now().Unix()

	filter := bson.M{
		"status":     entity.Active,
		"expires_at": bson.M{"$lte": now},
	}

	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var auctionEntitiesMongo []entity.AuctionEntityMongo
	if err := cursor.All(ctx, &auctionEntitiesMongo); err != nil {
		return nil, err
	}

	var auctions []entity.Auction
	for _, auctionMongo := range auctionEntitiesMongo {
		auctions = append(auctions, entity.Auction{
			Id:          auctionMongo.Id,
			ProductName: auctionMongo.ProductName,
			Category:    auctionMongo.Category,
			Description: auctionMongo.Description,
			Condition:   auctionMongo.Condition,
			Status:      auctionMongo.Status,
			Timestamp:   time.Unix(auctionMongo.Timestamp, 0),
			ExpiresAt:   time.Unix(auctionMongo.ExpiresAt, 0),
		})
	}

	return auctions, nil
}
