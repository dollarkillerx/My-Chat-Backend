package main

import (
	"flag"
	"strings"

	"github.com/my-chat/common/pkg/client"
	"github.com/my-chat/common/pkg/config"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/seaking/internal/conf"
	"github.com/my-chat/seaking/internal/model"
	"github.com/my-chat/seaking/internal/server"
	"github.com/my-chat/seaking/internal/storage"
)

var (
	configName = flag.String("c", "config", "config file name")
	configPath = flag.String("cPath", "./,./configs/", "config file paths")
	migrate    = flag.Bool("migrate", false, "run database migration")
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

	log.Info().Str("service", cfg.Service.Name).Msg("starting seaking service")

	// 创建PostgreSQL客户端
	db, err := client.PostgresClient(cfg.Postgres, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	// 数据库迁移
	if *migrate {
		log.Info().Msg("running database migration")
		if err := db.AutoMigrate(
			&model.User{},
			&model.Friendship{},
			&model.FriendRequest{},
			&model.Group{},
			&model.GroupMember{},
			&model.Conversation{},
			&model.ConversationMember{},
		); err != nil {
			log.Fatal().Err(err).Msg("failed to migrate database")
		}
		log.Info().Msg("database migration completed")
	}

	// 创建Redis客户端
	redisClient, err := client.RedisClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	// 创建存储层
	st := storage.NewStorage(db, redisClient)

	// 创建并启动服务器
	srv := server.NewServer(cfg, st)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
