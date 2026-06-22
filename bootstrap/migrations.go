package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"jobbin/backend/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20260621172547CreateUsersTable{},
		&migrations.M20260621172548CreateApplicationsTable{},
	}
}
