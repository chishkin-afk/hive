package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Real-IP") != "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}
