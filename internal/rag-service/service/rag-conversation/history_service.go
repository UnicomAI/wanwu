package rag_conversation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	minio_service "github.com/UnicomAI/wanwu/internal/rag-service/service/minio-service"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
)

// deleteScanSize 删除前捞附件地址的上限
const deleteScanSize = 1000

// SearchDetail 按会话查问答明细
func SearchDetail(ctx context.Context, conversationId, userId, orgId string, from, size int, sortOrder string) ([]*model.RagConversationDetail, int64, error) {
	if es.Rag() == nil || conversationId == "" {
		return nil, 0, nil
	}
	documents, total, err := es.Rag().SearchByFields(ctx, es.RagChatHistoryIndexPattern, map[string]interface{}{
		"conversationId": conversationId,
		"userId":         userId,
		"orgId":          orgId,
	}, from, size, sortOrder)
	if err != nil {
		return nil, 0, err
	}
	var details []*model.RagConversationDetail
	for _, doc := range documents {
		var detail model.RagConversationDetail
		if err := json.Unmarshal(doc, &detail); err != nil {
			log.Warnf("解析ES知识问答明细失败: %v", err)
			continue
		}
		details = append(details, &detail)
	}
	return details, total, nil
}

// UpdateDetailFeedback 更新单条问答明细的点赞/点踩状态，条件恒带 conversationId + userId，
// 传别的会话的 detailId 匹配不到文档
func UpdateDetailFeedback(ctx context.Context, conversationId, detailId, userId string, feedback int32, feedbackContent string) error {
	if es.Rag() == nil {
		return fmt.Errorf("ES未初始化")
	}
	conditions := map[string]interface{}{
		"conversationId": conversationId,
		"userId":         userId,
		"id":             detailId,
	}
	// UpdateByQuery 匹配不到文档也算成功，先查一次避免把不存在的明细报成反馈成功
	_, total, err := es.Rag().SearchByFields(ctx, es.RagChatHistoryIndexPattern, conditions, 0, 1, "desc")
	if err != nil {
		return err
	}
	if total == 0 {
		return fmt.Errorf("对话明细不存在，detailId: %s", detailId)
	}
	return es.Rag().UpdateByFields(ctx, es.RagChatHistoryIndexPattern, conditions, map[string]interface{}{
		"feedback":        feedback,
		"feedbackContent": feedbackContent,
	})
}

// DeleteDetail 删除会话下的问答明细，detailId 非空则只删单条。
// 条件恒带 conversationId + userId，传别的会话的 detailId 匹配不到文档，删不动
func DeleteDetail(ctx context.Context, conversationId, detailId, userId string) error {
	if es.Rag() == nil || conversationId == "" {
		return nil
	}
	conditions := map[string]interface{}{
		"conversationId": conversationId,
		"userId":         userId,
	}
	if detailId != "" {
		conditions["id"] = detailId
	}
	return deleteDetailByConditions(ctx, conditions)
}

// DeleteRagDetail 删除某个知识问答下的全部明细，用于删除应用时清场
func DeleteRagDetail(ctx context.Context, ragId string) error {
	if es.Rag() == nil || ragId == "" {
		return nil
	}
	return deleteDetailByConditions(ctx, map[string]interface{}{"ragId": ragId})
}

// deleteDetailByConditions 先把待删明细捞出来，删完 ES 再异步清掉它们引用的 minio 文件。
// 附件在落历史时已转永久存储，不跟着删就成了没人引用的垃圾
func deleteDetailByConditions(ctx context.Context, conditions map[string]interface{}) error {
	details := searchDetailForDelete(ctx, conditions)
	if err := es.Rag().DeleteByFields(ctx, es.RagChatHistoryIndexPattern, conditions); err != nil {
		return err
	}
	asyncDeleteMinio(details)
	return nil
}

// searchDetailForDelete 只为拿附件地址，查不到就当没有附件，不阻断删除
func searchDetailForDelete(ctx context.Context, conditions map[string]interface{}) []*model.RagConversationDetail {
	if !minio_service.Enabled() {
		return nil
	}
	documents, _, err := es.Rag().SearchByFields(ctx, es.RagChatHistoryIndexPattern, conditions, 0, deleteScanSize, "desc")
	if err != nil {
		log.Errorf("查询待删知识问答明细失败，conditions: %v, err: %v", conditions, err)
		return nil
	}
	var details []*model.RagConversationDetail
	for _, doc := range documents {
		var detail model.RagConversationDetail
		if err := json.Unmarshal(doc, &detail); err != nil {
			log.Warnf("解析ES知识问答明细失败: %v", err)
			continue
		}
		details = append(details, &detail)
	}
	return details
}

// asyncDeleteMinio 异步清理明细引用的附件，失败只记日志——ES 已经删了，
// 文件残留是可接受的，不能因此把已成功的删除报成失败
func asyncDeleteMinio(details []*model.RagConversationDetail) {
	if len(details) == 0 {
		return
	}
	safe_go_util.SafeGo(func() {
		for _, detail := range details {
			for _, fileInfo := range detail.FileInfo {
				if fileInfo == nil || fileInfo.FileUrl == "" {
					continue
				}
				log.Infof("异步删除知识问答附件：%s", fileInfo.FileUrl)
				_ = minio_service.DeleteFile(context.Background(), fileInfo.FileUrl)
			}
		}
	})
}
