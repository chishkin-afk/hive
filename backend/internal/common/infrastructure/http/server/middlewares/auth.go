package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	contextvalue "github.com/devathh/hive/backend/internal/common/context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type jwtManager interface {
	Validate(tokenString string) (uuid.UUID, error)
}

func NewAuthMiddleware(jm jwtManager, authRequire map[string]bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !authRequire[fmt.Sprintf("%s:%s", ctx.Request.Method, ctx.FullPath())] {
			ctx.Next()
			return
		}

		token, _ := strings.CutPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid authorization's header",
			})
			return
		}

		userID, err := jm.Validate(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &responses.ErrMsg{
				Error: "invalid token",
			})
			return
		}

		ctxValue := context.WithValue(ctx.Request.Context(), contextvalue.KeyUserID, userID)
		ctx.Request = ctx.Request.WithContext(ctxValue)

		ctx.Next()
	}
}
