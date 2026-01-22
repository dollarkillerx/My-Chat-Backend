package main

import (
	"flag"
	"strings"

	"github.com/my-chat/common/pkg/client"
	"github.com/my-chat/common/pkg/config"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/gateway/internal/conf"
	"github.com/my-chat/gateway/internal/server"
)

var (
	configName = flag.String("c", "config", "config file name")
	configPath = flag.String("cPath", "./,./configs/", "config file paths")
)

func main() {
	flag.Parse()

	// 加载配置
	var cfg conf.Config
	if err := config.InitConfiguration(*configName, strings.Split(*configPath, ","), &cfg); err != nil {
		panic(err)
	}

	// 初始化日志
	log.InitLog(cfg.Logger)

	log.Info().Str("service", cfg.Service.Name).Msg("starting gateway service")

	// 创建Redis客户端
	redisClient, err := client.RedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	// 创建并启动服务器
	srv := server.NewServer(cfg, redisClient)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
