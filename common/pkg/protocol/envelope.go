package protocol

// Command 命令类型常量
const (
	// CmdEvent 事件消息
	CmdEvent = "event"
	// CmdAck 确认消息
	CmdAck = "ack"
	// CmdError 错误消息
	CmdError = "error"
	// CmdPing 心跳请求
	CmdPing = "ping"
	// CmdPong 心跳响应
	CmdPong = "pong"
	// CmdAuth 认证请求
	CmdAuth = "auth"
	// CmdAuthResult 认证结果
	CmdAuthResult = "auth_result"
	// CmdSubscribe 订阅会话
	CmdSubscribe = "subscribe"
	// CmdUnsubscribe 取消订阅
	CmdUnsubscribe = "unsubscribe"
	// CmdSync 同步消息
	CmdSync = "sync"
	// CmdSearch 搜索请求
	CmdSearch = "search"
	// CmdSearchResult 搜索结果
	CmdSearchResult = "search_result"

	// 好友相关命令
	// CmdGetFriends 获取好友列表
	CmdGetFriends = "get_friends"
	// CmdSendFriendRequest 发送好友请求
	CmdSendFriendRequest = "send_friend_request"
	// CmdGetFriendRequests 获取待处理好友请求
	CmdGetFriendRequests = "get_friend_requests"
	// CmdAcceptFriendRequest 接受好友请求
	CmdAcceptFriendRequest = "accept_friend_request"
	// CmdRejectFriendRequest 拒绝好友请求
	CmdRejectFriendRequest = "reject_friend_request"
	// CmdDeleteFriend 删除好友
	CmdDeleteFriend = "delete_friend"
	// CmdResult 通用结果响应
	CmdResult = "result"

	// 会话相关命令
	// CmdGetConversations 获取会话列表
	CmdGetConversations = "get_conversations"
	// CmdCreateConversation 创建会话
	CmdCreateConversation = "create_conversation"
	// CmdGetConversationMembers 获取会话成员
	CmdGetConversationMembers = "get_conversation_members"

	// 群组相关命令
	// CmdGetGroups 获取群组列表
	CmdGetGroups = "get_groups"
	// CmdCreateGroup 创建群组
	CmdCreateGroup = "create_group"
	// CmdGetGroupInfo 获取群组信息
	CmdGetGroupInfo = "get_group_info"
	// CmdGetGroupMembers 获取群组成员
	CmdGetGroupMembers = "get_group_members"

	// 用户相关命令
	// CmdGetUserInfo 获取用户信息
	CmdGetUserInfo = "get_user_info"
)

// Envelope 网络封包
type Envelope struct {
	Version int         `msgpack:"0" json:"v"`   // 协议版本
	Cmd     string      `msgpack:"1" json:"cmd"` // 命令类型
	Seq     int64       `msgpack:"2" json:"seq"` // 客户端序列号
	Sid     string      `msgpack:"3" json:"sid"` // 会话ID（可选）
	Body    interface{} `msgpack:"4" json:"body"` // 负载
	Ext     interface{} `msgpack:"15" json:"ext"` // 扩展字段
}

// NewEnvelope 创建新封包
func NewEnvelope(cmd string, seq int64, body interface{}) *Envelope {
	return &Envelope{
		Version: 1,
		Cmd:     cmd,
		Seq:     seq,
		Body:    body,
	}
}

// AckBody 确认消息体
type AckBody struct {
	Seq int64 `msgpack:"0" json:"seq"` // 确认的序列号
	Mid int64 `msgpack:"1" json:"mid"` // 消息ID
}

// ErrorBody 错误消息体
type ErrorBody struct {
	Code    int    `msgpack:"0" json:"code"`    // 错误码
	Message string `msgpack:"1" json:"message"` // 错误信息
	Seq     int64  `msgpack:"2" json:"seq"`     // 关联的请求序列号
}

// AuthBody 认证请求体
type AuthBody struct {
	Token    string `msgpack:"0" json:"token"`     // 认证令牌
	DeviceId string `msgpack:"1" json:"device_id"` // 设备ID
	Platform string `msgpack:"2" json:"platform"`  // 平台（ios/android/web）
}

// AuthResultBody 认证结果体
type AuthResultBody struct {
	Success bool   `msgpack:"0" json:"success"` // 是否成功
	Uid     string `msgpack:"1" json:"uid"`     // 用户ID
	Message string `msgpack:"2" json:"message"` // 消息
}

// SyncBody 同步请求体
type SyncBody struct {
	Cid      string `msgpack:"0" json:"cid"`       // 会话ID
	LastMid  int64  `msgpack:"1" json:"last_mid"`  // 最后一条消息ID
	Limit    int    `msgpack:"2" json:"limit"`     // 数量限制
	Before   int64  `msgpack:"3" json:"before"`    // 时间戳上限
	After    int64  `msgpack:"4" json:"after"`     // 时间戳下限
}

// SearchBody 搜索请求体
type SearchBody struct {
	Cid    string `msgpack:"0" json:"cid"`    // 会话ID（可选）
	Query  string `msgpack:"1" json:"q"`      // 搜索词
	Kinds  []int  `msgpack:"2" json:"kinds"`  // 消息类型（可选）
	Before int64  `msgpack:"3" json:"before"` // 时间戳上限
	After  int64  `msgpack:"4" json:"after"`  // 时间戳下限
	Limit  int    `msgpack:"5" json:"limit"`  // 数量限制
	Offset int    `msgpack:"6" json:"offset"` // 偏移量
}

// SearchResultBody 搜索结果体
type SearchResultBody struct {
	Total int           `msgpack:"0" json:"total"` // 总数
	Items []SearchItem  `msgpack:"1" json:"items"` // 结果列表
}

// SearchItem 搜索结果项
type SearchItem struct {
	Mid       int64  `msgpack:"0" json:"mid"`
	Cid       string `msgpack:"1" json:"cid"`
	Kind      int    `msgpack:"2" json:"k"`
	Data      interface{} `msgpack:"3" json:"data"`
	Timestamp int64  `msgpack:"4" json:"t"`
	Highlight string `msgpack:"5" json:"highlight"`
}
