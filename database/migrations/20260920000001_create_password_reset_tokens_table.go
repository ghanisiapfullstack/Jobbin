package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260920000001CreatePasswordResetTokensTable struct{}

func (m *M20260920000001CreatePasswordResetTokensTable) Signature() string {
	return "20260920000001_create_password_reset_tokens_table"
}

func (m *M20260920000001CreatePasswordResetTokensTable) Up() error {
	return facades.Schema().Create("password_reset_tokens", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
		table.String("token_hash", 64)
		table.Unique("token_hash")
		table.TimestampTz("expires_at")
		table.TimestampTz("used_at").Nullable()
		table.TimestampTz("created_at")
		table.Index("user_id", "used_at")
		table.Index("expires_at")
	})
}

func (m *M20260920000001CreatePasswordResetTokensTable) Down() error {
	return facades.Schema().DropIfExists("password_reset_tokens")
}
