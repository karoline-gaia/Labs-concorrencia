package auction_usecase

import (
	"context"

	"github.com/auction-goexpert/internal/entity"
	"github.com/auction-goexpert/internal/internal_error"
)

type FindAuctionUseCase struct {
	auctionRepository entity.AuctionRepositoryInterface
}

func NewFindAuctionUseCase(auctionRepository entity.AuctionRepositoryInterface) *FindAuctionUseCase {
	return &FindAuctionUseCase{
		auctionRepository: auctionRepository,
	}
}

func (au *FindAuctionUseCase) FindAuctionById(ctx context.Context, id string) (*AuctionOutputDTO, *internal_error.InternalError) {
	auction, err := au.auctionRepository.FindAuctionById(ctx, id)
	if err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	if auction == nil {
		return nil, internal_error.NewNotFoundError("auction not found")
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

func (au *FindAuctionUseCase) FindAuctions(ctx context.Context, status entity.AuctionStatus, category, productName string) ([]AuctionOutputDTO, *internal_error.InternalError) {
	auctions, err := au.auctionRepository.FindAuctions(ctx, status, category, productName)
	if err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	var output []AuctionOutputDTO
	for _, auction := range auctions {
		output = append(output, AuctionOutputDTO{
			Id:          auction.Id,
			ProductName: auction.ProductName,
			Category:    auction.Category,
			Description: auction.Description,
			Condition:   auction.Condition,
			Status:      auction.Status,
			Timestamp:   auction.Timestamp,
			ExpiresAt:   auction.ExpiresAt,
		})
	}

	return output, nil
}
