package client

import (
	"fmt"
	"time"

	"github.com/my-chat/common/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresClient 创建PostgreSQL客户端
func PostgresClient(cfg config.PostgresConfiguration, gormConfig *gorm.Config) (*gorm.DB, error) {
	sslMode := "disable"
	if cfg.SSLMode {
		sslMode = "require"
	}

	timeZone := cfg.TimeZone
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, sslMode, timeZone,
	)

	if gormConfig == nil {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
