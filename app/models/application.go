package models

import (
	"github.com/goravel/framework/database/orm"
)

type Application struct {
	orm.Model
	UserID                 uint         `gorm:"column:user_id;not null" json:"user_id"`
	User                   User         `gorm:"foreignKey:UserID" json:"-"`
	JobTitle               string       `gorm:"column:job_title;not null" json:"job_title"`
	Company                string       `gorm:"column:company;not null" json:"company"`
	URL                    *string      `gorm:"column:url" json:"url"`
	Status                 string       `gorm:"column:status;default:wishlist" json:"status"`
	Notes                  *string      `gorm:"column:notes" json:"notes"`
	AppliedDate            *string      `gorm:"column:applied_date" json:"applied_date"`
	ReminderDate           *string      `gorm:"column:reminder_date" json:"reminder_date"`
	ReminderSentDayBefore  bool         `gorm:"column:reminder_sent_day_before;default:false" json:"reminder_sent_day_before"`
	ReminderSentDayOf      bool         `gorm:"column:reminder_sent_day_of;default:false" json:"reminder_sent_day_of"`
	IsArchived             bool         `gorm:"column:is_archived;default:false" json:"is_archived"`
	Position               float64      `gorm:"column:position;default:0" json:"position"`
}
