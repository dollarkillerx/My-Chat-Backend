package conf

import "github.com/my-chat/common/pkg/config"

// Config Gateway配置
type Config struct {
	Service config.ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	Redis   config.RedisConfiguration    `mapstructure:"RedisConfiguration"`
	Logger  config.LoggerConfiguration   `mapstructure:"LoggerConfiguration"`
	JWT     config.JWTConfiguration      `mapstructure:"JWTConfiguration"`
	Gateway GatewayConfiguration         `mapstructure:"GatewayConfiguration"`
	R2      config.R2Configuration       `mapstructure:"R2Configuration"`
}

// GatewayConfiguration Gateway专属配置
type GatewayConfiguration struct {
	// WebSocket配置
	WSPath           string `mapstructure:"WSPath"`
	MaxConnPerUser   int    `mapstructure:"MaxConnPerUser"`
	HeartbeatTimeout int    `mapstructure:"HeartbeatTimeout"` // 秒
	WriteTimeout     int    `mapstructure:"WriteTimeout"`     // 秒
	ReadTimeout      int    `mapstructure:"ReadTimeout"`      // 秒

	// 服务发现
	SeaKingAddr string `mapstructure:"SeaKingAddr"`
	RelayAddr   string `mapstructure:"RelayAddr"`

	// 上传限制
	UploadRateLimit int `mapstructure:"UploadRateLimit"` // 每小时每用户最大上传次数，0 表示不限制
}
