package console

import (
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"jobbin/backend/app/services"
)

type SendRemindersCommand struct{}

func NewSendRemindersCommand() *SendRemindersCommand {
	return &SendRemindersCommand{}
}

func (r *SendRemindersCommand) Signature() string {
	return "reminders:send"
}

func (r *SendRemindersCommand) Description() string {
	return "Send due Jobbin reminder emails"
}

func (r *SendRemindersCommand) Extend() command.Extend {
	return command.Extend{Category: "jobbin"}
}

func (r *SendRemindersCommand) Handle(ctx contractsconsole.Context) error {
	if err := services.NewReminderService().SendDailyReminders(); err != nil {
		ctx.Error("One or more reminders could not be processed")
		return err
	}

	ctx.Success("Due reminders processed")
	return nil
}
