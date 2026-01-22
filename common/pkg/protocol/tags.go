package protocol

// Tag类型常量
const (
	// TagReply 回复引用 [1, mid]
	TagReply = 1
	// TagMention @提及 [2, uid] or [2, "all"]
	TagMention = 2
	// TagTarget 目标消息引用（撤销/编辑/Reaction）[6, mid]
	TagTarget = 6
	// TagForwardCid 转发来源会话 [8, cid]
	TagForwardCid = 8
	// TagForwardMid 转发来源消息 [9, mid]
	TagForwardMid = 9
)

// Tag 标签结构
type Tag struct {
	Type  int         `msgpack:"0" json:"type"`
	Value interface{} `msgpack:"1" json:"value"`
}

// NewReplyTag 创建回复标签
func NewReplyTag(mid int64) Tag {
	return Tag{Type: TagReply, Value: mid}
}

// NewMentionTag 创建@提及标签
func NewMentionTag(uid string) Tag {
	return Tag{Type: TagMention, Value: uid}
}

// NewMentionAllTag 创建@全体成员标签
func NewMentionAllTag() Tag {
	return Tag{Type: TagMention, Value: "all"}
}

// NewTargetTag 创建目标消息标签（撤销/编辑/Reaction）
func NewTargetTag(mid int64) Tag {
	return Tag{Type: TagTarget, Value: mid}
}

// NewForwardCidTag 创建转发来源会话标签
func NewForwardCidTag(cid string) Tag {
	return Tag{Type: TagForwardCid, Value: cid}
}

// NewForwardMidTag 创建转发来源消息标签
func NewForwardMidTag(mid int64) Tag {
	return Tag{Type: TagForwardMid, Value: mid}
}

// ParseTags 解析标签数组
func ParseTags(tags []Tag) map[int][]interface{} {
	result := make(map[int][]interface{})
	for _, tag := range tags {
		result[tag.Type] = append(result[tag.Type], tag.Value)
	}
	return result
}

// GetReplyMid 从标签中获取回复的消息ID
func GetReplyMid(tags []Tag) (int64, bool) {
	for _, tag := range tags {
		if tag.Type == TagReply {
			if mid, ok := tag.Value.(int64); ok {
				return mid, true
			}
		}
	}
	return 0, false
}

// GetTargetMid 从标签中获取目标消息ID
func GetTargetMid(tags []Tag) (int64, bool) {
	for _, tag := range tags {
		if tag.Type == TagTarget {
			if mid, ok := tag.Value.(int64); ok {
				return mid, true
			}
		}
	}
	return 0, false
}

// GetMentions 从标签中获取所有@提及
func GetMentions(tags []Tag) []string {
	var mentions []string
	for _, tag := range tags {
		if tag.Type == TagMention {
			if uid, ok := tag.Value.(string); ok {
				mentions = append(mentions, uid)
			}
		}
	}
	return mentions
}
