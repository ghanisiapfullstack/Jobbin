package config

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/driver"
	postgresfacades "github.com/goravel/postgres/facades"
	"jobbin/backend/app/facades"
)

func init() {
	config := facades.Config()

	dbHost := config.Env("DB_HOST", "127.0.0.1")
	dbPort := config.Env("DB_PORT", "5432")
	dbUser := config.Env("DB_USERNAME", "jobbin")
	dbPass := config.Env("DB_PASSWORD", "jobbin_secret")
	dbName := config.Env("DB_DATABASE", "jobbin")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName,
	)

	config.Add("database", map[string]any{
		"default": "postgres",
		"connections": map[string]any{
			"postgres": map[string]any{
				"host":     dbHost,
				"port":     dbPort,
				"database": dbName,
				"username": dbUser,
				"password": dbPass,
				"sslmode":  "disable",
				"singular": false,
				"prefix":   "",
				"schema":   config.Env("DB_SCHEMA", "public"),
				"dsn":      dsn,
				"via": func() (driver.Driver, error) {
					return postgresfacades.Postgres("postgres")
				},
			},
		},
		"pool": map[string]any{
			"max_idle_conns":    10,
			"max_open_conns":    100,
			"conn_max_idletime": 3600,
			"conn_max_lifetime": 3600,
		},
		"slow_threshold": 200,
		"migrations": map[string]any{
			"table": "migrations",
		},
	})
}
