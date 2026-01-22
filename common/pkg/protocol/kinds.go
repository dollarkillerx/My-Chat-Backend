package protocol

// Kind 消息类型常量
const (
	// KindText 文本消息
	KindText = 1
	// KindFile 文件消息（图片/语音/文件）
	KindFile = 3
	// KindRevoke 撤销消息
	KindRevoke = 5
	// KindEdit 编辑消息
	KindEdit = 7
	// KindReadReceipt 已读回执
	KindReadReceipt = 10
	// KindTyping 正在输入
	KindTyping = 11
	// KindReaction 消息反应
	KindReaction = 12
	// KindForward 转发消息
	KindForward = 13
)

// KindName 获取Kind名称
func KindName(kind int) string {
	switch kind {
	case KindText:
		return "text"
	case KindFile:
		return "file"
	case KindRevoke:
		return "revoke"
	case KindEdit:
		return "edit"
	case KindReadReceipt:
		return "read_receipt"
	case KindTyping:
		return "typing"
	case KindReaction:
		return "reaction"
	case KindForward:
		return "forward"
	default:
		return "unknown"
	}
}

// IsPersistent 判断消息类型是否需要持久化
func IsPersistent(kind int) bool {
	switch kind {
	case KindTyping:
		return false
	default:
		return true
	}
}
