package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"jobbin/backend/app/models"
)

var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

type PasswordResetService struct{}

func NewPasswordResetService() *PasswordResetService {
	return &PasswordResetService{}
}

func (s *PasswordResetService) Create(userID uint) (string, error) {
	raw, hash, err := newPasswordResetToken()
	if err != nil {
		return "", err
	}

	now := time.Now()
	_, _ = facades.Orm().Query().Model(&models.PasswordResetToken{}).
		Where("user_id", userID).
		WhereNull("used_at").
		Update("used_at", now)

	token := models.PasswordResetToken{
		UserID: userID, TokenHash: hash, ExpiresAt: now.Add(30 * time.Minute),
	}
	if err := facades.Orm().Query().Create(&token); err != nil {
		return "", err
	}
	return raw, nil
}

func (s *PasswordResetService) Consume(rawToken, hashedPassword string) (models.User, error) {
	var user models.User
	err := facades.Orm().Transaction(func(tx contractsorm.Query) error {
		var token models.PasswordResetToken
		if err := tx.LockForUpdate().Where("token_hash", hashPasswordResetToken(rawToken)).First(&token); err != nil {
			return ErrInvalidPasswordResetToken
		}
		if token.ID == 0 || token.UsedAt != nil || !token.ExpiresAt.After(time.Now()) {
			return ErrInvalidPasswordResetToken
		}
		if err := tx.Find(&user, token.UserID); err != nil || user.ID == 0 {
			return ErrInvalidPasswordResetToken
		}

		user.Password = &hashedPassword
		if err := tx.Save(&user); err != nil {
			return err
		}
		now := time.Now()
		token.UsedAt = &now
		return tx.Save(&token)
	})
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func newPasswordResetToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(bytes)
	return raw, hashPasswordResetToken(raw), nil
}

func hashPasswordResetToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}
