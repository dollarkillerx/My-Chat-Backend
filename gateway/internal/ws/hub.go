package ws

import (
	"sync"

	"github.com/my-chat/common/pkg/log"
	"github.com/my-chat/gateway/internal/conf"
)

// Hub 连接管理中心
type Hub struct {
	config conf.GatewayConfiguration

	// 所有连接 connId -> *Conn
	conns sync.Map

	// 用户连接映射 uid -> map[connId]*Conn
	userConns sync.Map

	// 会话订阅 cid -> map[connId]*Conn
	subscriptions sync.Map

	register   chan *Conn
	unregister chan *Conn
	broadcast  chan *BroadcastMessage
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Cid  string
	Data []byte
}

// NewHub 创建Hub
func NewHub(config conf.GatewayConfiguration) *Hub {
	return &Hub{
		config:     config,
		register:   make(chan *Conn, 256),
		unregister: make(chan *Conn, 256),
		broadcast:  make(chan *BroadcastMessage, 1024),
	}
}

// Run 启动Hub
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.handleRegister(conn)

		case conn := <-h.unregister:
			h.handleUnregister(conn)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

// handleRegister 处理连接注册
func (h *Hub) handleRegister(conn *Conn) {
	// 存储连接
	h.conns.Store(conn.id, conn)

	// 用户连接映射
	userConnsI, _ := h.userConns.LoadOrStore(conn.uid, &sync.Map{})
	userConns := userConnsI.(*sync.Map)

	// 检查连接数限制
	count := 0
	userConns.Range(func(_, _ interface{}) bool {
		count++
		return true
	})

	if h.config.MaxConnPerUser > 0 && count >= h.config.MaxConnPerUser {
		// 踢掉最旧的连接
		var oldestConn *Conn
		userConns.Range(func(_, v interface{}) bool {
			if oldestConn == nil {
				oldestConn = v.(*Conn)
			}
			return false
		})
		if oldestConn != nil {
			oldestConn.Close()
		}
	}

	userConns.Store(conn.id, conn)

	log.Info().
		Str("conn_id", conn.id).
		Str("uid", conn.uid).
		Str("device_id", conn.deviceId).
		Msg("connection registered")
}

// handleUnregister 处理连接注销
func (h *Hub) handleUnregister(conn *Conn) {
	// 删除连接
	h.conns.Delete(conn.id)

	// 从用户连接映射删除
	if userConnsI, ok := h.userConns.Load(conn.uid); ok {
		userConns := userConnsI.(*sync.Map)
		userConns.Delete(conn.id)
	}

	// 从所有订阅中删除
	h.subscriptions.Range(func(_, v interface{}) bool {
		subs := v.(*sync.Map)
		subs.Delete(conn.id)
		return true
	})

	log.Info().
		Str("conn_id", conn.id).
		Str("uid", conn.uid).
		Msg("connection unregistered")
}

// handleBroadcast 处理广播消息
func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	if subsI, ok := h.subscriptions.Load(msg.Cid); ok {
		subs := subsI.(*sync.Map)
		subs.Range(func(_, v interface{}) bool {
			conn := v.(*Conn)
			conn.Send(msg.Data)
			return true
		})
	}
}

// Register 注册连接
func (h *Hub) Register(conn *Conn) {
	h.register <- conn
}

// Unregister 注销连接
func (h *Hub) Unregister(conn *Conn) {
	h.unregister <- conn
}

// Subscribe 订阅会话
func (h *Hub) Subscribe(conn *Conn, cid string) {
	subsI, _ := h.subscriptions.LoadOrStore(cid, &sync.Map{})
	subs := subsI.(*sync.Map)
	subs.Store(conn.id, conn)

	log.Debug().
		Str("conn_id", conn.id).
		Str("cid", cid).
		Msg("subscribed to conversation")
}

// Unsubscribe 取消订阅会话
func (h *Hub) Unsubscribe(conn *Conn, cid string) {
	if subsI, ok := h.subscriptions.Load(cid); ok {
		subs := subsI.(*sync.Map)
		subs.Delete(conn.id)
	}
}

// Broadcast 广播消息到会话
func (h *Hub) Broadcast(cid string, data []byte) {
	h.broadcast <- &BroadcastMessage{Cid: cid, Data: data}
}

// SendToUser 发送消息给用户的所有连接
func (h *Hub) SendToUser(uid string, data []byte) {
	if userConnsI, ok := h.userConns.Load(uid); ok {
		userConns := userConnsI.(*sync.Map)
		userConns.Range(func(_, v interface{}) bool {
			conn := v.(*Conn)
			conn.Send(data)
			return true
		})
	}
}

// GetConn 获取连接
func (h *Hub) GetConn(connId string) *Conn {
	if v, ok := h.conns.Load(connId); ok {
		return v.(*Conn)
	}
	return nil
}

// GetUserConns 获取用户的所有连接
func (h *Hub) GetUserConns(uid string) []*Conn {
	var conns []*Conn
	if userConnsI, ok := h.userConns.Load(uid); ok {
		userConns := userConnsI.(*sync.Map)
		userConns.Range(func(_, v interface{}) bool {
			conns = append(conns, v.(*Conn))
			return true
		})
	}
	return conns
}

// GetOnlineUsers 获取在线用户数
func (h *Hub) GetOnlineUsers() int {
	count := 0
	h.userConns.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// GetTotalConns 获取总连接数
func (h *Hub) GetTotalConns() int {
	count := 0
	h.conns.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
