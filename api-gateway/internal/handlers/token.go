package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type TokenHandler struct {
	logger zerolog.Logger
}

func NewTokenHandler(logger zerolog.Logger) *TokenHandler {
	return &TokenHandler{
		logger: logger,
	}
}

type Holder struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
	Percent float64 `json:"percent"`
}

type TopHoldersResponse struct {
	Token   string   `json:"token"`
	Holders []Holder `json:"holders"`
}

func (h *TokenHandler) GetTopHolders(c *gin.Context) {
	tokenAddress := c.Param("tokenAddress")
	chain := c.Query("chain")

	h.logger.Info().
		Str("token_address", tokenAddress).
		Str("chain", chain).
		Msg("get top holders request")

	
	c.JSON(http.StatusOK, TopHoldersResponse{
		Token:   tokenAddress,
		Holders: []Holder{},
	})
}
