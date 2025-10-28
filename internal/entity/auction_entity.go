package entity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuctionStatus int

const (
	Active AuctionStatus = iota
	Completed
)

type Auction struct {
	Id          string
	ProductName string
	Category    string
	Description string
	Condition   ProductCondition
	Status      AuctionStatus
	Timestamp   time.Time
	ExpiresAt   time.Time
}

type AuctionEntityMongo struct {
	Id          string           `bson:"_id"`
	ProductName string           `bson:"product_name"`
	Category    string           `bson:"category"`
	Description string           `bson:"description"`
	Condition   ProductCondition `bson:"condition"`
	Status      AuctionStatus    `bson:"status"`
	Timestamp   int64            `bson:"timestamp"`
	ExpiresAt   int64            `bson:"expires_at"`
}

type ProductCondition int

const (
	New ProductCondition = iota
	Used
	Refurbished
)

type AuctionRepositoryInterface interface {
	CreateAuction(ctx context.Context, auction *Auction) error
	FindAuctionById(ctx context.Context, id string) (*Auction, error)
	FindAuctions(ctx context.Context, status AuctionStatus, category, productName string) ([]Auction, error)
	UpdateAuctionStatus(ctx context.Context, id string, status AuctionStatus) error
	FindExpiredAuctions(ctx context.Context) ([]Auction, error)
}

func CreateAuction(productName, category, description string, condition ProductCondition, duration time.Duration) (*Auction, error) {
	auction := &Auction{
		Id:          uuid.New().String(),
		ProductName: productName,
		Category:    category,
		Description: description,
		Condition:   condition,
		Status:      Active,
		Timestamp:   time.Now(),
		ExpiresAt:   time.Now().Add(duration),
	}

	return auction, nil
}

func (a *Auction) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}
