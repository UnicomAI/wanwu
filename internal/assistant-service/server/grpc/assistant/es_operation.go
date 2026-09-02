package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	es_service "github.com/UnicomAI/wanwu/internal/assistant-service/service/es-service"
	minio_service "github.com/UnicomAI/wanwu/internal/assistant-service/service/minio-service"
	service_model "github.com/UnicomAI/wanwu/internal/assistant-service/service/service-model"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	feedbackNone    int32 = 0 // 无反馈
	feedbackLike    int32 = 1 // 点赞
	feedbackDislike int32 = 2 // 点踩
)

// SaveToES saves a document to ES.
func (s *Service) SaveToES(ctx context.Context, req *assistant_service.SaveToESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(req.DocJson), &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc json: %v", err)
	}

	if err := es.Assistant().IndexDocument(ctx, req.IndexName, doc); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteFromES deletes documents from ES by conditions.
func (s *Service) DeleteFromES(ctx context.Context, req *assistant_service.DeleteFromESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	//查询数据
	fields, _, _ := es.Assistant().SearchByFields(ctx, req.IndexName, conditions, 0, 1000, "desc")
	detailList := buildConversationDetails(fields)

	if err := es.Assistant().DeleteByFields(ctx, req.IndexName, conditions); err != nil {
		return nil, err
	}

	//删除历史记录中的minio数据
	asyncDeleteMinio(detailList)

	return &emptypb.Empty{}, nil
}

// LogicalDeleteFromES 逻辑删除：将匹配文档标记为 deleted=true，不物理删除。
// 供对话详情删除入口使用，配合 SearchFromES 的 exclude_deleted 过滤。
func (s *Service) LogicalDeleteFromES(ctx context.Context, req *assistant_service.DeleteFromESReq) (*emptypb.Empty, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	updates := map[string]interface{}{"deleted": true}
	if err := es.Assistant().UpdateByFields(ctx, req.IndexName, conditions, updates); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SearchFromES searches documents in ES by conditions.
func (s *Service) SearchFromES(ctx context.Context, req *assistant_service.SearchFromESReq) (*assistant_service.SearchFromESResp, error) {
	if req.IndexName == "" {
		return nil, fmt.Errorf("index name is empty")
	}

	conditions := make(map[string]interface{})
	for k, v := range req.Conditions {
		conditions[k] = v
	}

	from := int((req.PageNo - 1) * req.PageSize)
	size := int(req.PageSize)

	var docs []json.RawMessage
	var total int64
	var err error
	if req.ExcludeDeleted {
		docs, total, err = es.Assistant().SearchByFields(ctx, req.IndexName, conditions, from, size, req.SortOrder)
	} else {
		docs, total, err = es.Assistant().SearchByFieldsWithDelete(ctx, req.IndexName, conditions, from, size, req.SortOrder)
	}
	if err != nil {
		return nil, err
	}

	docJsonList := make([]string, 0, len(docs))
	for _, doc := range docs {
		// 替换minio地址为用户访问的服务器地址
		docStr := strings.ReplaceAll(string(doc), "http://"+config.Cfg().Minio.EndPoint, os.Getenv("MINIO_DOWNLOAD_URL"))
		docJsonList = append(docJsonList, docStr)
	}

	return &assistant_service.SearchFromESResp{
		DocJsonList: docJsonList,
		Total:       total,
	}, nil
}

// MessageFeedback 智能体消息点赞/点踩，语义为互斥，重复点击同类型则保留状态仅更新内容：
//   - 当前无反馈(0) -> 设为目标
//   - 当前已是目标  -> 保留状态，仅更新 feedbackContent
//   - 当前为另一项  -> 切换为目标(互斥)
//
// 一个 detailId 仅有一个对话者操作，故无需记录用户维度。
func (s *Service) MessageFeedback(ctx context.Context, req *assistant_service.MessageFeedbackReq) (*assistant_service.MessageFeedbackResp, error) {
	// 反馈类型处理
	if req.FeedbackType != feedbackLike && req.FeedbackType != feedbackDislike {
		req.FeedbackType = feedbackNone
		req.FeedbackContent = ""
	}
	if req.DetailId == "" {
		return nil, fmt.Errorf("detailId is empty")
	}

	conversation, status := s.cli.GetConversation(ctx, req.ConversationId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	//校验1.会话属于对应智能体，校验2：对话属于对应组织和人
	assistantRow, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	if conversation.AssistantId != assistantRow.ID || conversation.OrgId != req.Identity.OrgId || conversation.UserId != req.Identity.UserId {
		return nil, errCode(errs.Code_AssistantConversationErr)
	}

	pageParams := service_model.NewDetailPageParamsBuilder().
		WithDetailID(req.DetailId).
		WithConversationID(buildConversationId(conversation)).
		WithIdentity(req.Identity).
		WithPageParam(0, 1).
		Build()
	// 复用 SearchFromES 查询ES数据
	_, docs, err := es_service.SearchDetailPageList(ctx, pageParams)

	if err != nil {
		return nil, fmt.Errorf("查询消息反馈状态失败: %v", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("消息不存在, detailId: %s", req.DetailId)
	}

	newFeedback := req.FeedbackType
	newContent := req.FeedbackContent

	//更新数据：同类型时保留状态仅更新 feedbackContent，不同类型则切换(互斥)
	if err = es_service.UpdateDetailFeedback(ctx, pageParams, newFeedback, newContent); err != nil {
		return nil, fmt.Errorf("更新消息反馈状态失败: %v", err)
	}

	return &assistant_service.MessageFeedbackResp{
		FeedbackType: newFeedback,
	}, nil
}

func asyncDeleteMinio(detailList []*model.ConversationDetails) {
	marshal, _ := json.Marshal(detailList)
	log.Infof("开始异步删除文件数据, detailList %s", string(marshal))
	if len(detailList) > 0 {
		safe_go_util.SafeGo(func() {
			for _, detail := range detailList {
				if len(detail.FileInfo) > 0 {
					for _, info := range detail.FileInfo {
						log.Infof("异步删除输入文件：%s", info.FileUrl)
						_ = minio_service.DeleteFile(context.Background(), info.FileUrl)
					}
				}
				if len(detail.ResponseFiles) > 0 {
					for _, file := range detail.ResponseFiles {
						log.Infof("异步删除输出文件：%s", file.FileUrl)
						_ = minio_service.DeleteFile(context.Background(), file.FileUrl)
					}
				}
			}
		})
	}
}

func buildConversationDetails(fields []json.RawMessage) []*model.ConversationDetails {
	var detailList []*model.ConversationDetails
	if len(fields) > 0 {
		for _, field := range fields {
			var detail model.ConversationDetails
			if err := json.Unmarshal(field, &detail); err != nil {
				log.Warnf("解析ES文档失败: %v", err)
				continue
			}
			detailList = append(detailList, &detail)
		}
	}
	return detailList
}
