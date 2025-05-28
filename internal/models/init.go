package models

import (
	"context"
	"fmt"

	"awesomeProject/config"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var RDB *redis.Client

func InitMySQL(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQLUser, cfg.MySQLPassword, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("MySQL connect failed: " + err.Error())
	}
	DB = db
	DB.AutoMigrate(&User{}, &Image{}, &Task{}, &Favorite{})
}

func InitRedis(cfg *config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		panic("Redis connect failed: " + err.Error())
	}
}
