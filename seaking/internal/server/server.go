package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/common/pkg/middleware"
	"github.com/my-chat/seaking/internal/conf"
	"github.com/my-chat/seaking/internal/rpc"
	"github.com/my-chat/seaking/internal/service/conversation"
	"github.com/my-chat/seaking/internal/service/group"
	"github.com/my-chat/seaking/internal/service/key"
	"github.com/my-chat/seaking/internal/service/relation"
	"github.com/my-chat/seaking/internal/service/user"
	"github.com/my-chat/seaking/internal/storage"
)

// Server SeaKing服务器
type Server struct {
	config     conf.Config
	storage    *storage.Storage
	rpcHandler *rpc.Handler
	jwtManager *auth.JWTManager
	engine     *gin.Engine
}

// NewServer 创建服务器
func NewServer(config conf.Config, storage *storage.Storage) *Server {
	jwtManager := auth.NewJWTManager(config.JWT.Secret, config.JWT.ExpireHour)

	// 创建服务
	userService := user.NewService(storage)
	relationService := relation.NewService(storage)
	groupService := group.NewService(storage)
	convService := conversation.NewService(storage)
	keyService := key.NewService(storage)

	// 创建RPC处理器（内部服务通信）
	rpcHandler := rpc.NewHandler(userService, convService, relationService, groupService, keyService, jwtManager)

	return &Server{
		config:     config,
		storage:    storage,
		rpcHandler: rpcHandler,
		jwtManager: jwtManager,
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	if !s.config.Service.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	s.engine = gin.New()
	s.engine.Use(middleware.Recover())
	s.engine.Use(gin.Logger())
	s.engine.Use(middleware.Cors())

	s.registerRoutes()

	addr := fmt.Sprintf(":%s", s.config.Service.Port)
	log.Info().Str("addr", addr).Msg("seaking server starting")
	return s.engine.Run(addr)
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 健康检查
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "seaking",
		})
	})

	// JSON-RPC 接口（供内部服务调用）
	s.engine.POST("/api/rpc", s.rpcHandler.Handle)
}
