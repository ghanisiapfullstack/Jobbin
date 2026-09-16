package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260916000002CreateRefreshSessionsTable struct{}

func (m *M20260916000002CreateRefreshSessionsTable) Signature() string {
	return "20260916000002_create_refresh_sessions_table"
}

func (m *M20260916000002CreateRefreshSessionsTable) Up() error {
	return facades.Schema().Create("refresh_sessions", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
		table.String("token_hash", 64)
		table.Unique("token_hash")
		table.Boolean("remember_me").Default(false)
		table.TimestampTz("expires_at")
		table.TimestampTz("revoked_at").Nullable()
		table.TimestampsTz()
		table.Index("user_id", "revoked_at")
		table.Index("expires_at")
	})
}

func (m *M20260916000002CreateRefreshSessionsTable) Down() error {
	return facades.Schema().DropIfExists("refresh_sessions")
}
