package handlers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
)

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.AuthServiceURL == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable", "message": "auth service not configured"})
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
			return
		}
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, cfg.AuthServiceURL+"/api/v1/auth/login", bytes.NewReader(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "service_unavailable"})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
			return
		}
		defer resp.Body.Close()
		for k, vals := range resp.Header {
			for _, v := range vals {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Status(resp.StatusCode)
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			c.Error(err)
		}
	}
}

func Refresh(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.AuthServiceURL == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
			return
		}
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, cfg.AuthServiceURL+"/api/v1/auth/refresh", c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "service_unavailable"})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
			return
		}
		defer resp.Body.Close()
		for k, vals := range resp.Header {
			for _, v := range vals {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Status(resp.StatusCode)
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			c.Error(err)
		}
	}
}
