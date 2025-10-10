package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type ExportHandler struct {
	logger zerolog.Logger
}

func NewExportHandler(logger zerolog.Logger) *ExportHandler {
	return &ExportHandler{
		logger: logger,
	}
}

func (h *ExportHandler) ExportWallet(c *gin.Context) {
	address := c.Param("address")
	chain := c.Query("chain")
	format := c.DefaultQuery("format", "json")

	h.logger.Info().
		Str("address", address).
		Str("chain", chain).
		Str("format", format).
		Msg("export wallet request")

	
	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=wallet_"+address+".csv")
		c.String(http.StatusOK, "# CSV export not yet implemented\n")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "export not yet implemented",
		"address": address,
		"chain":   chain,
	})
}
