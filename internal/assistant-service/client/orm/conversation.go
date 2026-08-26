package orm

import (
	"context"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateConversation(ctx context.Context, conversation *model.Conversation) *err_code.Status {
	if conversation.ID != 0 {
		return toErrStatus("assistant_conversation_create", "create conversation but id not 0")
	}
	if err := c.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return toErrStatus("assistant_conversation_create", err.Error())
	}
	return nil
}

func (c *Client) UpdateConversation(ctx context.Context, conversation *model.Conversation) *err_code.Status {
	if conversation.ConversationId == "" {
		return toErrStatus("assistant_conversation_update", "update conversation but conversationId empty")
	}
	if err := c.db.WithContext(ctx).Model(&model.Conversation{}).Where("conversation_id = ?", conversation.ConversationId).Updates(map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
	}).Error; err != nil {
		return toErrStatus("assistant_conversation_update", err.Error())
	}
	return nil
}

func (c *Client) DeleteConversation(ctx context.Context, conversationId string, userId, orgId string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithConversationId(conversationId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx).Model(&model.Conversation{})).Delete(&model.Conversation{}).Error; err != nil {
		return toErrStatus("assistant_conversation_delete", err.Error())
	}
	return nil

}

func (c *Client) GetConversationByAssistantID(ctx context.Context, assistantID uint32, conversationType string) (*model.Conversation, *err_code.Status) {
	conversation := &model.Conversation{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantID),
		sqlopt.WithConversationType(conversationType)).Apply(c.db.WithContext(ctx).Model(&model.Conversation{})).First(&conversation).Error; err != nil {
		return nil, toErrStatus("assistant_conversation_get", err.Error())
	}
	return conversation, nil
}

func (c *Client) GetConversation(ctx context.Context, conversationID, userID, orgID string) (*model.Conversation, *err_code.Status) {
	conversation := &model.Conversation{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithConversationId(conversationID),
		sqlopt.DataPerm(userID, orgID)).Apply(c.db.WithContext(ctx)).Model(&model.Conversation{}).First(&conversation).Error; err != nil {
		return nil, toErrStatus("assistant_conversation_get", err.Error())
	}
	return conversation, nil
}

func (c *Client) GetConversationList(ctx context.Context, assistantID uint32, conversationType, userID, orgID, searchText string, offset, limit int32) ([]*model.Conversation, int64, *err_code.Status) {
	var conversations []*model.Conversation
	var count int64
	query := sqlopt.SQLOptions(
		sqlopt.DataPerm(userID, orgID),
		sqlopt.WithConversationType(conversationType),
		sqlopt.WithTitleLike(searchText)).Apply(c.db.WithContext(ctx).Model(&model.Conversation{}))

	if assistantID != 0 {
		query = query.Where("assistant_id = ?", assistantID)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, toErrStatus("assistant_conversations_get_list", err.Error())
	}

	if err := query.Offset(int(offset)).Limit(int(limit)).Order("updated_at DESC").Find(&conversations).Error; err != nil {
		return nil, 0, toErrStatus("assistant_conversations_get_list", err.Error())
	}

	return conversations, count, nil
}
