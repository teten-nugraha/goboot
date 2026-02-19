package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teten-nugraha/goboot/internal/domain"
	"gorm.io/gorm"
)

func MountUserProtected(g *gin.RouterGroup, db *gorm.DB) {
	g.GET("/users", func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DB disabled"})
			return
		}
		var users []domain.User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})
}
