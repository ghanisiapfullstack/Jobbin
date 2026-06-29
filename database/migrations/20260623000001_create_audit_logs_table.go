package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260623000001CreateAuditLogsTable struct{}

func (m *M20260623000001CreateAuditLogsTable) Signature() string {
	return "20260623000001_create_audit_logs_table"
}

func (m *M20260623000001CreateAuditLogsTable) Up() error {
	return facades.Schema().Create("audit_logs", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id").Nullable()
		table.String("action", 50)
		table.UnsignedBigInteger("resource_id").Nullable()
		table.String("ip_address", 45).Nullable()
		table.String("user_agent", 255).Nullable()
		table.Timestamp("created_at").UseCurrent()

		table.Index("user_id")
		table.Index("action")
		table.Index("created_at")
	})
}

func (m *M20260623000001CreateAuditLogsTable) Down() error {
	return facades.Schema().DropIfExists("audit_logs")
}
