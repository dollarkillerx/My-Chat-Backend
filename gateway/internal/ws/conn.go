package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/my-chat/common/pkg/protocol"
)

// Conn WebSocket连接封装
type Conn struct {
	id        string
	uid       string
	deviceId  string
	platform  string
	conn      *websocket.Conn
	send      chan []byte
	hub       *Hub
	closeChan chan struct{}
	closeOnce sync.Once
	lastPing  time.Time
}

// NewConn 创建新连接
func NewConn(id, uid, deviceId, platform string, conn *websocket.Conn, hub *Hub) *Conn {
	return &Conn{
		id:        id,
		uid:       uid,
		deviceId:  deviceId,
		platform:  platform,
		conn:      conn,
		send:      make(chan []byte, 256),
		hub:       hub,
		closeChan: make(chan struct{}),
		lastPing:  time.Now(),
	}
}

// ID 获取连接ID
func (c *Conn) ID() string {
	return c.id
}

// UID 获取用户ID
func (c *Conn) UID() string {
	return c.uid
}

// DeviceId 获取设备ID
func (c *Conn) DeviceId() string {
	return c.deviceId
}

// Platform 获取平台
func (c *Conn) Platform() string {
	return c.platform
}

// Send 发送消息
func (c *Conn) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		// 发送队列满，关闭连接
		c.Close()
	}
}

// SendEnvelope 发送封包
func (c *Conn) SendEnvelope(env *protocol.Envelope) error {
	data, err := protocol.EncodeEnvelope(env)
	if err != nil {
		return err
	}
	c.Send(data)
	return nil
}

// Close 关闭连接
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
		c.conn.Close()
		c.hub.Unregister(c)
	})
}

// ReadPump 读取消息循环
func (c *Conn) ReadPump(handler func(*Conn, []byte)) {
	defer c.Close()

	c.conn.SetReadLimit(64 * 1024) // 64KB
	c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.hub.config.ReadTimeout) * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.hub.config.ReadTimeout) * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		handler(c, message)
	}
}

// WritePump 写入消息循环
func (c *Conn) WritePump() {
	ticker := time.NewTicker(time.Duration(c.hub.config.HeartbeatTimeout/2) * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.hub.config.WriteTimeout) * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.hub.config.WriteTimeout) * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.closeChan:
			return
		}
	}
}
