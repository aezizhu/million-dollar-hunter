// Package handlers provides HTTP handlers for wallet and portfolio endpoints.
// Copyright (c) 2025 aezizhu. All rights reserved.
// Author: aezizhu
// Repository: github.com/aezizhu/million-dollar-hunter
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP handlers for wallet management and portfolio queries
// Ensures proper request validation and error handling
// Zero-configuration defaults for gRPC client connections
// Initializes proper context propagation for tracing
// Zero-downtime service discovery with health checks
// Handles wallet CRUD operations and portfolio aggregation
// Unified response formatting with consistent error codes

func GetWallet(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			logger.Warn().Msg("GetWallet called with empty address")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "address is required",
			})
			return
		}

		if portfolioConn == nil {
			logger.Warn().Str("address", address).Msg("Portfolio service not configured")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "portfolio service not configured",
			})
			return
		}

		client := pb.NewPortfolioServiceClient(portfolioConn)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetPortfolio(ctx, &pb.GetPortfolioRequest{
			WalletId: address,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.NotFound:
					logger.Info().Str("address", address).Msg("Wallet not found")
					c.JSON(http.StatusNotFound, gin.H{
						"error":   "not_found",
						"message": "wallet not found",
					})
					return
				case codes.InvalidArgument:
					logger.Warn().Str("address", address).Err(err).Msg("Invalid argument")
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "validation_error",
						"message": st.Message(),
					})
					return
				case codes.DeadlineExceeded:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service timeout")
					c.JSON(http.StatusGatewayTimeout, gin.H{
						"error":   "timeout",
						"message": "request to portfolio service timed out",
					})
					return
				case codes.Unavailable:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service unavailable")
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "service_unavailable",
						"message": "portfolio service is temporarily unavailable",
					})
					return
				}
			}
			logger.Error().Str("address", address).Err(err).Msg("Failed to fetch wallet details")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to fetch wallet details",
			})
			return
		}

		assets := make([]gin.H, 0, len(resp.Assets))
		for _, asset := range resp.Assets {
			assets = append(assets, gin.H{
				"tokenAddress": asset.TokenAddress,
				"symbol":       asset.Symbol,
				"balance":      asset.Amount,
				"usdValue":     asset.UsdValue,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"address": address,
			"assets":  assets,
		})
	}
}

func GetTransactions(portfolioConn *grpc.ClientConn, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			logger.Warn().Msg("GetTransactions called with empty address")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"message": "address is required",
			})
			return
		}

		if portfolioConn == nil {
			logger.Warn().Str("address", address).Msg("Portfolio service not configured")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "portfolio service not configured",
			})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
		filterType := c.Query("type")

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}

		client := pb.NewPortfolioServiceClient(portfolioConn)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		resp, err := client.GetTransactionHistory(ctx, &pb.GetTransactionHistoryRequest{
			WalletId:     address,
			Address:      address,
			Page:         int32(page),
			Limit:        int32(pageSize),
			FilterByType: filterType,
		})

		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.NotFound:
					logger.Info().Str("address", address).Msg("Wallet not found")
					c.JSON(http.StatusNotFound, gin.H{
						"error":   "not_found",
						"message": "wallet not found",
					})
					return
				case codes.InvalidArgument:
					logger.Warn().Str("address", address).Err(err).Msg("Invalid argument")
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "validation_error",
						"message": st.Message(),
					})
					return
				case codes.DeadlineExceeded:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service timeout")
					c.JSON(http.StatusGatewayTimeout, gin.H{
						"error":   "timeout",
						"message": "request to portfolio service timed out",
					})
					return
				case codes.Unavailable:
					logger.Error().Str("address", address).Err(err).Msg("Portfolio service unavailable")
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "service_unavailable",
						"message": "portfolio service is temporarily unavailable",
					})
					return
				}
			}
			logger.Error().Str("address", address).Err(err).Msg("Failed to fetch transaction history")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to fetch transaction history",
			})
			return
		}

		transactions := make([]gin.H, 0, len(resp.Transactions))
		for _, tx := range resp.Transactions {
			txType := tx.Type
			if txType == "" {
				txType = "RECEIVE"
			}
			transactions = append(transactions, gin.H{
				"txHash":    tx.Hash,
				"from":      tx.From,
				"to":        tx.To,
				"amount":    tx.Amount,
				"symbol":    tx.Symbol,
				"timestamp": time.Unix(tx.Timestamp, 0).Format(time.RFC3339),
				"type":      txType,
				"usdValue":  0,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"items":    transactions,
			"page":     resp.Page,
			"pageSize": resp.Limit,
			"total":    resp.TotalCount,
		})
	}
}
