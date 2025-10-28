package bid_usecase

import (
	"context"

	"github.com/auction-goexpert/internal/entity"
	"github.com/auction-goexpert/internal/internal_error"
)

type BidInputDTO struct {
	UserId    string  `json:"user_id" binding:"required"`
	AuctionId string  `json:"auction_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

type BidOutputDTO struct {
	Id        string  `json:"id"`
	UserId    string  `json:"user_id"`
	AuctionId string  `json:"auction_id"`
	Amount    float64 `json:"amount"`
}

type CreateBidUseCase struct {
	bidRepository entity.BidRepositoryInterface
}

func NewCreateBidUseCase(bidRepository entity.BidRepositoryInterface) *CreateBidUseCase {
	return &CreateBidUseCase{
		bidRepository: bidRepository,
	}
}

func (bu *CreateBidUseCase) Execute(ctx context.Context, input BidInputDTO) (*BidOutputDTO, *internal_error.InternalError) {
	bid, err := entity.CreateBid(input.UserId, input.AuctionId, input.Amount)
	if err != nil {
		return nil, internal_error.NewInternalServerError(err.Error())
	}

	if err := bu.bidRepository.CreateBid(ctx, bid); err != nil {
		return nil, internal_error.NewBadRequestError(err.Error())
	}

	return &BidOutputDTO{
		Id:        bid.Id,
		UserId:    bid.UserId,
		AuctionId: bid.AuctionId,
		Amount:    bid.Amount,
	}, nil
}
