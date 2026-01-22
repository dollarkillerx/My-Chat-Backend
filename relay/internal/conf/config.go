package conf

import "github.com/my-chat/common/pkg/config"

// Config Relay配置
type Config struct {
	Service  config.ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	Postgres config.PostgresConfiguration `mapstructure:"PostgresConfiguration"`
	Redis    config.RedisConfiguration    `mapstructure:"RedisConfiguration"`
	Logger   config.LoggerConfiguration   `mapstructure:"LoggerConfiguration"`
	Relay    RelayConfiguration           `mapstructure:"RelayConfiguration"`
}

// RelayConfiguration Relay专属配置
type RelayConfiguration struct {
	// 消息保留天数（0表示永久保留）
	RetentionDays int `mapstructure:"RetentionDays"`
	// 单次查询最大数量
	MaxQueryLimit int `mapstructure:"MaxQueryLimit"`
}
