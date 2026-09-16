package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"jobbin/backend/app/models"
)

const RefreshCookieName = "jobbin_refresh"

var ErrInvalidRefreshSession = errors.New("invalid refresh session")

type SessionService struct{}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func (s *SessionService) Create(userID uint, rememberMe bool) (models.RefreshSession, string, error) {
	rawToken, tokenHash, err := newRefreshToken()
	if err != nil {
		return models.RefreshSession{}, "", err
	}

	session := models.RefreshSession{
		UserID:     userID,
		TokenHash:  tokenHash,
		RememberMe: rememberMe,
		ExpiresAt:  refreshExpiry(rememberMe),
	}
	if err := facades.Orm().Query().Create(&session); err != nil {
		return models.RefreshSession{}, "", err
	}
	return session, rawToken, nil
}

func (s *SessionService) Rotate(rawToken string) (models.User, models.RefreshSession, string, error) {
	if rawToken == "" {
		return models.User{}, models.RefreshSession{}, "", ErrInvalidRefreshSession
	}

	var user models.User
	var nextSession models.RefreshSession
	var nextRawToken string
	err := facades.Orm().Transaction(func(tx contractsorm.Query) error {
		var current models.RefreshSession
		if err := tx.LockForUpdate().Where("token_hash", hashRefreshToken(rawToken)).First(&current); err != nil {
			return ErrInvalidRefreshSession
		}
		if current.ID == 0 || current.RevokedAt != nil || !current.ExpiresAt.After(time.Now()) {
			return ErrInvalidRefreshSession
		}
		if err := tx.Find(&user, current.UserID); err != nil || user.ID == 0 {
			return ErrInvalidRefreshSession
		}

		rotatedRaw, rotatedHash, err := newRefreshToken()
		if err != nil {
			return err
		}
		now := time.Now()
		current.RevokedAt = &now
		if err := tx.Save(&current); err != nil {
			return err
		}

		nextSession = models.RefreshSession{
			UserID:     current.UserID,
			TokenHash:  rotatedHash,
			RememberMe: current.RememberMe,
			ExpiresAt:  refreshExpiry(current.RememberMe),
		}
		if err := tx.Create(&nextSession); err != nil {
			return err
		}
		nextRawToken = rotatedRaw
		return nil
	})
	if err != nil {
		return models.User{}, models.RefreshSession{}, "", err
	}
	return user, nextSession, nextRawToken, nil
}

func (s *SessionService) Revoke(rawToken string) {
	if rawToken == "" {
		return
	}
	now := time.Now()
	_, _ = facades.Orm().Query().Model(&models.RefreshSession{}).
		Where("token_hash", hashRefreshToken(rawToken)).
		WhereNull("revoked_at").
		Update("revoked_at", now)
}

func (s *SessionService) RevokeUser(userID uint) {
	now := time.Now()
	_, _ = facades.Orm().Query().Model(&models.RefreshSession{}).
		Where("user_id", userID).
		WhereNull("revoked_at").
		Update("revoked_at", now)
}

func (s *SessionService) CleanupExpired() error {
	_, err := facades.Orm().Query().
		Where("expires_at < ? OR revoked_at < ?", time.Now(), time.Now().AddDate(0, 0, -7)).
		ForceDelete(&models.RefreshSession{})
	return err
}

func RefreshCookie(rawToken string, session models.RefreshSession) http.Cookie {
	maxAge := 0
	expires := time.Time{}
	if session.RememberMe {
		maxAge = int(time.Until(session.ExpiresAt).Seconds())
		expires = session.ExpiresAt
	}
	return http.Cookie{
		Name:     RefreshCookieName,
		Value:    rawToken,
		Path:     "/api/v1/auth",
		Expires:  expires,
		MaxAge:   maxAge,
		Secure:   facades.Config().GetString("app.env") == "production",
		HttpOnly: true,
		SameSite: "Lax",
	}
}

func ExpiredRefreshCookie() http.Cookie {
	return http.Cookie{
		Name:     RefreshCookieName,
		Path:     "/api/v1/auth",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		Secure:   facades.Config().GetString("app.env") == "production",
		HttpOnly: true,
		SameSite: "Lax",
	}
}

func newRefreshToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(bytes)
	return rawToken, hashRefreshToken(rawToken), nil
}

func hashRefreshToken(rawToken string) string {
	digest := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(digest[:])
}

func refreshExpiry(rememberMe bool) time.Time {
	if rememberMe {
		return time.Now().Add(30 * 24 * time.Hour)
	}
	return time.Now().Add(24 * time.Hour)
}
