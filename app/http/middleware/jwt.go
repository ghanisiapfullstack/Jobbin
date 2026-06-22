package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type Jwt struct{}

func NewJwtMiddleware() *Jwt {
	return &Jwt{}
}

func (m *Jwt) Handle() http.Middleware {
	return func(ctx http.Context) {
		_, err := facades.Auth(ctx).Parse(ctx.Request().Header("Authorization", ""))
		if err != nil {
			ctx.Request().AbortWithStatusJson(401, http.Json{
				"message": "Unauthorized",
			})
			return
		}
		ctx.Request().Next()
	}
}
