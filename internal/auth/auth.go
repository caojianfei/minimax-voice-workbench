package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"minimax-voice-workbench/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

// JWT Secret - in production this should be a secure random key
var jwtSecret = []byte("minimax-voice-workbench-secret-key")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateTokens returns access and refresh tokens
func GenerateTokens(username string) (string, string, error) {
	// Access Token (15 mins)
	accessClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "minimax-voice-workbench",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (7 days)
	refreshClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "minimax-voice-workbench",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	return signedAccess, signedRefresh, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// CheckCredentials verifies the username and password
func CheckCredentials(username, password string) bool {
	// Simple plaintext comparison as per current config structure
	adminUser := config.GlobalConfig.Security.Admin.Username
	adminPass := config.GlobalConfig.Security.Admin.Password

	// Fallback if not set
	if adminUser == "" && adminPass == "" {
		return false
	}

	return username == adminUser && password == adminPass
}

// Middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GlobalConfig.Security.Enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

// Rate Limiter
// Simple IP-based rate limiter
var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// 5 requests per minute = 1 request every 12 seconds, burst 5
		// rate.Limit is requests per second.
		limiter = rate.NewLimiter(rate.Limit(5.0/60.0), 5)
		visitors[ip] = limiter
	}

	return limiter
}

func LoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GlobalConfig.Security.Captcha.Enabled {
			c.Next()
			return
		}

		limiter := getVisitor(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
			return
		}
		c.Next()
	}
}

type TurnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
	CData       string   `json:"cdata,omitempty"`
}

// VerifyTurnstileToken verifies a Cloudflare Turnstile token with the configured secret key.
func VerifyTurnstileToken(ctx context.Context, responseToken string, remoteIP string) (*TurnstileVerifyResponse, error) {
	secret := config.GlobalConfig.Security.Captcha.TurnstileSecretKey
	if secret == "" {
		return nil, errors.New("turnstile secret key is not configured")
	}
	if responseToken == "" {
		return nil, errors.New("turnstile response token is required")
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", responseToken)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("turnstile verify failed with status %s", resp.Status)
	}

	var result TurnstileVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
