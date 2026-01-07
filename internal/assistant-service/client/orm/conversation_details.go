package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CreateConversationDetails(ctx context.Context, details *model.ConversationDetails) *err_code.Status {
	if details.ID != 0 {
		return toErrStatus("assistant_conversation_details_create", "create conversation details but id not 0")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := tx.Create(details).Error; err != nil {
			return toErrStatus("assistant_conversation_details_create", err.Error())
		}
		return nil
	})
}

func (c *Client) GetConversationDetailsList(ctx context.Context, conversationID, userID, orgID string, offset, limit int32) ([]*model.ConversationDetails, int64, *err_code.Status) {
	var conversations []*model.ConversationDetails
	var count int64
	return conversations, count, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		query := sqlopt.DataPerm(userID, orgID).Apply(tx.Model(&model.ConversationDetails{}))

		if conversationID != "" {
			query = query.Where("conversation_id = ?", conversationID)
		}

		if err := query.Count(&count).Error; err != nil {
			return toErrStatus("assistant_conversations_get_list", err.Error())
		}

		if err := query.Offset(int(offset)).Limit(int(limit)).Order("created_at DESC").Find(&conversations).Error; err != nil {
			return toErrStatus("assistant_conversations_get_list", err.Error())
		}

		return nil
	})
}
