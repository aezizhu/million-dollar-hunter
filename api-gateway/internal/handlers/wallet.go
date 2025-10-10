package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type WalletHandler struct {
	logger zerolog.Logger
}

func NewWalletHandler(logger zerolog.Logger) *WalletHandler {
	return &WalletHandler{
		logger: logger,
	}
}

type Asset struct {
	TokenAddress string  `json:"tokenAddress"`
	Symbol       string  `json:"symbol"`
	Balance      float64 `json:"balance"`
	UsdValue     float64 `json:"usdValue"`
}

type HistoryPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

type WalletView struct {
	Address string         `json:"address"`
	Chain   string         `json:"chain"`
	Assets  []Asset        `json:"assets"`
	History []HistoryPoint `json:"history"`
}

type Transaction struct {
	Timestamp string  `json:"timestamp"`
	Type      string  `json:"type"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Symbol    string  `json:"symbol"`
	Amount    float64 `json:"amount"`
	UsdValue  float64 `json:"usdValue"`
	TxHash    string  `json:"txHash"`
}

type TransactionPage struct {
	Items    []Transaction `json:"items"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Total    int           `json:"total"`
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	address := c.Param("address")
	chain := c.Query("chain")

	h.logger.Info().
		Str("address", address).
		Str("chain", chain).
		Msg("get wallet request")

	
	c.JSON(http.StatusOK, WalletView{
		Address: address,
		Chain:   chain,
		Assets:  []Asset{},
		History: []HistoryPoint{},
	})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	address := c.Param("address")
	page := 1
	pageSize := 50

	h.logger.Info().
		Str("address", address).
		Int("page", page).
		Int("page_size", pageSize).
		Msg("get transactions request")

	
	c.JSON(http.StatusOK, TransactionPage{
		Items:    []Transaction{},
		Page:     page,
		PageSize: pageSize,
		Total:    0,
	})
}
