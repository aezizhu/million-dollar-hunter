package handlers

import (
	"net/http"

	"api-gateway/internal/metrics"
	"api-gateway/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AuthHandler struct {
	authMiddleware *middleware.AuthMiddleware
	logger         zerolog.Logger
}

func NewAuthHandler(authMiddleware *middleware.AuthMiddleware, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type RegisterResponse struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

type TokenRefreshResponse struct {
	AccessToken string `json:"accessToken"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		metrics.AuthAttempts.WithLabelValues("validation_error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("login attempt")

	username := "aezi"
	if !h.authMiddleware.ValidateMVPCredentials(username, req.Password) {
		metrics.AuthAttempts.WithLabelValues("failed").Inc()
		h.logger.Warn().Str("email", req.Email).Msg("login failed - invalid credentials")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_credentials",
			"message": "invalid email or password",
		})
		return
	}

	userID := uuid.New().String() // For MVP, generate a user ID
	accessToken, err := h.authMiddleware.GenerateToken(userID, req.Email, username)
	if err != nil {
		metrics.AuthAttempts.WithLabelValues("token_generation_error").Inc()
		h.logger.Error().Err(err).Msg("failed to generate access token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to generate token",
		})
		return
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(userID)
	if err != nil {
		metrics.AuthAttempts.WithLabelValues("token_generation_error").Inc()
		h.logger.Error().Err(err).Msg("failed to generate refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to generate token",
		})
		return
	}

	metrics.AuthAttempts.WithLabelValues("success").Inc()
	h.logger.Info().Str("user_id", userID).Msg("login successful")

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("registration attempt (disabled in MVP)")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "registration is disabled in MVP - use hardcoded credentials",
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	h.logger.Info().Msg("token refresh attempt")
	
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "token refresh not yet implemented in MVP",
	})
}
