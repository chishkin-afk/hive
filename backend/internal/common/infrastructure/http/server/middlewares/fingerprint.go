package middlewares

import (
	"context"

	contextvalue "github.com/devathh/hive/backend/internal/common/context"
	"github.com/gin-gonic/gin"
)

func NewFingerprint() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fp := ctx.GetHeader("X-Fingerprint")

		cv := context.WithValue(ctx.Request.Context(), contextvalue.KeyFingerprint, fp)
		ctx.Request = ctx.Request.WithContext(cv)

		ctx.Next()
	}
}
