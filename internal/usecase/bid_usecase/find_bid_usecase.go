package bid_usecase

import (
	"context"

	"github.com/auction-goexpert/internal/entity"
	"github.com/auction-goexpert/internal/internal_error"
)

type FindBidUseCase struct {
	bidRepository entity.BidRepositoryInterface
}

func NewFindBidUseCase(bidRepository entity.BidRepositoryInterface) *FindBidUseCase {
	return &FindBidUseCase{
		bidRepository: bidRepository,
	}
}

func (bu *FindBidUseCase) FindBidByAuctionId(ctx context.Context, auctionId string) ([]BidOutputDTO, *internal_error.InternalError) {
	bids, err := bu.bidRepository.FindBidByAuctionId(ctx, auctionId)
	if err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	var output []BidOutputDTO
	for _, bid := range bids {
		output = append(output, BidOutputDTO{
			Id:        bid.Id,
			UserId:    bid.UserId,
			AuctionId: bid.AuctionId,
			Amount:    bid.Amount,
		})
	}

	return output, nil
}

func (bu *FindBidUseCase) FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*BidOutputDTO, *internal_error.InternalError) {
	bid, err := bu.bidRepository.FindWinningBidByAuctionId(ctx, auctionId)
	if err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	if bid == nil {
		return nil, internal_error.NewNotFoundError("no bids found for this auction")
	}

	return &BidOutputDTO{
		Id:        bid.Id,
		UserId:    bid.UserId,
		AuctionId: bid.AuctionId,
		Amount:    bid.Amount,
	}, nil
}
