package controllers

import (
	"awesomeProject/internal/services"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc     *services.AuthService
	passwordSvc *services.PasswordService
}

func NewAuthController(a *services.AuthService, p *services.PasswordService) *AuthController {
	return &AuthController{
		authSvc:     a,
		passwordSvc: p,
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	// 应该有邮箱验证码
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !ac.authSvc.VerifyEmailCode(req.Email, req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
		return
	}

	// 注册用户
	_, err := ac.authSvc.Register(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "registered"})
}

func (ac *AuthController) Login(c *gin.Context) {
	// 按前端 fetch/json 提交改写，不再渲染 HTML
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ac.authSvc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 写入 Cookie
	expireSec := ac.authSvc.GetJWTExpireHours() * 3600
	c.SetCookie("token", token, expireSec, "/", "", false, true)

	// 返回 JSON，前端收到后 location.href = '/'
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (ac *AuthController) Status(c *gin.Context) {
	t, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}
	claims, err := ac.authSvc.ParseToken(t)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"email":         claims.Email,
	})
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

func (ac *AuthController) SendEmailCode(c *gin.Context) {
	var req struct{ Email string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成随机验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 存储验证码到 Redis
	err := ac.authSvc.SendEmailCode(req.Email, code, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码存储失败"})
		return
	}

	// 发送验证码邮件
	err = ac.passwordSvc.SendVerificationEmail(req.Email, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码发送失败"})
		// 如果邮件发送失败，可以选择删除 Redis 中的验证码
		ac.authSvc.DeleteEmailCode(req.Email)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "验证码已发送"})
}

func (ac *AuthController) Logout(c *gin.Context) {
	fmt.Print("Logout called\n")
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}
