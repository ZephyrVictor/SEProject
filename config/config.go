// config/config.go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServerAddr           string
	MySQLUser            string
	MySQLPassword        string
	MySQLHost            string
	MySQLPort            string
	MySQLDB              string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	JWTSecret            string
	JWTExpireHours       int
	DASHSCOPE_API_KEY    string // 通义
	DASHSCOPE_BASE_URL   string
	SMTPHost             string
	SMTPPort             int
	SMTPUser             string
	SMTPPass             string
	ResetTokenTTLMinutes int
	BaseUrl              string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("配置文件读取失败: %w", err))
	}
	return &Config{
		ServerAddr:           viper.GetString("SERVER_ADDR"),
		MySQLUser:            viper.GetString("MYSQL_USER"),
		MySQLPassword:        viper.GetString("MYSQL_PASSWORD"),
		MySQLHost:            viper.GetString("MYSQL_HOST"),
		MySQLPort:            viper.GetString("MYSQL_PORT"),
		MySQLDB:              viper.GetString("MYSQL_DB"),
		RedisAddr:            viper.GetString("REDIS_ADDR"),
		RedisPassword:        viper.GetString("REDIS_PASSWORD"),
		RedisDB:              viper.GetInt("REDIS_DB"),
		JWTSecret:            viper.GetString("JWT_SECRET"),
		JWTExpireHours:       viper.GetInt("JWT_EXPIRE_HOURS"),
		DASHSCOPE_API_KEY:    viper.GetString("DASHSCOPE_API_KEY"),
		DASHSCOPE_BASE_URL:   viper.GetString("DASHSCOPE_BASE_URL"),
		SMTPHost:             viper.GetString("SMTP_HOST"),
		SMTPPort:             viper.GetInt("SMTP_PORT"),
		SMTPUser:             viper.GetString("SMTP_USER"),
		SMTPPass:             viper.GetString("SMTP_PASS"),
		ResetTokenTTLMinutes: viper.GetInt("RESET_TOKEN_TTL_MINUTES"),
		BaseUrl:              viper.GetString("BASE_URL"),
	}
}
