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
	// 从表单获取
	email := c.PostForm("email")
	password := c.PostForm("password")
	token, err := ac.authSvc.Login(email, password)
	if err != nil {
		// 登录失败重回登录页并带错误信息
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": err.Error()})
		return
	}
	// 写入 Cookie 并重定向到首页
	expireSec := ac.authSvc.GetJWTExpireHours() * 3600
	c.SetCookie("token", token, expireSec, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
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
