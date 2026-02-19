package web

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceId := uuid.New().String()
		c.Set("traceId", traceId)
		c.Writer.Header().Set("X-Trace-Id", traceId)
		c.Next()
	}
}
