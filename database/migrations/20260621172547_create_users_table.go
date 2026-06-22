package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"jobbin/backend/app/facades"
)

type M20260621172547CreateUsersTable struct{}

func (r *M20260621172547CreateUsersTable) Signature() string {
	return "20260621172547_create_users_table"
}

func (r *M20260621172547CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID()
			table.String("name", 100)
			table.String("email", 255)
			table.Unique("email")
			table.String("password", 255)
			table.Timestamp("email_verified_at").Nullable()
			table.String("email_verify_token", 255).Nullable()
			table.Timestamp("email_verify_expires").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260621172547CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
