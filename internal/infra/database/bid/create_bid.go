package bid

import (
	"context"
	"errors"
	"log"

	"github.com/auction-goexpert/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BidRepository struct {
	Collection        *mongo.Collection
	AuctionRepository entity.AuctionRepositoryInterface
}

func NewBidRepository(database *mongo.Database, auctionRepo entity.AuctionRepositoryInterface) *BidRepository {
	return &BidRepository{
		Collection:        database.Collection("bids"),
		AuctionRepository: auctionRepo,
	}
}

// CreateBid cria um novo lance, validando se o leilão está ativo
func (br *BidRepository) CreateBid(ctx context.Context, bid *entity.Bid) error {
	// Valida se o leilão existe e está ativo
	auction, err := br.AuctionRepository.FindAuctionById(ctx, bid.AuctionId)
	if err != nil {
		return err
	}

	if auction == nil {
		return errors.New("auction not found")
	}

	// Verifica se o leilão está ativo
	if auction.Status != entity.Active {
		return errors.New("auction is not active")
	}

	// Verifica se o leilão expirou
	if auction.IsExpired() {
		return errors.New("auction has expired")
	}

	bidEntityMongo := &entity.BidEntityMongo{
		Id:        bid.Id,
		UserId:    bid.UserId,
		AuctionId: bid.AuctionId,
		Amount:    bid.Amount,
		Timestamp: bid.Timestamp.Unix(),
	}

	_, err = br.Collection.InsertOne(ctx, bidEntityMongo)
	if err != nil {
		log.Printf("Error creating bid: %v", err)
		return err
	}

	log.Printf("Bid created successfully: %s for auction: %s", bid.Id, bid.AuctionId)
	return nil
}

// FindBidByAuctionId busca todos os lances de um leilão
func (br *BidRepository) FindBidByAuctionId(ctx context.Context, auctionId string) ([]entity.Bid, error) {
	filter := bson.M{"auction_id": auctionId}

	cursor, err := br.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bidEntitiesMongo []entity.BidEntityMongo
	if err := cursor.All(ctx, &bidEntitiesMongo); err != nil {
		return nil, err
	}

	var bids []entity.Bid
	for _, bidMongo := range bidEntitiesMongo {
		bids = append(bids, entity.Bid{
			Id:        bidMongo.Id,
			UserId:    bidMongo.UserId,
			AuctionId: bidMongo.AuctionId,
			Amount:    bidMongo.Amount,
		})
	}

	return bids, nil
}

// FindWinningBidByAuctionId busca o lance vencedor (maior valor) de um leilão
func (br *BidRepository) FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*entity.Bid, error) {
	filter := bson.M{"auction_id": auctionId}
	opts := options.FindOne().SetSort(bson.D{{Key: "amount", Value: -1}})

	var bidEntityMongo entity.BidEntityMongo
	err := br.Collection.FindOne(ctx, filter, opts).Decode(&bidEntityMongo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &entity.Bid{
		Id:        bidEntityMongo.Id,
		UserId:    bidEntityMongo.UserId,
		AuctionId: bidEntityMongo.AuctionId,
		Amount:    bidEntityMongo.Amount,
	}, nil
}
