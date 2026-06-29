package models

import "time"

type AuditLog struct {
	ID         uint       `json:"id"`
	UserID     *uint      `json:"user_id"`
	Action     string     `json:"action"`
	ResourceID *uint      `json:"resource_id"`
	IPAddress  *string    `json:"ip_address"`
	UserAgent  *string    `json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at"`
}
