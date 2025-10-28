package auction_controller

import (
	"net/http"

	"github.com/auction-goexpert/internal/usecase/auction_usecase"
	"github.com/gin-gonic/gin"
)

type AuctionController struct {
	createAuctionUseCase *auction_usecase.CreateAuctionUseCase
	findAuctionUseCase   *auction_usecase.FindAuctionUseCase
}

func NewAuctionController(
	createAuctionUseCase *auction_usecase.CreateAuctionUseCase,
	findAuctionUseCase *auction_usecase.FindAuctionUseCase,
) *AuctionController {
	return &AuctionController{
		createAuctionUseCase: createAuctionUseCase,
		findAuctionUseCase:   findAuctionUseCase,
	}
}

func (ac *AuctionController) CreateAuction(c *gin.Context) {
	var input auction_usecase.AuctionInputDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, internalErr := ac.createAuctionUseCase.Execute(c.Request.Context(), input)
	if internalErr != nil {
		c.JSON(internalErr.Code, gin.H{"error": internalErr.Message})
		return
	}

	c.JSON(http.StatusCreated, output)
}

func (ac *AuctionController) FindAuctionById(c *gin.Context) {
	auctionId := c.Param("auctionId")

	output, internalErr := ac.findAuctionUseCase.FindAuctionById(c.Request.Context(), auctionId)
	if internalErr != nil {
		c.JSON(internalErr.Code, gin.H{"error": internalErr.Message})
		return
	}

	c.JSON(http.StatusOK, output)
}

func (ac *AuctionController) FindAuctions(c *gin.Context) {
	status := c.Query("status")
	category := c.Query("category")
	productName := c.Query("productName")

	var auctionStatus int = -1
	if status == "0" {
		auctionStatus = 0
	} else if status == "1" {
		auctionStatus = 1
	}

	output, internalErr := ac.findAuctionUseCase.FindAuctions(c.Request.Context(), auctionStatus, category, productName)
	if internalErr != nil {
		c.JSON(internalErr.Code, gin.H{"error": internalErr.Message})
		return
	}

	c.JSON(http.StatusOK, output)
}
