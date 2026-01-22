package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/common/pkg/client"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/common/pkg/middleware"
	"github.com/my-chat/gateway/internal/conf"
	"github.com/my-chat/gateway/internal/handler"
	"github.com/my-chat/gateway/internal/ws"
	"github.com/redis/go-redis/v9"
)

// Server Gateway服务器
type Server struct {
	config        conf.Config
	hub           *ws.Hub
	handler       *handler.Handler
	jwtManager    *auth.JWTManager
	seakingClient *client.SeaKingClient
	redis         *redis.Client
	engine        *gin.Engine
	upgrader      websocket.Upgrader
	connIdGen     int64
}

// NewServer 创建服务器
func NewServer(config conf.Config, redisClient *redis.Client) *Server {
	jwtManager := auth.NewJWTManager(config.JWT.Secret, config.JWT.ExpireHour)
	hub := ws.NewHub(config.Gateway)
	h := handler.NewHandler(hub, jwtManager, config.Gateway.RelayAddr, config.Gateway.SeaKingAddr)
	seakingClient := client.NewSeaKingClient(config.Gateway.SeaKingAddr)

	return &Server{
		config:        config,
		hub:           hub,
		handler:       h,
		jwtManager:    jwtManager,
		seakingClient: seakingClient,
		redis:         redisClient,
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

	// WebSocket连接
	s.engine.GET(s.config.Gateway.WSPath, s.handleWebSocket)

	// API接口
	api := s.engine.Group("/api")
	{
		api.GET("/stats", s.getStats)
		// 认证接口（无需token）
		api.POST("/register", s.handleRegister)
		api.POST("/login", s.handleLogin)
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

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
}

// handleRegister 处理用户注册
func (s *Server) handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := s.seakingClient.Register(c.Request.Context(), &client.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Email:    req.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":      resp.Uid,
		"username": resp.Username,
		"nickname": resp.Nickname,
	})
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceId string `json:"device_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
}

// handleLogin 处理用户登录
func (s *Server) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := s.seakingClient.Login(c.Request.Context(), &client.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}, req.DeviceId, req.Platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": resp.Token,
		"user":  resp.User,
	})
}
