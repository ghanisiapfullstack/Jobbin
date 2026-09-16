package models

import "time"

type RefreshSession struct {
	ID         uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User       User       `gorm:"foreignKey:UserID" json:"-"`
	TokenHash  string     `gorm:"column:token_hash;size:64;not null;uniqueIndex" json:"-"`
	RememberMe bool       `gorm:"column:remember_me;not null;default:false" json:"remember_me"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	RevokedAt  *time.Time `gorm:"column:revoked_at;index" json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
