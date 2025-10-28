package bid_controller

import (
	"net/http"

	"github.com/auction-goexpert/internal/usecase/bid_usecase"
	"github.com/gin-gonic/gin"
)

type BidController struct {
	createBidUseCase *bid_usecase.CreateBidUseCase
	findBidUseCase   *bid_usecase.FindBidUseCase
}

func NewBidController(
	createBidUseCase *bid_usecase.CreateBidUseCase,
	findBidUseCase *bid_usecase.FindBidUseCase,
) *BidController {
	return &BidController{
		createBidUseCase: createBidUseCase,
		findBidUseCase:   findBidUseCase,
	}
}

func (bc *BidController) CreateBid(c *gin.Context) {
	var input bid_usecase.BidInputDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, internalErr := bc.createBidUseCase.Execute(c.Request.Context(), input)
	if internalErr != nil {
		c.JSON(internalErr.Code, gin.H{"error": internalErr.Message})
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (bc *BidController) FindBidByAuctionId(c *gin.Context) {
	auctionId := c.Param("auctionId")

	output, internalErr := bc.findBidUseCase.FindBidByAuctionId(c.Request.Context(), auctionId)
	if internalErr != nil {
		c.JSON(internalErr.Code, gin.H{"error": internalErr.Message})
		return
	}

	c.JSON(http.StatusOK, output)
}

func (bc *BidController) FindWinningBidByAuctionId(c *gin.Context) {
	auctionId := c.Param("auctionId")

	output, internalErr := bc.findBidUseCase.FindWinningBidByAuctionId(c.Request.Context(), auctionId)
	if internalErr != nil {
		c.JSON(internalErr.Code, gin.H{"error": internalErr.Message})
		return
	}

	c.JSON(http.StatusOK, output)
}
