package conf

import "github.com/my-chat/common/pkg/config"

// Config SeaKing配置
type Config struct {
	Service  config.ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	Postgres config.PostgresConfiguration `mapstructure:"PostgresConfiguration"`
	Redis    config.RedisConfiguration    `mapstructure:"RedisConfiguration"`
	Logger   config.LoggerConfiguration   `mapstructure:"LoggerConfiguration"`
	JWT      config.JWTConfiguration      `mapstructure:"JWTConfiguration"`
}
