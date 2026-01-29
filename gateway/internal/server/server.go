package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/common/pkg/middleware"
	"github.com/my-chat/common/pkg/storage"
	"github.com/my-chat/gateway/internal/conf"
	"github.com/my-chat/gateway/internal/handler"
	"github.com/my-chat/gateway/internal/rpc"
	"github.com/my-chat/gateway/internal/ws"
	"github.com/redis/go-redis/v9"
)

// Server Gateway服务器
type Server struct {
	config        conf.Config
	hub           *ws.Hub
	handler       *handler.Handler
	rpcHandler    *rpc.Handler
	uploadHandler *handler.UploadHandler
	jwtManager    *auth.JWTManager
	redis         *redis.Client
	r2            *storage.R2Storage
	engine        *gin.Engine
	upgrader      websocket.Upgrader
	connIdGen     int64
}

// NewServer 创建服务器
func NewServer(config conf.Config, redisClient *redis.Client, r2 *storage.R2Storage) *Server {
	jwtManager := auth.NewJWTManager(config.JWT.Secret, config.JWT.ExpireHour)
	hub := ws.NewHub(config.Gateway)
	h := handler.NewHandler(hub, jwtManager, config.Gateway.RelayAddr, config.Gateway.SeaKingAddr)
	rpcHandler := rpc.NewHandler(jwtManager, config.Gateway.SeaKingAddr, config.Gateway.RelayAddr)
	uploadHandler := handler.NewUploadHandler(r2, redisClient, config.Gateway.UploadRateLimit)

	return &Server{
		config:        config,
		hub:           hub,
		handler:       h,
		rpcHandler:    rpcHandler,
		uploadHandler: uploadHandler,
		jwtManager:    jwtManager,
		redis:         redisClient,
		r2:            r2,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境应该检查Origin
			},
		},
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	// 启动Hub
	go s.hub.Run()

	// 设置Gin模式
	if !s.config.Service.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	s.engine = gin.New()
	s.engine.Use(middleware.Recover())
	s.engine.Use(gin.Logger())
	s.engine.Use(middleware.Cors())

	// 注册路由
	s.registerRoutes()

	// 启动HTTP服务
	addr := fmt.Sprintf(":%s", s.config.Service.Port)
	log.Info().Str("addr", addr).Msg("gateway server starting")
	return s.engine.Run(addr)
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 健康检查
	s.engine.GET("/health", s.healthCheck)

	// JSON-RPC 接口（客户端调用）
	s.engine.POST("/api/rpc", s.rpcHandler.Handle)

	// WebSocket连接（消息推送）
	s.engine.GET(s.config.Gateway.WSPath, s.handleWebSocket)

	// 统计接口
	s.engine.GET("/api/stats", s.getStats)

	// 文件上传接口（需要认证）
	upload := s.engine.Group("/api/upload")
	upload.Use(s.authMiddleware())
	{
		upload.POST("/base64", s.uploadHandler.Upload)         // Base64 上传
		upload.POST("/multipart", s.uploadHandler.UploadMultipart) // Multipart 上传
	}
}

// authMiddleware JWT认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := s.jwtManager.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("uid", claims.Uid)
		c.Next()
	}
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "gateway",
	})
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(c *gin.Context) {
	// 获取token
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	// 验证token
	claims, err := s.jwtManager.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 升级为WebSocket
	wsConn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade connection")
		return
	}

	// 生成连接ID
	s.connIdGen++
	connId := fmt.Sprintf("%s-%d", claims.Uid, s.connIdGen)

	// 创建连接
	conn := ws.NewConn(connId, claims.Uid, claims.DeviceId, claims.Platform, wsConn, s.hub)

	// 注册连接
	s.hub.Register(conn)

	// 启动读写协程
	go conn.WritePump()
	conn.ReadPump(s.handler.HandleMessage)
}

// getStats 获取统计信息
func (s *Server) getStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"online_users": s.hub.GetOnlineUsers(),
		"total_conns":  s.hub.GetTotalConns(),
	})
}
