package main

import (
	"awesomeProject/config"
	"awesomeProject/internal/router"
	"awesomeProject/pkg/utils"

	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	logger := utils.InitLogger()
	defer logger.Sync()

	r := router.SetupRouter(cfg, logger)
	logger.Info("server running", zap.String("addr", cfg.ServerAddr))
	r.Run(cfg.ServerAddr)
}
