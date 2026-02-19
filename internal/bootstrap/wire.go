package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/teten-nugraha/goboot/pkg/data"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/teten-nugraha/goboot/internal/controller"
	"github.com/teten-nugraha/goboot/pkg/security"
)

// internal/bootstrap/wire.go (potongan)
func WireRoutes(r *gin.Engine, db *gorm.DB, secCfg security.JWTConfig, jwtTTLMin int, tx data.TxManager, logger *zap.Logger) {
	// Public
	controller.MountActuator(r)
	controller.MountAuth(r, db, secCfg, jwtTTLMin, tx, logger)

	// Protected
	protected := r.Group("/api")
	protected.Use(security.MiddlewareJWT(secCfg))
	controller.MountUserProtected(protected, db)
}
