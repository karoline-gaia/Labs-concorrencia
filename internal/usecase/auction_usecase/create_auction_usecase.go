package auction_usecase

import (
	"context"
	"time"

	"github.com/auction-goexpert/internal/entity"
	"github.com/auction-goexpert/internal/internal_error"
)

type AuctionInputDTO struct {
	ProductName string                `json:"product_name" binding:"required,min=1"`
	Category    string                `json:"category" binding:"required,min=2"`
	Description string                `json:"description" binding:"required,min=10,max=200"`
	Condition   entity.ProductCondition `json:"condition" binding:"oneof=0 1 2"`
}

type AuctionOutputDTO struct {
	Id          string                `json:"id"`
	ProductName string                `json:"product_name"`
	Category    string                `json:"category"`
	Description string                `json:"description"`
	Condition   entity.ProductCondition `json:"condition"`
	Status      entity.AuctionStatus    `json:"status"`
	Timestamp   time.Time             `json:"timestamp"`
	ExpiresAt   time.Time             `json:"expires_at"`
}

type CreateAuctionUseCase struct {
	auctionRepository entity.AuctionRepositoryInterface
}

func NewCreateAuctionUseCase(auctionRepository entity.AuctionRepositoryInterface) *CreateAuctionUseCase {
	return &CreateAuctionUseCase{
		auctionRepository: auctionRepository,
	}
}

func (au *CreateAuctionUseCase) Execute(ctx context.Context, input AuctionInputDTO) (*AuctionOutputDTO, *internal_error.InternalError) {
	auction, err := entity.CreateAuction(
		input.ProductName,
		input.Category,
		input.Description,
		input.Condition,
		0, // Duration será calculada no repository
	)
	if err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	if err := au.auctionRepository.CreateAuction(ctx, auction); err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	return &AuctionOutputDTO{
		Id:          auction.Id,
		ProductName: auction.ProductName,
		Category:    auction.Category,
		Description: auction.Description,
		Condition:   auction.Condition,
		Status:      auction.Status,
		Timestamp:   auction.Timestamp,
		ExpiresAt:   auction.ExpiresAt,
	}, nil
}
