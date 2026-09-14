package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260707000001AlterUsersAddGoogleFields struct{}

func (m *M20260707000001AlterUsersAddGoogleFields) Signature() string {
	return "20260707000001_alter_users_add_google_fields"
}

func (m *M20260707000001AlterUsersAddGoogleFields) Up() error {
	return facades.Schema().Table("users", func(table schema.Blueprint) {
		table.String("google_id", 255).Nullable()
		table.String("avatar", 500).Nullable()
		table.Text("password").Nullable().Change()
	})
}

func (m *M20260707000001AlterUsersAddGoogleFields) Down() error {
	return facades.Schema().Table("users", func(table schema.Blueprint) {
		table.DropColumn("google_id")
		table.DropColumn("avatar")
	})
}
