package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/my-chat/common/pkg/auth"
	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/common/pkg/middleware"
	"github.com/my-chat/seaking/internal/api"
	"github.com/my-chat/seaking/internal/conf"
	"github.com/my-chat/seaking/internal/rpc"
	"github.com/my-chat/seaking/internal/service/conversation"
	"github.com/my-chat/seaking/internal/service/group"
	"github.com/my-chat/seaking/internal/service/relation"
	"github.com/my-chat/seaking/internal/service/user"
	"github.com/my-chat/seaking/internal/storage"
)

// Server SeaKing服务器
type Server struct {
	config     conf.Config
	storage    *storage.Storage
	api        *api.API
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

	// 创建API
	apiHandler := api.NewAPI(userService, relationService, groupService, jwtManager)

	// 创建RPC处理器
	rpcHandler := rpc.NewHandler(userService, convService, jwtManager)

	return &Server{
		config:     config,
		storage:    storage,
		api:        apiHandler,
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

	// 公开接口
	public := s.engine.Group("/api/v1")
	{
		public.POST("/register", s.api.Register)
		public.POST("/login", s.api.Login)
	}

	// 需要认证的接口
	private := s.engine.Group("/api/v1")
	private.Use(s.authMiddleware())
	{
		// 用户
		private.GET("/profile", s.api.GetProfile)
		private.PUT("/profile", s.api.UpdateProfile)
		private.PUT("/password", s.api.ChangePassword)

		// 好友
		private.GET("/friends", s.api.GetFriends)
		private.POST("/friends/request", s.api.SendFriendRequest)
		private.POST("/friends/accept", s.api.AcceptFriendRequest)
		private.POST("/friends/reject", s.api.RejectFriendRequest)
		private.DELETE("/friends/:uid", s.api.DeleteFriend)
		private.POST("/friends/block", s.api.BlockFriend)
		private.POST("/friends/unblock", s.api.UnblockFriend)
		private.GET("/friends/requests", s.api.GetPendingRequests)

		// 群组
		private.GET("/groups", s.api.GetUserGroups)
		private.POST("/groups", s.api.CreateGroup)
		private.GET("/groups/:group_id", s.api.GetGroup)
		private.PUT("/groups/:group_id", s.api.UpdateGroup)
		private.DELETE("/groups/:group_id", s.api.DismissGroup)
		private.GET("/groups/:group_id/members", s.api.GetGroupMembers)
		private.POST("/groups/:group_id/members", s.api.AddGroupMember)
		private.DELETE("/groups/:group_id/members/:member_id", s.api.RemoveGroupMember)
		private.POST("/groups/:group_id/leave", s.api.LeaveGroup)
		private.POST("/groups/:group_id/transfer", s.api.TransferGroupOwner)
		private.POST("/groups/:group_id/admin", s.api.SetGroupAdmin)
	}
}

// authMiddleware 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			api.Error(c, errors.ErrLoginRequired)
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := s.jwtManager.ParseToken(token)
		if err != nil {
			api.Error(c, errors.ErrInvalidToken)
			c.Abort()
			return
		}

		c.Set("uid", claims.Uid)
		c.Set("device_id", claims.DeviceId)
		c.Set("platform", claims.Platform)
		c.Next()
	}
}
