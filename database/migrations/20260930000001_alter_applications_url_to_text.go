package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260930000001AlterApplicationsUrlToText struct{}

func (m *M20260930000001AlterApplicationsUrlToText) Signature() string {
	return "20260930000001_alter_applications_url_to_text"
}

// Up widens applications.url from VARCHAR(500) to TEXT. LinkedIn and other job
// board apply URLs frequently exceed 500 characters due to tracking query
// parameters, which previously caused inserts to fail with a 500 error.
func (m *M20260930000001AlterApplicationsUrlToText) Up() error {
	return facades.Schema().Table("applications", func(table schema.Blueprint) {
		table.Text("url").Nullable().Change()
	})
}

// Down narrows the column back to VARCHAR(500). This is only safe if no stored
// URL exceeds 500 characters; otherwise the rollback will fail.
func (m *M20260930000001AlterApplicationsUrlToText) Down() error {
	return facades.Schema().Table("applications", func(table schema.Blueprint) {
		table.String("url", 500).Nullable().Change()
	})
}
