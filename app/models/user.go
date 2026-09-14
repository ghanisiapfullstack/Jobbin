package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type User struct {
	orm.Model
	Name               string           `gorm:"column:name;not null" json:"name"`
	Email              string           `gorm:"column:email;uniqueIndex;not null" json:"email"`
	Password           *string          `gorm:"column:password" json:"-"`
	GoogleID           *string          `gorm:"column:google_id;uniqueIndex" json:"-"`
	Avatar             *string          `gorm:"column:avatar" json:"avatar,omitempty"`
	EmailVerifiedAt    *carbon.DateTime `gorm:"column:email_verified_at" json:"email_verified_at"`
	EmailVerifyToken   *string          `gorm:"column:email_verify_token" json:"-"`
	EmailVerifyExpires *carbon.DateTime `gorm:"column:email_verify_expires" json:"-"`
	Applications       []Application    `gorm:"foreignKey:UserID" json:"applications,omitempty"`
}
