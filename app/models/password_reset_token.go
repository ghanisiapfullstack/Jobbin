package models

import "time"

type PasswordResetToken struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	TokenHash string     `gorm:"column:token_hash;size:64;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at;index" json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}
