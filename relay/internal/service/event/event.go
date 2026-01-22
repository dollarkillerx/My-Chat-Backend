package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/my-chat/common/pkg/errors"
	"github.com/my-chat/common/pkg/protocol"
	"github.com/my-chat/relay/internal/conf"
	"github.com/my-chat/relay/internal/model"
	"github.com/my-chat/relay/internal/storage"
	"github.com/redis/go-redis/v9"
)

// Service 事件服务
type Service struct {
	storage *storage.Storage
	config  conf.RelayConfiguration
}

// NewService 创建事件服务
func NewService(storage *storage.Storage, config conf.RelayConfiguration) *Service {
	return &Service{
		storage: storage,
		config:  config,
	}
}

// StoreEvent 存储事件
func (s *Service) StoreEvent(ctx context.Context, event *protocol.Event) (*model.Event, error) {
	// 生成消息ID
	mid, err := s.generateMid(ctx, event.Cid)
	if err != nil {
		return nil, err
	}

	// 序列化tags和data
	tagsJSON, _ := json.Marshal(event.Tags)
	dataJSON, _ := json.Marshal(event.Data)

	// 创建存储模型
	e := &model.Event{
		Mid:       mid,
		Cid:       event.Cid,
		Kind:      event.Kind,
		Sender:    event.Sender,
		Tags:      string(tagsJSON),
		Data:      string(dataJSON),
		Flags:     event.Flags,
		Sig:       event.Sig,
		Timestamp: time.Now().Unix(),
	}

	if err := s.storage.DB().Create(e).Error; err != nil {
		return nil, err
	}

	return e, nil
}

// generateMid 生成消息ID（使用Redis自增）
func (s *Service) generateMid(ctx context.Context, cid string) (int64, error) {
	key := fmt.Sprintf("mid:%s", cid)
	mid, err := s.storage.Redis().Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return mid, nil
}

// GetEvent 获取单条事件
func (s *Service) GetEvent(ctx context.Context, mid int64) (*model.Event, error) {
	var event model.Event
	if err := s.storage.DB().Where("mid = ?", mid).First(&event).Error; err != nil {
		return nil, errors.ErrMessageNotFound
	}
	return &event, nil
}

// QueryRequest 查询请求
type QueryRequest struct {
	Cid     string `json:"cid"`
	LastMid int64  `json:"last_mid"` // 从这条消息之后查询
	Before  int64  `json:"before"`   // 时间戳上限
	After   int64  `json:"after"`    // 时间戳下限
	Kinds   []int  `json:"kinds"`    // 消息类型筛选
	Limit   int    `json:"limit"`
}

// QueryEvents 查询事件列表
func (s *Service) QueryEvents(ctx context.Context, req *QueryRequest) ([]model.Event, error) {
	if req.Limit <= 0 || req.Limit > s.config.MaxQueryLimit {
		req.Limit = s.config.MaxQueryLimit
	}

	query := s.storage.DB().Where("cid = ?", req.Cid)

	if req.LastMid > 0 {
		query = query.Where("mid > ?", req.LastMid)
	}

	if req.Before > 0 {
		query = query.Where("timestamp < ?", req.Before)
	}

	if req.After > 0 {
		query = query.Where("timestamp > ?", req.After)
	}

	if len(req.Kinds) > 0 {
		query = query.Where("kind IN ?", req.Kinds)
	}

	var events []model.Event
	err := query.Order("mid ASC").Limit(req.Limit).Find(&events).Error
	return events, err
}

// QueryEventsDesc 逆序查询事件（获取最新消息）
func (s *Service) QueryEventsDesc(ctx context.Context, cid string, limit int) ([]model.Event, error) {
	if limit <= 0 || limit > s.config.MaxQueryLimit {
		limit = s.config.MaxQueryLimit
	}

	var events []model.Event
	err := s.storage.DB().
		Where("cid = ?", cid).
		Order("mid DESC").
		Limit(limit).
		Find(&events).Error

	// 反转顺序
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events, err
}

// UpdateReadReceipt 更新已读回执
func (s *Service) UpdateReadReceipt(ctx context.Context, cid, uid string, lastReadMid int64) error {
	receipt := &model.ReadReceipt{
		Cid:         cid,
		Uid:         uid,
		LastReadMid: lastReadMid,
		UpdatedAt:   time.Now(),
	}

	return s.storage.DB().
		Where("cid = ? AND uid = ?", cid, uid).
		Assign(receipt).
		FirstOrCreate(receipt).Error
}

// GetReadReceipt 获取已读回执
func (s *Service) GetReadReceipt(ctx context.Context, cid, uid string) (*model.ReadReceipt, error) {
	var receipt model.ReadReceipt
	err := s.storage.DB().Where("cid = ? AND uid = ?", cid, uid).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetConversationReadReceipts 获取会话的所有已读回执
func (s *Service) GetConversationReadReceipts(ctx context.Context, cid string) ([]model.ReadReceipt, error) {
	var receipts []model.ReadReceipt
	err := s.storage.DB().Where("cid = ?", cid).Find(&receipts).Error
	return receipts, err
}

// AddReaction 添加反应
func (s *Service) AddReaction(ctx context.Context, mid int64, cid, uid, emoji string) error {
	reaction := &model.Reaction{
		Mid:   mid,
		Cid:   cid,
		Uid:   uid,
		Emoji: emoji,
	}

	// 检查是否已存在
	var existing model.Reaction
	err := s.storage.DB().Where("mid = ? AND uid = ? AND emoji = ?", mid, uid, emoji).First(&existing).Error
	if err == nil {
		// 已存在，不需要重复添加
		return nil
	}

	return s.storage.DB().Create(reaction).Error
}

// RemoveReaction 移除反应
func (s *Service) RemoveReaction(ctx context.Context, mid int64, uid, emoji string) error {
	return s.storage.DB().
		Where("mid = ? AND uid = ? AND emoji = ?", mid, uid, emoji).
		Delete(&model.Reaction{}).Error
}

// GetReactions 获取消息的所有反应
func (s *Service) GetReactions(ctx context.Context, mid int64) ([]model.Reaction, error) {
	var reactions []model.Reaction
	err := s.storage.DB().Where("mid = ?", mid).Find(&reactions).Error
	return reactions, err
}

// GetReactionSummary 获取消息的反应汇总
func (s *Service) GetReactionSummary(ctx context.Context, mid int64) (map[string]int, error) {
	var results []struct {
		Emoji string
		Count int
	}

	err := s.storage.DB().Model(&model.Reaction{}).
		Select("emoji, count(*) as count").
		Where("mid = ?", mid).
		Group("emoji").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, r := range results {
		summary[r.Emoji] = r.Count
	}

	return summary, nil
}

// CacheLastMid 缓存会话最后消息ID
func (s *Service) CacheLastMid(ctx context.Context, cid string, mid int64) error {
	key := fmt.Sprintf("last_mid:%s", cid)
	return s.storage.Redis().Set(ctx, key, mid, 24*time.Hour).Err()
}

// GetCachedLastMid 获取缓存的最后消息ID
func (s *Service) GetCachedLastMid(ctx context.Context, cid string) (int64, error) {
	key := fmt.Sprintf("last_mid:%s", cid)
	mid, err := s.storage.Redis().Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return mid, err
}
