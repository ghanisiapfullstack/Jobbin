package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260919000001AddApplicationJobDetails struct{}

func (m *M20260919000001AddApplicationJobDetails) Signature() string {
	return "20260919000001_add_application_job_details"
}

func (m *M20260919000001AddApplicationJobDetails) Up() error {
	return facades.Schema().Table("applications", func(table schema.Blueprint) {
		table.String("employment_type", 20).Nullable()
		table.BigInteger("salary_min").Nullable()
		table.BigInteger("salary_max").Nullable()
	})
}

func (m *M20260919000001AddApplicationJobDetails) Down() error {
	return facades.Schema().Table("applications", func(table schema.Blueprint) {
		table.DropColumn("employment_type", "salary_min", "salary_max")
	})
}
