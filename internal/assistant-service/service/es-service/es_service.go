package es_service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service/service-model"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	"os"
	"strings"
)

const (
	detailIndexName = "conversation_detail_infos_*"
)

func SearchDetailPageList(ctx context.Context, detailPageParams *service_model.DetailPageParams) (int64, []*model.ConversationDetails, error) {
	fieldConditions := buildSearchConditions(detailPageParams)
	var docs []json.RawMessage
	var total int64
	var err error
	if detailPageParams.HasDeleted {
		docs, total, err = es.Assistant().SearchByFieldsWithDelete(ctx, detailIndexName, fieldConditions, detailPageParams.PageNum, detailPageParams.PageSize, "desc")
	} else {
		docs, total, err = es.Assistant().SearchByFields(ctx, detailIndexName, fieldConditions, detailPageParams.PageNum, detailPageParams.PageSize, "desc")
	}
	if err != nil {
		return 0, nil, err
	}
	var conversationList []*model.ConversationDetails
	for _, doc := range docs {
		docStr := string(doc)
		if detailPageParams.OpenMinioUrl {
			// 替换minio地址为用户访问的服务器地址
			docStr = strings.ReplaceAll(docStr, "http://"+config.Cfg().Minio.EndPoint, os.Getenv("MINIO_DOWNLOAD_URL"))
		}
		var detail model.ConversationDetails
		if err := json.Unmarshal([]byte(docStr), &detail); err != nil {
			log.Warnf("Assistant服务解析ES历史聊天记录失败: %v", err)
			continue
		}
		conversationList = append(conversationList, &detail)
	}

	return total, conversationList, nil
}

func UpdateDetailFeedback(ctx context.Context, detailPageParams *service_model.DetailPageParams, feedback int32, feedbackContent string) error {
	searchConditions := buildSearchConditions(detailPageParams)
	// 写回新状态
	updates := map[string]interface{}{
		"feedback":        feedback,
		"feedbackContent": feedbackContent,
	}
	if err := es.Assistant().UpdateByFields(ctx, detailIndexName, searchConditions, updates); err != nil {
		return fmt.Errorf("更新消息反馈状态失败: %v", err)
	}
	return nil
}

// buildSearchConditions 构造查询条件
func buildSearchConditions(detailParams *service_model.DetailPageParams) map[string]interface{} {
	var paramsMaps = make(map[string]interface{})
	paramsMaps["userId.keyword"] = detailParams.UserID
	paramsMaps["orgId.keyword"] = detailParams.OrgID

	if len(detailParams.DetailID) > 0 {
		paramsMaps["id"] = detailParams.DetailID
	}
	if len(detailParams.ConversationID) > 0 {
		paramsMaps["conversationId"] = detailParams.ConversationID
	}
	return paramsMaps
}
