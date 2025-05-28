package services

import (
	"awesomeProject/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db             *gorm.DB
	jwtSecret      string
	jwtExpireHours int
}

func NewAuthService(db *gorm.DB, secret string, expireHr int) *AuthService {
	return &AuthService{db, secret, expireHr}
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
		return "", errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * time.Duration(s.jwtExpireHours)).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}
