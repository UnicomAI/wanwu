package service

import (
	"context"
	"encoding/json"
	"fmt"
	minio_service "github.com/UnicomAI/wanwu/internal/assistant-service/service/minio-service"
	"github.com/UnicomAI/wanwu/pkg/util"
	"net/http"
	"sort"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service/conversation"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

const (
	esTimeout = 1 * time.Minute
)

type ConversationParams struct {
	AssistantId    string           `json:"assistantId"`
	DetailId       string           `json:"detailId"`
	ConversationId string           `json:"conversationId"`
	UserId         string           `json:"userId"`
	OrgId          string           `json:"orgId"`
	Query          string           `json:"query"`
	FileInfo       []model.FileInfo `json:"fileInfo"`
}

type AgentChatResp struct {
	Code       int           `json:"code"`
	Message    string        `json:"message"`
	Response   string        `json:"response"`
	SearchList []interface{} `json:"search_list"`
	Finish     int           `json:"finish"`
	EventType  int           `json:"eventType"`
	EventData  interface{}   `json:"eventData"`
}

type ConversationProcessor struct {
	SSEWriter *sse_util.SSEWriter[assistant_service.AssistantConversionStreamResp]
}

func (cp *ConversationProcessor) Process(ctx context.Context, req *ConversationParams, sendRequest func(ctx context.Context) (string, *http.Response, context.CancelFunc, error)) (err error) {
	var conversationResp = conversation.CreateConversationResp()
	defer func() {
		if err != nil {
			//错误信息通知
			_ = cp.SSEWriter.WriteLine(assistant_service.AssistantConversionStreamResp{
				Content: buildErrMsg(err),
			}, false, nil, nil)
			//处理异常
			processError(ctx, req, conversationResp, err)
		}
		//保存会话
		saveConversation(ctx, req, conversationResp, req.DetailId)
	}()
	//1.执行请求
	businessKey, sseResp, cancel, err := sendRequest(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	//2.读取结果
	SSEReader := &sse_util.SSEReader[string]{
		BusinessKey:    businessKey,
		StreamReceiver: sse_util.NewHttpStreamReceiver(sseResp),
	}
	stream, err := SSEReader.ReadStream(ctx)
	if err != nil {
		return err
	}
	//3.回写结果
	err = cp.SSEWriter.WriteStream(stream, nil, cp.conversationLineBuilder(conversationResp), nil)
	return err
}

// conversationLineBuilder 构造agent对话结果行处理器
func (cp *ConversationProcessor) conversationLineBuilder(conversationResp *conversation.ConversationResp) func(sse_util.SSEWriterClient[assistant_service.AssistantConversionStreamResp], string, interface{}) (assistant_service.AssistantConversionStreamResp, bool, error) {
	return func(s sse_util.SSEWriterClient[assistant_service.AssistantConversionStreamResp], strLine string, streamContextParams interface{}) (assistant_service.AssistantConversionStreamResp, bool, error) {
		err := conversation.BuildConversationResp(conversationResp, strLine)
		if err != nil {
			log.Errorf("BuildConversationResp error %s", err)
		}
		return assistant_service.AssistantConversionStreamResp{
			Content: strLine,
		}, false, nil
	}
}

// 处理会话异常
func processError(ctx context.Context, req *ConversationParams, conversationResp *conversation.ConversationResp, err error) {
	log.Errorf("[Conversation] err: %v", err)
	//错误会话异常
	//判断异常是不是敏感词
	sensitiveMsg := redis.GetSensitiveConversation(req.ConversationId, req.DetailId)
	if len(sensitiveMsg) > 0 {
		conversationResp.SensitiveResponse(sensitiveMsg)
	} else {
		conversationResp.ErrorResponse(err)
	}
}

// 构建错误信息,todo 后续考虑创建枚举明细错误信息
func buildErrMsg(err error) string {
	var agentChatResp = &AgentChatResp{
		Code:     1,
		Message:  err.Error(),
		Response: "智能体处理异常，请稍后重试",
		Finish:   1,
	}
	respString, errR := json.Marshal(agentChatResp)
	if errR != nil {
		log.Errorf("buildErrMsg error: %v", errR)
		return ""
	}
	return string(respString)
}

// 使用独立上下文保存对话的辅助函数
func saveConversation(originalCtx context.Context, req *ConversationParams, conversationResp *conversation.ConversationResp, detailId string) {
	//增加日志
	if originalCtx.Err() != nil {
		err := originalCtx.Err()
		log.Errorf("[Conversation] context err: %v", err)
	}
	//todo delete
	log.Infof("[Conversation] fullResponse: %s", conversationResp.Response())

	if len(req.ConversationId) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(trace_util.DetachContext(originalCtx), esTimeout)
	defer cancel()

	if err := saveConversationDetailToES(ctx, req, conversationResp, detailId); err != nil {
		log.Errorf("保存聊天记录到ES失败，assistantId: %s, conversationId: %s, error: %v",
			req.AssistantId, req.ConversationId, err)
	}
}

// saveConversationDetailToES 保存聊天记录到ES
func saveConversationDetailToES(ctx context.Context, req *ConversationParams, conversationResp *conversation.ConversationResp, detailId string) error {
	// 根据当前时间生成索引名称，格式为conversation_detail_infos_YYYYMM
	now := time.Now()
	indexName := fmt.Sprintf("conversation_detail_infos_%d%02d", now.Year(), now.Month())

	// 组装ConversationDetails数据，使用传入的detailId
	conversationDetail := buildConversationDetail(req, conversationResp, now.UnixMilli(), detailId)
	// 写入ES
	if err := es.Assistant().IndexDocument(ctx, indexName, conversationDetail); err != nil {
		return fmt.Errorf("写入ES失败: %v", err)
	}

	log.Infof("成功保存聊天记录到ES，索引: %s, assistantId: %s, conversationId: %s, detailId: %s",
		indexName, req.AssistantId, req.ConversationId, detailId)
	return nil
}

func buildConversationDetail(req *ConversationParams, conversationResp *conversation.ConversationResp, nowMilli int64, detailId string) *model.ConversationDetails {
	if len(conversationResp.SensitiveMessage) > 0 {
		conversationResp.CurrentData = conversationResp.SensitiveMessage
	}
	err := conversation.FinishConversationResp(conversationResp)
	if err != nil {
		log.Errorf("FinishConversationResp error: %v", err)
	}
	return &model.ConversationDetails{
		Id:                        detailId,
		AssistantId:               req.AssistantId,
		ConversationId:            req.ConversationId,
		Prompt:                    req.Query,
		FileInfo:                  restoreFileList(req.FileInfo),
		ResponseList:              conversationResp.ResponseList(),
		Response:                  conversationResp.Response(),
		SearchList:                conversationResp.References(),
		UserId:                    req.UserId,
		OrgId:                     req.OrgId,
		CreatedAt:                 nowMilli,
		UpdatedAt:                 nowMilli,
		SubConversationDetailList: buildSubConversationDetailList(conversationResp),
		ResponseFiles:             conversationResp.ResponseFiles,
	}
}

func buildSubConversationDetailList(conversationResp *conversation.ConversationResp) []*model.SubConversationDetail {
	if len(conversationResp.ConversationEventMap) == 0 {
		return make([]*model.SubConversationDetail, 0)
	}
	var dataList []*conversation.ConversationResp
	for _, subConversationResp := range conversationResp.ConversationEventMap {
		dataList = append(dataList, subConversationResp)
	}
	// 对切片进行排序
	sort.Slice(dataList, func(i, j int) bool {
		// Order值小的时间更早，排在前面
		return dataList[i].Order < dataList[j].Order
	})
	var retList []*model.SubConversationDetail
	for _, subConversationResp := range dataList {
		eventData := subConversationResp.EventData
		if eventData == nil {
			continue
		}
		if eventData.Status != model.EventEndStatus &&
			eventData.Status != model.EventFailStatus {
			eventData.Status = model.EventEndStatus // 暂定设置成结束状态
		}
		retList = append(retList, &model.SubConversationDetail{
			BusinessId:       eventData.Id,
			ConversationType: buildConversationType(subConversationResp.EventType),
			Order:            subConversationResp.Order,
			Content:          subConversationResp.Response(),
			SearchList:       subConversationResp.References(),
			EventData:        eventData,
		})
	}
	return retList
}

func buildConversationType(eventType int) model.ConversationType {
	switch eventType {
	case conversation.SubAgentEventType:
		return model.SubAgent
	case conversation.KnowledgeEventType:
		return model.AgentKnowledge
	case conversation.ToolEventType:
		return model.AgentTool
	case conversation.ThinkingEventType:
		return model.AgentThink
	case conversation.SkillEventType:
		return model.AgentSkill
	case conversation.SubTextEventType:
		return model.AgentSubText

	default:
		return model.SubAgent
	}
}

func restoreFileList(agentFileList []model.FileInfo) (retFileList []model.FileInfo) {
	if len(agentFileList) == 0 {
		return agentFileList
	}
	defer util.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			retFileList = agentFileList
			return
		}
	})
	for _, file := range agentFileList {
		retFileList = append(retFileList, model.FileInfo{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  restoreFile(file.FileUrl),
		})
	}
	return
}

// restoreFile 重新存储文件地址，原来是临时的，需要做永久存储
func restoreFile(fileUrl string) string {
	newFileUrl, _, _, err := minio_service.CopyFile(context.Background(), fileUrl, "file-not-expire", true)
	if err != nil || len(newFileUrl) == 0 {
		return fileUrl
	}
	return newFileUrl
}
