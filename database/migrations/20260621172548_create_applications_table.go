package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"jobbin/backend/app/facades"
)

type M20260621172548CreateApplicationsTable struct{}

func (r *M20260621172548CreateApplicationsTable) Signature() string {
	return "20260621172548_create_applications_table"
}

func (r *M20260621172548CreateApplicationsTable) Up() error {
	if !facades.Schema().HasTable("applications") {
		return facades.Schema().Create("applications", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.Foreign("user_id").References("id").On("users")
			table.String("job_title", 255)
			table.String("company", 255)
			table.String("url", 500).Nullable()
			table.String("status", 20).Default("wishlist")
			table.Text("notes").Nullable()
			table.Date("applied_date").Nullable()
			table.Date("reminder_date").Nullable()
			table.Boolean("reminder_sent_day_before").Default(false)
			table.Boolean("reminder_sent_day_of").Default(false)
			table.Boolean("is_archived").Default(false)
			table.Float("position", 53).Default(0)
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260621172548CreateApplicationsTable) Down() error {
	return facades.Schema().DropIfExists("applications")
}
