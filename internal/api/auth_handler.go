package api

import (
	"net/http"

	"minimax-voice-workbench/internal/auth"
	"minimax-voice-workbench/internal/config"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	TurnstileToken string `json:"turnstile_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if config.GlobalConfig.Security.Captcha.Enabled {
		if req.TurnstileToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Captcha token required"})
			return
		}

		result, err := auth.VerifyTurnstileToken(c.Request.Context(), req.TurnstileToken, c.ClientIP())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Captcha verification failed"})
			return
		}
		if !result.Success {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Captcha verification failed"})
			return
		}
	}

	if !auth.CheckCredentials(req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokens(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type PublicConfigResponse struct {
	AuthEnabled    bool   `json:"auth_enabled"`
	CaptchaEnabled bool   `json:"captcha_enabled"`
	CaptchaSiteKey string `json:"captcha_sitekey,omitempty"`
}

func GetPublicConfig(c *gin.Context) {
	c.JSON(http.StatusOK, PublicConfigResponse{
		AuthEnabled:    config.GlobalConfig.Security.Enabled,
		CaptchaEnabled: config.GlobalConfig.Security.Captcha.Enabled,
		CaptchaSiteKey: config.GlobalConfig.Security.Captcha.TurnstileSiteKey,
	})
}
