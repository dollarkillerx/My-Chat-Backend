package protocol

// Event 消息事件
type Event struct {
	Version   int                    `msgpack:"0" json:"ev_v"`   // 事件版本
	Cid       string                 `msgpack:"1" json:"cid"`    // 会话ID
	Kind      int                    `msgpack:"2" json:"k"`      // 消息类型
	Mid       int64                  `msgpack:"3" json:"mid"`    // 消息ID（服务端生成）
	Timestamp int64                  `msgpack:"4" json:"t"`      // 时间戳（秒）
	Flags     int                    `msgpack:"5" json:"flg"`    // 标志位
	Tags      []Tag                  `msgpack:"6" json:"tags"`   // 关联标签
	Data      map[int]interface{}    `msgpack:"7" json:"data"`   // 结构化消息体
	Sig       string                 `msgpack:"8" json:"sig"`    // 可选签名
	Sender    string                 `msgpack:"9" json:"sender"` // 发送者UID
	Ext       map[string]interface{} `msgpack:"15" json:"ext"`   // 扩展字段
}

// NewEvent 创建新事件
func NewEvent(kind int, cid string, sender string) *Event {
	return &Event{
		Version: 1,
		Kind:    kind,
		Cid:     cid,
		Sender:  sender,
		Tags:    make([]Tag, 0),
		Data:    make(map[int]interface{}),
	}
}

// SetText 设置文本内容（Kind=1）
func (e *Event) SetText(content string) *Event {
	e.Data[0] = content
	return e
}

// GetText 获取文本内容
func (e *Event) GetText() string {
	if v, ok := e.Data[0].(string); ok {
		return v
	}
	return ""
}

// SetFileData 设置文件数据（Kind=3）
func (e *Event) SetFileData(fid, name string, size int64, mime, sha256, url string) *Event {
	e.Data[0] = fid
	e.Data[1] = name
	e.Data[2] = size
	e.Data[3] = mime
	e.Data[4] = sha256
	e.Data[5] = url
	return e
}

// FileData 文件数据结构
type FileData struct {
	Fid    string `json:"fid"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Mime   string `json:"mime"`
	SHA256 string `json:"sha256"`
	URL    string `json:"url"`
}

// GetFileData 获取文件数据
func (e *Event) GetFileData() *FileData {
	fd := &FileData{}
	if v, ok := e.Data[0].(string); ok {
		fd.Fid = v
	}
	if v, ok := e.Data[1].(string); ok {
		fd.Name = v
	}
	if v, ok := e.Data[2].(int64); ok {
		fd.Size = v
	}
	if v, ok := e.Data[3].(string); ok {
		fd.Mime = v
	}
	if v, ok := e.Data[4].(string); ok {
		fd.SHA256 = v
	}
	if v, ok := e.Data[5].(string); ok {
		fd.URL = v
	}
	return fd
}

// SetRevokeData 设置撤销数据（Kind=5）
func (e *Event) SetRevokeData(targetMid int64, scope int, reason string) *Event {
	e.Tags = append(e.Tags, NewTargetTag(targetMid))
	e.Data[0] = scope
	e.Data[1] = reason
	return e
}

// SetEditData 设置编辑数据（Kind=7）
func (e *Event) SetEditData(targetMid int64, newContent string, version int) *Event {
	e.Tags = append(e.Tags, NewTargetTag(targetMid))
	e.Data[0] = newContent
	e.Data[1] = version
	return e
}

// SetReadReceipt 设置已读回执数据（Kind=10）
func (e *Event) SetReadReceipt(lastReadMid int64) *Event {
	e.Data[0] = lastReadMid
	return e
}

// SetTyping 设置正在输入状态（Kind=11）
func (e *Event) SetTyping(state int) *Event {
	e.Data[0] = state
	return e
}

// SetReaction 设置消息反应（Kind=12）
func (e *Event) SetReaction(targetMid int64, emoji string, action int) *Event {
	e.Tags = append(e.Tags, NewTargetTag(targetMid))
	e.Data[0] = emoji
	e.Data[1] = action
	return e
}

// ForwardType 转发类型
const (
	ForwardTypeSingle = 1 // 单条转发
	ForwardTypeMerge  = 2 // 合并转发
)

// SetForward 设置转发数据（Kind=13）
func (e *Event) SetForward(sourceCid string, sourceMid int64, forwardType int, snapshot interface{}) *Event {
	e.Tags = append(e.Tags, NewForwardCidTag(sourceCid))
	if forwardType == ForwardTypeSingle {
		e.Tags = append(e.Tags, NewForwardMidTag(sourceMid))
	}
	e.Data[0] = forwardType
	e.Data[1] = snapshot
	return e
}

// AddReplyTag 添加回复标签
func (e *Event) AddReplyTag(mid int64) *Event {
	e.Tags = append(e.Tags, NewReplyTag(mid))
	return e
}

// AddMentionTag 添加@提及标签
func (e *Event) AddMentionTag(uid string) *Event {
	e.Tags = append(e.Tags, NewMentionTag(uid))
	return e
}

// AddMentionAllTag 添加@全体成员标签
func (e *Event) AddMentionAllTag() *Event {
	e.Tags = append(e.Tags, NewMentionAllTag())
	return e
}
