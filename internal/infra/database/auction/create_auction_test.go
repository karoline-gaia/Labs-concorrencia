package auction

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/auction-goexpert/internal/entity"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()
	
	// Conecta ao MongoDB de teste
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:admin@localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	database := client.Database("auctions_test")

	// Limpa a coleção antes dos testes
	database.Collection("auctions").Drop(ctx)

	cleanup := func() {
		database.Collection("auctions").Drop(ctx)
		client.Disconnect(ctx)
	}

	return database, cleanup
}

func TestCreateAuction(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Define duração de teste
	os.Setenv("AUCTION_DURATION", "5")
	defer os.Unsetenv("AUCTION_DURATION")

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	auction, err := entity.CreateAuction(
		"iPhone 13",
		"Electronics",
		"Brand new iPhone 13 with 128GB storage",
		entity.New,
		0,
	)
	assert.NoError(t, err)

	err = repo.CreateAuction(ctx, auction)
	assert.NoError(t, err)
	assert.NotEmpty(t, auction.Id)
	assert.Equal(t, entity.Active, auction.Status)
	assert.False(t, auction.ExpiresAt.IsZero())
}

func TestFindAuctionById(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.Setenv("AUCTION_DURATION", "300")
	defer os.Unsetenv("AUCTION_DURATION")

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	// Cria um leilão
	auction, _ := entity.CreateAuction(
		"MacBook Pro",
		"Electronics",
		"MacBook Pro 2021 with M1 chip",
		entity.New,
		0,
	)
	repo.CreateAuction(ctx, auction)

	// Busca o leilão
	foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
	assert.NoError(t, err)
	assert.NotNil(t, foundAuction)
	assert.Equal(t, auction.Id, foundAuction.Id)
	assert.Equal(t, auction.ProductName, foundAuction.ProductName)
}

func TestAuctionAutomaticClosure(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Define duração curta para teste (3 segundos)
	os.Setenv("AUCTION_DURATION", "3")
	// Define intervalo de verificação curto (1 segundo)
	os.Setenv("AUCTION_CHECK_INTERVAL", "1")
	defer os.Unsetenv("AUCTION_DURATION")
	defer os.Unsetenv("AUCTION_CHECK_INTERVAL")

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	// Cria um leilão
	auction, err := entity.CreateAuction(
		"Test Product",
		"Test Category",
		"This is a test product for automatic closure",
		entity.New,
		0,
	)
	assert.NoError(t, err)

	err = repo.CreateAuction(ctx, auction)
	assert.NoError(t, err)

	// Verifica que o leilão está ativo
	foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
	assert.NoError(t, err)
	assert.Equal(t, entity.Active, foundAuction.Status)

	t.Logf("Auction created at: %s", foundAuction.Timestamp.Format(time.RFC3339))
	t.Logf("Auction expires at: %s", foundAuction.ExpiresAt.Format(time.RFC3339))
	t.Logf("Current time: %s", time.Now().Format(time.RFC3339))

	// Aguarda o leilão expirar + tempo para a goroutine processar
	// 3 segundos (duração) + 2 segundos (margem para processamento)
	time.Sleep(5 * time.Second)

	// Verifica que o leilão foi fechado automaticamente
	closedAuction, err := repo.FindAuctionById(ctx, auction.Id)
	assert.NoError(t, err)
	assert.NotNil(t, closedAuction)
	assert.Equal(t, entity.Completed, closedAuction.Status, "Auction should be automatically closed")

	t.Logf("Auction status after expiration: %d (0=Active, 1=Completed)", closedAuction.Status)
}

func TestFindExpiredAuctions(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.Setenv("AUCTION_DURATION", "1")
	defer os.Unsetenv("AUCTION_DURATION")

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	// Cria um leilão que expirará rapidamente
	auction, _ := entity.CreateAuction(
		"Expired Product",
		"Test",
		"This product should expire quickly",
		entity.Used,
		0,
	)
	repo.CreateAuction(ctx, auction)

	// Aguarda expiração
	time.Sleep(2 * time.Second)

	// Busca leilões expirados
	expiredAuctions, err := repo.FindExpiredAuctions(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, expiredAuctions)
}

func TestUpdateAuctionStatus(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.Setenv("AUCTION_DURATION", "300")
	defer os.Unsetenv("AUCTION_DURATION")

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	// Cria um leilão
	auction, _ := entity.CreateAuction(
		"Test Product",
		"Test",
		"Test description for status update",
		entity.New,
		0,
	)
	repo.CreateAuction(ctx, auction)

	// Atualiza o status
	err := repo.UpdateAuctionStatus(ctx, auction.Id, entity.Completed)
	assert.NoError(t, err)

	// Verifica se o status foi atualizado
	updatedAuction, err := repo.FindAuctionById(ctx, auction.Id)
	assert.NoError(t, err)
	assert.Equal(t, entity.Completed, updatedAuction.Status)
}

func TestConcurrentAuctionCreation(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	os.Setenv("AUCTION_DURATION", "300")
	defer os.Unsetenv("AUCTION_DURATION")

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	// Cria múltiplos leilões concorrentemente
	numAuctions := 10
	done := make(chan bool, numAuctions)

	for i := 0; i < numAuctions; i++ {
		go func(index int) {
			auction, _ := entity.CreateAuction(
				"Concurrent Product",
				"Test",
				"Testing concurrent creation",
				entity.New,
				0,
			)
			err := repo.CreateAuction(ctx, auction)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Aguarda todas as goroutines terminarem
	for i := 0; i < numAuctions; i++ {
		<-done
	}

	// Verifica se todos os leilões foram criados
	auctions, err := repo.FindAuctions(ctx, entity.Active, "", "")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(auctions), numAuctions)
}

func TestCalculateAuctionDuration(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{
			name:     "Valid duration",
			envValue: "300",
			expected: 300 * time.Second,
		},
		{
			name:     "Empty value - default",
			envValue: "",
			expected: 5 * time.Minute,
		},
		{
			name:     "Invalid value - default",
			envValue: "invalid",
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("AUCTION_DURATION", tt.envValue)
				defer os.Unsetenv("AUCTION_DURATION")
			}

			duration := calculateAuctionDuration()
			assert.Equal(t, tt.expected, duration)
		})
	}
}

func TestCloseExpiredAuctionsDirectly(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewAuctionRepository(database)
	ctx := context.Background()

	// Cria um leilão já expirado manualmente no banco
	expiredAuction := &entity.AuctionEntityMongo{
		Id:          "expired-auction-id",
		ProductName: "Expired Product",
		Category:    "Test",
		Description: "This auction is already expired",
		Condition:   entity.New,
		Status:      entity.Active,
		Timestamp:   time.Now().Add(-10 * time.Minute).Unix(),
		ExpiresAt:   time.Now().Add(-5 * time.Minute).Unix(), // Expirado há 5 minutos
	}

	_, err := database.Collection("auctions").InsertOne(ctx, expiredAuction)
	assert.NoError(t, err)

	// Chama diretamente o método de fechamento
	err = repo.closeExpiredAuctions(ctx)
	assert.NoError(t, err)

	// Verifica se o leilão foi fechado
	var result entity.AuctionEntityMongo
	err = database.Collection("auctions").FindOne(ctx, bson.M{"_id": "expired-auction-id"}).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, entity.Completed, result.Status)
}
