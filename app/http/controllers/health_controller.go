package controllers

import (
	"context"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type HealthController struct {
	mu        sync.Mutex
	checkedAt time.Time
	ready     bool
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (r *HealthController) Live(ctx http.Context) http.Response {
	return ctx.Response().Json(200, http.Json{"status": "ok"})
}

func (r *HealthController) Ready(ctx http.Context) http.Response {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Azure may probe frequently. Cache briefly so readiness does not become a
	// database load source while still detecting outages quickly.
	if time.Since(r.checkedAt) < 5*time.Second {
		if r.ready {
			return ctx.Response().Json(200, http.Json{"status": "ready"})
		}
		return ctx.Response().Json(503, http.Json{"status": "unavailable"})
	}

	db, err := facades.Orm().Query().DB()
	if err != nil {
		r.checkedAt = time.Now()
		r.ready = false
		return ctx.Response().Json(503, http.Json{"status": "unavailable"})
	}

	pingCtx, cancel := context.WithTimeout(ctx.Request().Origin().Context(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		facades.Log().Warningf("Readiness database ping failed: %v", err)
		r.checkedAt = time.Now()
		r.ready = false
		return ctx.Response().Json(503, http.Json{"status": "unavailable"})
	}

	r.checkedAt = time.Now()
	r.ready = true
	return ctx.Response().Json(200, http.Json{"status": "ready"})
}
