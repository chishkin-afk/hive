package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired — middleware, который проверяет наличие куки "token"
// и редиректит на /auth, если её нет.
// Публичные страницы (лэндинг, auth, reset, документы) должны
// использовать этот middleware с параметром skipAuth=true.
func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusFound, "/auth")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// RedirectIfAuthenticated — middleware, который проверяет наличие куки "token"
// и редиректит на /services, если она есть.
// Используется на странице /auth, чтобы залогиненные пользователи не видели форму входа.
func RedirectIfAuthenticated() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := ctx.Cookie("token")
		if err == nil {
			ctx.Redirect(http.StatusFound, "/services")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
