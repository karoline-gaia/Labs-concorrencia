package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Bid struct {
	Id        string
	UserId    string
	AuctionId string
	Amount    float64
	Timestamp time.Time
}

type BidEntityMongo struct {
	Id        string  `bson:"_id"`
	UserId    string  `bson:"user_id"`
	AuctionId string  `bson:"auction_id"`
	Amount    float64 `bson:"amount"`
	Timestamp int64   `bson:"timestamp"`
}

type BidRepositoryInterface interface {
	CreateBid(ctx context.Context, bid *Bid) error
	FindBidByAuctionId(ctx context.Context, auctionId string) ([]Bid, error)
	FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*Bid, error)
}

func CreateBid(userId, auctionId string, amount float64) (*Bid, error) {
	bid := &Bid{
		Id:        uuid.New().String(),
		UserId:    userId,
		AuctionId: auctionId,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	return bid, nil
}
