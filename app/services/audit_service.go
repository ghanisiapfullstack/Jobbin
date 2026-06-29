package services

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"jobbin/backend/app/models"
)

// Action constants
const (
	ActionLogin          = "LOGIN"
	ActionLogout         = "LOGOUT"
	ActionCreateApp      = "CREATE_APP"
	ActionUpdateApp      = "UPDATE_APP"
	ActionDeleteApp      = "DELETE_APP"
	ActionArchiveApp     = "ARCHIVE_APP"
	ActionRestoreApp     = "RESTORE_APP"
	ActionChangePassword = "CHANGE_PASSWORD"
	ActionUpdateProfile  = "UPDATE_PROFILE"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

// Log catat audit log ke database
func (s *AuditService) Log(ctx http.Context, userID *uint, action string, resourceID *uint) {
	ip := ctx.Request().Ip()
	ua := ctx.Request().Header("User-Agent")

	log := models.AuditLog{
		UserID:     userID,
		Action:     action,
		ResourceID: resourceID,
		IPAddress:  &ip,
		UserAgent:  &ua,
	}

	// Silent fail — audit log tidak boleh ganggu main flow
	_ = facades.Orm().Query().Create(&log)
}

// CleanupOldLogs hapus log lebih dari 90 hari
func (s *AuditService) CleanupOldLogs() error {
	_, err := facades.Orm().Query().
		Where("created_at < NOW() - INTERVAL '90 days'").
		Delete(&models.AuditLog{})
	return err
}
