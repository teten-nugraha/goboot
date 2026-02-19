// internal/controller/actuator.go
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MountActuator(r *gin.Engine) {
	r.GET("/actuator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
}
