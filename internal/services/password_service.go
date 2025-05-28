package services

import (
	"awesomeProject/config"
	"awesomeProject/internal/models"
	"awesomeProject/internal/services/mailer"
	"context"
	"crypto/rand"

	"encoding/hex"
	"errors"
	"time"

	//"awesomeProject/services/mailer"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PasswordService struct {
	db       *gorm.DB
	rdb      *redis.Client
	mailer   *mailer.SimpleMailer
	tokenTTL time.Duration
}

func NewPasswordService(cfg *config.Config) *PasswordService {
	mailConf := mailer.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPass,
	}
	return &PasswordService{
		db:       models.DB,
		rdb:      models.RDB,
		mailer:   mailer.NewSimpleMailer(mailConf, cfg.BaseUrl),
		tokenTTL: time.Duration(cfg.ResetTokenTTLMinutes) * time.Minute,
	}
}

func (s *PasswordService) Forgot(ctx context.Context, email string) error {
	var u models.User
	if err := s.db.Where("email = ?", email).First(&u).Error; err != nil {
		return errors.New("email not registered")
	}

	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	key := "pwd_reset:" + email
	if err := s.rdb.Set(ctx, key, token, s.tokenTTL).Err(); err != nil {
		return err
	}
	return s.mailer.SendResetEmail(email, token, s.tokenTTL)
}

func (s *PasswordService) Reset(ctx context.Context, email, token, newPwd string) error {
	key := "pwd_reset:" + email
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil || val != token {
		return errors.New("申请无效或已经过期")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err := s.db.Model(&models.User{}).
		Where("email = ?", email).
		Update("password_hash", string(hash)).Error; err != nil {
		return err
	}
	s.rdb.Del(ctx, key)
	return nil
}

func (s *PasswordService) SendVerificationEmail(email, code string) error {
	return s.mailer.SendVerificationEmail(email, code)
}
