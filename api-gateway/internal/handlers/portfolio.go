package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type PortfolioHandler struct {
	logger zerolog.Logger
}

func NewPortfolioHandler(logger zerolog.Logger) *PortfolioHandler {
	return &PortfolioHandler{
		logger: logger,
	}
}

type PortfolioSummary struct {
	ID          string  `json:"id"`
	Address     string  `json:"address"`
	Chain       string  `json:"chain"`
	NetWorthUsd float64 `json:"netWorthUsd"`
}

type PortfolioListResponse struct {
	Items    []PortfolioSummary `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
	Total    int                `json:"total"`
}

type AddWalletRequest struct {
	Address  string `json:"address" binding:"required"`
	Chain    string `json:"chain" binding:"required"`
	Nickname string `json:"nickname" binding:"max=64"`
}

type AddWalletResponse struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

func (h *PortfolioHandler) ListPortfolios(c *gin.Context) {
	
	page := 1
	pageSize := 20
	
	h.logger.Info().
		Int("page", page).
		Int("page_size", pageSize).
		Msg("listing portfolios")

	c.JSON(http.StatusOK, PortfolioListResponse{
		Items:    []PortfolioSummary{},
		Page:     page,
		PageSize: pageSize,
		Total:    0,
	})
}

func (h *PortfolioHandler) AddWallet(c *gin.Context) {
	var req AddWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("address", req.Address).
		Str("chain", req.Chain).
		Msg("add wallet request")

	
	jobID := uuid.New().String()
	
	c.JSON(http.StatusAccepted, AddWalletResponse{
		JobID:  jobID,
		Status: "queued",
	})
}
