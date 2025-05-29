package router

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"awesomeProject/config"
	"awesomeProject/internal"
	"awesomeProject/internal/controllers"
	"awesomeProject/internal/middleware"
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
	// 加载模板，用于渲染 HTML 页面
	r.LoadHTMLGlob(filepath.Join(templatesPath, "*.html"))

	// Serve static files (CSS, JS, images, uploads)
	// 这里将 /static 指向项目根目录下的 static 文件夹，用于提供上传图片等资源
	r.Static("/static", filepath.Join(wd, "static"))
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

	api := r.Group("/api")
	{
		api.POST("/register", authCtl.Register)
		api.POST("/login", authCtl.Login)
		api.POST("/password/forgot", authCtl.Forgot)
		api.POST("/password/reset", authCtl.Reset)
		api.POST("/send-email-code", authCtl.SendEmailCode)
		api.GET("/auth/status", authCtl.Status)
		api.POST("/auth/logout", authCtl.Logout)
		// 获取当前用户的图片列表（JSON）
		imgQueue := services.NewTaskQueue(models.RDB)
		imgCtl := controllers.NewImageController(imgQueue, models.DB, cfg)
		api.GET("/images", imgCtl.ListJSON)
	}
	// 以下路由需要登录后访问
	imgQueue := services.NewTaskQueue(models.RDB)
	imgCtl := controllers.NewImageController(imgQueue, models.DB, cfg)
	authGroup := r.Group("/")
	authGroup.Use(middleware.JWTAuth(cfg))
	{
		// 图像编辑页
		authGroup.GET("/edit.html", imgCtl.ShowEdit)
		// 提交编辑任务
		authGroup.POST("/api/image/edit", imgCtl.SubmitEdit)
		// 查看历史列表
		authGroup.GET("/history.html", imgCtl.ShowHistory)
		// 分支详情
		authGroup.GET("/history/:task_id", imgCtl.ShowBranch)
		// WebSocket 连接
		authGroup.GET("/ws", internal.HandleWs)
	}

	return r
}
