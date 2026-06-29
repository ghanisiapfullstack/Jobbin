package middleware

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type Cors struct{}

func NewCorsMiddleware() *Cors {
	return &Cors{}
}

func (m *Cors) Handle() http.Middleware {
	return func(ctx http.Context) {
		allowedOrigins := facades.Config().Get("cors.allowed_origins").([]string)
		origin := ctx.Request().Header("Origin")

		// Tentukan apakah origin diizinkan
		allowedOrigin := ""
		for _, o := range allowedOrigins {
			if o == "*" {
				allowedOrigin = "*"
				break
			}
			if strings.EqualFold(o, origin) {
				allowedOrigin = origin
				break
			}
		}

		if allowedOrigin != "" {
			ctx.Response().Header("Access-Control-Allow-Origin", allowedOrigin)
			ctx.Response().Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			ctx.Response().Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
			ctx.Response().Header("Access-Control-Max-Age", "86400")
			ctx.Response().Header("Vary", "Origin")
		}

		// Handle preflight OPTIONS request
		if ctx.Request().Method() == "OPTIONS" {
			ctx.Request().AbortWithStatusJson(204, nil)
			return
		}

		ctx.Request().Next()
	}
}
