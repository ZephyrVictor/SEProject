package router

import (
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

	r := gin.New()
	r.Use(gin.Recovery())

	// Serve static HTML
	r.Static("/static", "./public")
	r.GET("/", func(c *gin.Context) { c.File("./public/login.html") })
	r.GET("/register.html", func(c *gin.Context) { c.File("./public/register.html") })
	r.GET("/login.html", func(c *gin.Context) { c.File("./public/login.html") })
	r.GET("/forgot.html", func(c *gin.Context) { c.File("./public/forgot.html") })
	r.GET("/reset.html", func(c *gin.Context) { c.File("./public/reset.html") })

	// Services & Controllers
	authSvc := services.NewAuthService(models.DB, cfg.JWTSecret, cfg.JWTExpireHours)
	pwdSvc := services.NewPasswordService(cfg)
	authCtl := controllers.NewAuthController(authSvc, pwdSvc)

	// JSON API
	api := r.Group("/api")
	{
		api.POST("/register", authCtl.Register)
		api.POST("/login", authCtl.Login)
		api.POST("/password/forgot", authCtl.Forgot)
		api.POST("/password/reset", authCtl.Reset)
	}

	return r
}
