package security

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Enabled  bool
	Secret   string
	Issuer   string
	Audience string
}

type TokenResult struct {
	Token        string
	ExpiresAt    time.Time
	ExpiresInSec int64
}

func GenerateJWT(cfg JWTConfig, subject string, claims map[string]any, ttlMinutes int) (TokenResult, error) {
	now := time.Now()
	exp := now.Add(time.Duration(ttlMinutes) * time.Minute)

	std := jwt.MapClaims{
		"iss": cfg.Issuer,
		"aud": cfg.Audience,
		"sub": subject,
		"iat": now.Unix(),
		"exp": exp.Unix(),
	}
	for k, v := range claims {
		std[k] = v
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, std)
	signed, err := t.SignedString([]byte(cfg.Secret))
	if err != nil {
		return TokenResult{}, err
	}
	return TokenResult{
		Token:        signed,
		ExpiresAt:    exp,
		ExpiresInSec: int64(time.Until(exp).Seconds()),
	}, nil
}

func MiddlewareJWT(cfg JWTConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		// No-op middleware kalau dimatikan
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
