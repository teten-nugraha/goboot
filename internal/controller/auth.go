// internal/controller/auth.go (potongan register & login)
package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teten-nugraha/goboot/pkg/data"
	"github.com/teten-nugraha/goboot/pkg/web"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/teten-nugraha/goboot/internal/dto"
	"github.com/teten-nugraha/goboot/internal/service"
	"github.com/teten-nugraha/goboot/pkg/security"
)

func MountAuth(r *gin.Engine, db *gorm.DB, secCfg security.JWTConfig, ttlMin int, tx data.TxManager, logger *zap.Logger) {
	svc := service.NewAuthService(db, secCfg, ttlMin, tx, logger)

	r.POST("/auth/register", func(c *gin.Context) {
		var req dto.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			web.BadRequest(c, "invalid payload", err.Error())
			return
		}
		u, err := svc.Register(c.Request.Context(), req.Email, req.Name, req.Password)
		if err != nil {
			web.BadRequest(c, "User already registered", err.Error())
			return
		}
		tok, err := svc.IssueToken(u)
		if err != nil {
			web.BadRequest(c, "failed to issue token", err.Error())
			return
		}
		expiresIn := int64(time.Until(tok.ExpiresAt).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}

		resp := dto.AuthResponse{
			AccessToken: tok.Token,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
		}
		resp.User.ID = u.ID
		resp.User.Email = u.Email
		resp.User.Name = u.Name

		web.Created(c, resp)

	})

	r.POST("/auth/login", func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		u, err := svc.Authenticate(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		tok, err := svc.IssueToken(u)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}
		expiresIn := int64(time.Until(tok.ExpiresAt).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}

		resp := dto.AuthResponse{
			AccessToken: tok.Token,
			TokenType:   "Bearer",
			ExpiresIn:   expiresIn,
		}
		resp.User.ID = u.ID
		resp.User.Email = u.Email
		resp.User.Name = u.Name

		web.OK(c, resp)

	})
}
