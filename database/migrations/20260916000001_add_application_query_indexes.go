package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260916000001AddApplicationQueryIndexes struct{}

func (m *M20260916000001AddApplicationQueryIndexes) Signature() string {
	return "20260916000001_add_application_query_indexes"
}

func (m *M20260916000001AddApplicationQueryIndexes) Up() error {
	return facades.Schema().Table("applications", func(table schema.Blueprint) {
		table.Index("user_id", "is_archived", "status", "position")
		table.Index("reminder_date", "is_archived")
	})
}

func (m *M20260916000001AddApplicationQueryIndexes) Down() error {
	return facades.Schema().Table("applications", func(table schema.Blueprint) {
		table.DropIndex("user_id", "is_archived", "status", "position")
		table.DropIndex("reminder_date", "is_archived")
	})
}
