package controllers

import (
	"awesomeProject/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc     *services.AuthService
	passwordSvc *services.PasswordService
}

func NewAuthController(a *services.AuthService, p *services.PasswordService) *AuthController {
	return &AuthController{a, p}
}

func (ac *AuthController) Register(c *gin.Context) {
	var req struct{ Email, Password string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := ac.authSvc.Register(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "registered"})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req struct{ Email, Password string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := ac.authSvc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (ac *AuthController) Forgot(c *gin.Context) {
	var req struct{ Email string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ac.passwordSvc.Forgot(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

func (ac *AuthController) Reset(c *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ac.passwordSvc.Reset(c.Request.Context(), req.Email, req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reset"})
}
