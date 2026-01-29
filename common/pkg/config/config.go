package config

import (
	"strings"

	"github.com/spf13/viper"
)

// ServiceConfiguration 服务配置
type ServiceConfiguration struct {
	Name  string `mapstructure:"Name"`
	Port  string `mapstructure:"Port"`
	Debug bool   `mapstructure:"Debug"`
}

// PostgresConfiguration PostgreSQL配置
type PostgresConfiguration struct {
	Host     string `mapstructure:"Host"`
	Port     int    `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	DBName   string `mapstructure:"DBName"`
	SSLMode  bool   `mapstructure:"SSLMode"`
	TimeZone string `mapstructure:"TimeZone"`
}

// RedisConfiguration Redis配置
type RedisConfiguration struct {
	Addr     string `mapstructure:"Addr"`
	Password string `mapstructure:"Password"`
	DB       int    `mapstructure:"Db"`
}

// LoggerConfiguration 日志配置
type LoggerConfiguration struct {
	Filename   string `mapstructure:"Filename"`
	MaxSize    int    `mapstructure:"MaxSize"`
	MaxBackups int    `mapstructure:"MaxBackups"`
	MaxAge     int    `mapstructure:"MaxAge"`
	Compress   bool   `mapstructure:"Compress"`
}

// JWTConfiguration JWT配置
type JWTConfiguration struct {
	Secret     string `mapstructure:"Secret"`
	ExpireHour int    `mapstructure:"ExpireHour"`
}

// R2Configuration Cloudflare R2 存储配置
type R2Configuration struct {
	Endpoint        string `mapstructure:"Endpoint"`        // R2 端点
	AccessKeyID     string `mapstructure:"AccessKeyID"`     // 访问密钥ID
	SecretAccessKey string `mapstructure:"SecretAccessKey"` // 访问密钥
	BucketName      string `mapstructure:"BucketName"`      // 存储桶名称
	ExportEndpoint  string `mapstructure:"ExportEndpoint"`  // 公开访问端点
	Region          string `mapstructure:"Region"`          // 区域，通常为 "auto"
}

// InitConfiguration 初始化配置
func InitConfiguration(configName string, configPaths []string, config interface{}) error {
	viper.SetConfigName(configName)
	viper.SetConfigType("toml")

	for _, path := range configPaths {
		viper.AddConfigPath(strings.TrimSpace(path))
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return viper.Unmarshal(config)
}
