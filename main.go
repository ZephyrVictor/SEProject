package main

import (
	"awesomeProject/config"
	"awesomeProject/internal"
	"awesomeProject/internal/models"
	"awesomeProject/internal/router"
	"awesomeProject/internal/services"
	"awesomeProject/pkg/utils"
	"context"

	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	logger := utils.InitLogger()

	defer logger.Sync()

	r := router.SetupRouter(cfg, logger)
	logger.Info("server running", zap.String("addr", cfg.ServerAddr))

	// 启动后台图像处理 Worker，使用已初始化的 models.DB
	go internal.StartImageWorker(context.Background(),
		services.NewDashScopeClient(cfg.DASHSCOPE_API_KEY, cfg.DASHSCOPE_BASE_URL),
		services.NewTaskQueue(models.RDB),
		models.DB)

	r.Run(cfg.ServerAddr)
}
