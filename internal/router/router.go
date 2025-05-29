package router

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"awesomeProject/config"
	"awesomeProject/internal/controllers"
	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// 初始化 DB & Redis
	models.InitMySQL(cfg)
	models.InitRedis(cfg)

	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal("Failed to get working directory", zap.Error(err))
	}
	logger.Info("Working directory", zap.String("path", wd))
	templatesPath := filepath.Join(wd, "templates")

	r := gin.New()
	r.Use(gin.Recovery())

	// Serve static HTML
	r.Static("/static", templatesPath)
	r.GET("/", func(c *gin.Context) { c.File(filepath.Join(templatesPath, "index.html")) })
	r.GET("/register.html", func(c *gin.Context) { c.File(filepath.Join(templatesPath, "register.html")) })
	r.GET("/login.html", func(c *gin.Context) { c.File(filepath.Join(templatesPath, "login.html")) })
	r.GET("/forgot.html", func(c *gin.Context) { c.File(filepath.Join(templatesPath, "forgot.html")) })
	r.GET("/reset.html", func(c *gin.Context) { c.File(filepath.Join(templatesPath, "reset.html")) })
	r.GET("/niloong", func(c *gin.Context) { c.String(200, "pong") })
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	// Services & Controllers
	authSvc := services.NewAuthService(models.DB, models.RDB, cfg.JWTSecret, cfg.JWTExpireHours)
	pwdSvc := services.NewPasswordService(cfg)

	authCtl := controllers.NewAuthController(authSvc, pwdSvc)

	// JSON API
	api := r.Group("/api")
	{
		api.POST("/register", authCtl.Register)
		api.POST("/login", authCtl.Login)
		api.POST("/password/forgot", authCtl.Forgot)
		api.POST("/password/reset", authCtl.Reset)
		api.POST("/send-email-code", authCtl.SendEmailCode)
		api.GET("/auth/status", authCtl.Status)
		api.POST("/auth/logout", authCtl.Logout)
	}

	return r
}
