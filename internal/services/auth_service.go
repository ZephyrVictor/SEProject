package services

import (
	"awesomeProject/internal/models"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CustomClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db             *gorm.DB
	rdb            *redis.Client
	jwtSecret      string
	jwtExpireHours int
}

func NewAuthService(db *gorm.DB, rdb *redis.Client, secret string, expireHr int) *AuthService {
	return &AuthService{
		db:             db,
		rdb:            rdb,
		jwtSecret:      secret,
		jwtExpireHours: expireHr,
	}
}

func (s *AuthService) Register(email, password string) (*models.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("用户名不存在")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", errors.New("密码错误")
	}
	// 生成 JWT
	exp := time.Now().Add(time.Duration(s.jwtExpireHours) * time.Hour)
	claims := CustomClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
	return token, err
}

func (s *AuthService) ParseToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (s *AuthService) VerifyEmailCode(email, code string) bool {
	storedCode, err := s.rdb.Get(context.Background(), "email_code:"+email).Result()
	if err != nil || storedCode != code {
		return false
	}

	s.rdb.Del(context.Background(), "email_code:"+email) // 删除验证码
	return true
}

func (a *AuthService) SendEmailCode(email string, code string, ttl time.Duration) error {
	// 将验证码存储到 Redis，设置过期时间
	return a.rdb.Set(context.Background(), "email_code:"+email, code, ttl).Err()
}

func (s *AuthService) DeleteEmailCode(email string) {
	s.rdb.Del(context.Background(), "email_code:"+email)
}

func (s *AuthService) GetJWTExpireHours() int {
	return s.jwtExpireHours
}
