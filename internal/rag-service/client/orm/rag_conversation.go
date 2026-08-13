package orm

import (
	"context"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"gorm.io/gorm"
)

// RagConversationNotExistKey 会话不存在 / 不属于请求方（归属过滤未命中）的统一 text key
const RagConversationNotExistKey = "rag_conversation_not_exist"

func (c *Client) CreateRagConversation(ctx context.Context, conversation *model.RagConversation) *err_code.Status {
	if conversation.ConversationID == "" {
		return toErrStatus("rag_conversation_create_err", "create conversation but conversation id empty")
	}
	if err := c.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return toErrStatus("rag_conversation_create_err", err.Error())
	}
	return nil
}

// TouchRagConversation 刷新 updated_at，使会话列表按最近聊天排序
func (c *Client) TouchRagConversation(ctx context.Context, conversationID string) *err_code.Status {
	if err := sqlopt.WithConversationID(conversationID).
		Apply(c.db.WithContext(ctx).Model(&model.RagConversation{})).
		Update("updated_at", time.Now().UnixMilli()).Error; err != nil {
		return toErrStatus("rag_conversation_update_err", err.Error())
	}
	return nil
}

// FetchRagConversation 按归属取会话，归属过滤未命中与不存在都返回 RagConversationNotExistKey
func (c *Client) FetchRagConversation(ctx context.Context, conversationID, ragID, userId, orgId string) (*model.RagConversation, *err_code.Status) {
	conversation := &model.RagConversation{}
	err := sqlopt.SQLOptions(append(ownerOpts(userId, orgId),
		sqlopt.WithConversationID(conversationID),
		sqlopt.WithRagID(ragID))...).
		Apply(c.db.WithContext(ctx)).First(conversation).Error
	if err != nil {
		// 文案里没有占位符，传参会让用户看到 %!(EXTRA string=...)
		return nil, toErrStatus(RagConversationNotExistKey)
	}
	return conversation, nil
}

// FetchDraftRagConversation 取草稿会话，草稿态每个知识问答仅维护一条
func (c *Client) FetchDraftRagConversation(ctx context.Context, ragID, userId, orgId string) (*model.RagConversation, *err_code.Status) {
	conversation := &model.RagConversation{}
	err := sqlopt.SQLOptions(append(ownerOpts(userId, orgId),
		sqlopt.WithRagID(ragID),
		sqlopt.WithConversationType(constant.ConversationTypeDraft))...).
		Apply(c.db.WithContext(ctx)).First(conversation).Error
	if err != nil {
		return nil, toErrStatus(RagConversationNotExistKey)
	}
	return conversation, nil
}

func (c *Client) DeleteRagConversation(ctx context.Context, conversationID, ragID, userId, orgId string) *err_code.Status {
	del := sqlopt.SQLOptions(append(ownerOpts(userId, orgId),
		sqlopt.WithConversationID(conversationID),
		sqlopt.WithRagID(ragID))...).
		Apply(c.db.WithContext(ctx)).Delete(&model.RagConversation{})
	if del.Error != nil {
		return toErrStatus("rag_conversation_delete_err", del.Error.Error())
	}
	// 删除是无条件成功的，行数为 0 说明会话不存在或不属于请求方
	if del.RowsAffected == 0 {
		return toErrStatus(RagConversationNotExistKey)
	}
	return nil
}

// DeleteRagConversationByRagID 删除某个知识问答下的全部会话，用于删除应用时清场。
// 归属已在删应用那步校验过，这里只按 ragID 删
func (c *Client) DeleteRagConversationByRagID(ctx context.Context, ragID string) *err_code.Status {
	if ragID == "" {
		return toErrStatus("rag_conversation_delete_err", "delete conversation but rag id empty")
	}
	if err := sqlopt.WithRagID(ragID).
		Apply(c.db.WithContext(ctx)).Delete(&model.RagConversation{}).Error; err != nil {
		return toErrStatus("rag_conversation_delete_err", err.Error())
	}
	return nil
}

func (c *Client) ListRagConversation(ctx context.Context, ragID, conversationType, userId, orgId, searchText string, offset, limit int32) ([]*model.RagConversation, int64, *err_code.Status) {
	var conversations []*model.RagConversation
	var total int64
	return conversations, total, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		query := sqlopt.SQLOptions(
			sqlopt.WithUserID(userId),
			sqlopt.WithOrgID(orgId),
			sqlopt.WithRagID(ragID),
			sqlopt.WithConversationType(conversationType),
			sqlopt.LikeTitle(searchText),
		).Apply(tx.WithContext(ctx).Model(&model.RagConversation{}))

		if err := query.Count(&total).Error; err != nil {
			return toErrStatus("rag_conversation_list_err", err.Error())
		}
		if err := query.Order("updated_at desc").Offset(int(offset)).Limit(int(limit)).
			Find(&conversations).Error; err != nil {
			return toErrStatus("rag_conversation_list_err", err.Error())
		}
		return nil
	})
}
