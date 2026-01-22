package storage

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Storage 存储层
type Storage struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewStorage 创建存储层
func NewStorage(db *gorm.DB, redis *redis.Client) *Storage {
	return &Storage{
		db:    db,
		redis: redis,
	}
}

// DB 获取数据库连接
func (s *Storage) DB() *gorm.DB {
	return s.db
}

// Redis 获取Redis连接
func (s *Storage) Redis() *redis.Client {
	return s.redis
}
