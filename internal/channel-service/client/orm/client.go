package orm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/orm/sqlopt"
)

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) (*Client, error) {
	if err := db.AutoMigrate(
		&model.Channel{},
		&model.QRSession{},
		&model.ChannelConversation{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate channel tables: %w", err)
	}
	return &Client{db: db}, nil
}

// --- Channel CRUD ---

func (c *Client) CreateChannel(ctx context.Context, channel *model.Channel) (*model.Channel, error) {
	if err := c.db.WithContext(ctx).Create(channel).Error; err != nil {
		return nil, fmt.Errorf("channel_create: %w", err)
	}
	return channel, nil
}

func (c *Client) GetChannel(ctx context.Context, channelID string) (*model.Channel, error) {
	var channel model.Channel
	if err := sqlopt.SQLOptions(
		sqlopt.WithChannelID(channelID),
	).Apply(c.db.WithContext(ctx)).First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("channel_not_found: %s", channelID)
		}
		return nil, fmt.Errorf("channel_get: %w", err)
	}
	return &channel, nil
}

func (c *Client) ListChannels(ctx context.Context, userID, orgID, name string, pageNo, pageSize int32) ([]*model.Channel, int64, error) {
	opts := []sqlopt.SQLOption{
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithChannelName(name),
	}

	// 查询总数
	var total int64
	if err := sqlopt.SQLOptions(opts...).Apply(c.db.WithContext(ctx).Model(&model.Channel{})).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("channel_count: %w", err)
	}

	// 分页查询
	var channels []*model.Channel
	offset := int((pageNo - 1) * pageSize)
	if offset < 0 {
		offset = 0
	}
	if err := sqlopt.SQLOptions(opts...).Apply(c.db.WithContext(ctx)).
		Order("created_at DESC").
		Offset(offset).Limit(int(pageSize)).
		Find(&channels).Error; err != nil {
		return nil, 0, fmt.Errorf("channel_list: %w", err)
	}
	return channels, total, nil
}

func (c *Client) UpdateChannel(ctx context.Context, channelID string, updates map[string]interface{}) (*model.Channel, error) {
	result := sqlopt.SQLOptions(
		sqlopt.WithChannelID(channelID),
	).Apply(c.db.WithContext(ctx)).Model(&model.Channel{}).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("channel_update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("channel_not_found: %s", channelID)
	}
	return c.GetChannel(ctx, channelID)
}

func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	result := sqlopt.SQLOptions(
		sqlopt.WithChannelID(channelID),
	).Apply(c.db.WithContext(ctx)).Delete(&model.Channel{})
	if result.Error != nil {
		return fmt.Errorf("channel_delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("channel_not_found: %s", channelID)
	}
	return nil
}

func (c *Client) ListEnabledChannels(ctx context.Context) ([]*model.Channel, error) {
	var channels []*model.Channel
	if err := c.db.WithContext(ctx).Where("enabled = ? AND status = ?", true, "loggedIn").
		Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("channel_list_enabled: %w", err)
	}
	return channels, nil
}

// --- QR Session CRUD ---

func (c *Client) CreateQRSession(ctx context.Context, session *model.QRSession) (*model.QRSession, error) {
	if err := c.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("qr_session_create: %w", err)
	}
	return session, nil
}

func (c *Client) GetQRSession(ctx context.Context, sessionID string) (*model.QRSession, error) {
	var session model.QRSession
	if err := c.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("qr_session_not_found: %s", sessionID)
		}
		return nil, fmt.Errorf("qr_session_get: %w", err)
	}
	return &session, nil
}

func (c *Client) UpdateQRSession(ctx context.Context, sessionID string, updates map[string]interface{}) error {
	result := c.db.WithContext(ctx).Model(&model.QRSession{}).Where("session_id = ?", sessionID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("qr_session_update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("qr_session_not_found: %s", sessionID)
	}
	return nil
}

func (c *Client) DeleteQRSession(ctx context.Context, sessionID string) error {
	result := c.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&model.QRSession{})
	if result.Error != nil {
		return fmt.Errorf("qr_session_delete: %w", result.Error)
	}
	return nil
}

// DeleteExpiredQRSessions 删除所有已过期的扫码会话（expire_at < now），返回删除条数。
// 用于定时清理，避免 qr_sessions 表无限堆积。
func (c *Client) DeleteExpiredQRSessions(ctx context.Context, now int64) (int64, error) {
	result := c.db.WithContext(ctx).Where("expire_at < ?", now).Delete(&model.QRSession{})
	if result.Error != nil {
		return 0, fmt.Errorf("qr_session_delete_expired: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// --- ChannelConversation 会话映射 ---

// GetConversation 按 (channelID, userID, appType) 查询会话映射，未找到返回 gorm.ErrRecordNotFound。
func (c *Client) GetConversation(ctx context.Context, channelID, userID, appType string) (*model.ChannelConversation, error) {
	var conv model.ChannelConversation
	if err := c.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ? AND app_type = ?", channelID, userID, appType).
		First(&conv).Error; err != nil {
		return nil, fmt.Errorf("conversation_get: %w", err)
	}
	return &conv, nil
}

// UpsertConversation 按 (channelID, userID, appType) 插入或更新会话映射。
// 同一 channel+user+appType 复用同一行，仅更新 conversation_id。
func (c *Client) UpsertConversation(ctx context.Context, conv *model.ChannelConversation) error {
	// 按 (channelID, userID, appType) 唯一索引 upsert：
	// 不存在 → 用完整字段插入；存在 → 仅更新 conversation_id。
	// 注意：Where 只用于查询条件，FirstOrCreate 插入时取的是传入的空结构体，
	// 必须用 Attrs 显式带上唯一索引键 + conversation_id，否则会插入全空行。
	result := c.db.WithContext(ctx).Where(
		"channel_id = ? AND user_id = ? AND app_type = ?",
		conv.ChannelID, conv.UserID, conv.AppType,
	).Attrs(model.ChannelConversation{
		ChannelID:      conv.ChannelID,
		UserID:         conv.UserID,
		AppType:        conv.AppType,
		ConversationID: conv.ConversationID,
	}).Assign(model.ChannelConversation{
		ConversationID: conv.ConversationID,
	}).FirstOrCreate(&model.ChannelConversation{})
	if result.Error != nil {
		return fmt.Errorf("conversation_upsert: %w", result.Error)
	}
	return nil
}

// DeleteConversationsByChannel 删除指定通道下的所有会话映射，用于删通道时级联清理。
func (c *Client) DeleteConversationsByChannel(ctx context.Context, channelID string) error {
	result := c.db.WithContext(ctx).Where("channel_id = ?", channelID).Delete(&model.ChannelConversation{})
	if result.Error != nil {
		return fmt.Errorf("conversation_delete_by_channel: %w", result.Error)
	}
	return nil
}
