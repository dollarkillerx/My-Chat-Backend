package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/common/pkg/middleware"
	"github.com/my-chat/relay/internal/conf"
	"github.com/my-chat/relay/internal/rpc"
	"github.com/my-chat/relay/internal/service/event"
	"github.com/my-chat/relay/internal/storage"
)

// Server Relay服务器
type Server struct {
	config     conf.Config
	storage    *storage.Storage
	rpcHandler *rpc.Handler
	engine     *gin.Engine
}

// NewServer 创建服务器
func NewServer(config conf.Config, storage *storage.Storage) *Server {
	// 创建服务
	eventService := event.NewService(storage, config.Relay)

	// 创建RPC处理器（内部服务通信）
	rpcHandler := rpc.NewHandler(eventService, config.Relay)

	return &Server{
		config:     config,
		storage:    storage,
		rpcHandler: rpcHandler,
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
	log.Info().Str("addr", addr).Msg("relay server starting")
	return s.engine.Run(addr)
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 健康检查
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "relay",
		})
	})

	// JSON-RPC 接口（供内部服务调用）
	s.engine.POST("/api/rpc", s.rpcHandler.Handle)
}
