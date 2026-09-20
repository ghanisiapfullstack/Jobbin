package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// internalError keeps implementation details in server logs while returning a
// stable, supportable error contract to API clients.
func internalError(ctx http.Context, message, code string, err error) http.Response {
	if err != nil {
		facades.Log().Errorf("%s: %v", code, err)
	}

	return ctx.Response().Json(500, http.Json{
		"message": message,
		"code":    code,
	})
}
