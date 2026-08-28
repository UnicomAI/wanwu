package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/api/proto/common"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service/es-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service/service-model"
	"github.com/UnicomAI/wanwu/pkg/constant"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ConversationCreate 创建对话
func (s *Service) ConversationCreate(ctx context.Context, req *assistant_service.ConversationCreateReq) (*assistant_service.ConversationCreateResp, error) {
	// 通过 UUID 获取自增 ID（带权限校验）
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	// 组装model参数
	conversation := &model.Conversation{
		AssistantId:      assistant.ID,
		ConversationId:   util.NewID(),
		Title:            req.Prompt, // 使用prompt作为初始标题
		ConversationType: req.ConversationType,
		UserId:           req.Identity.UserId,
		OrgId:            req.Identity.OrgId,
	}

	// 调用client方法创建对话
	if status := s.cli.CreateConversation(ctx, conversation); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	return &assistant_service.ConversationCreateResp{
		ConversationId: conversation.ConversationId,
	}, nil
}

// ConversationDelete 删除对话
func (s *Service) ConversationDelete(ctx context.Context, req *assistant_service.ConversationDeleteReq) (*emptypb.Empty, error) {
	conversation, status := s.cli.GetConversation(ctx, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	// 调用client方法删除对话
	if status = s.cli.DeleteConversation(ctx, req.ConversationId, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	conversationId := buildConversationId(conversation)
	// 逻辑删除es中的对话详情（标记 deleted=true）
	if _, err := s.LogicalDeleteFromES(ctx, &assistant_service.DeleteFromESReq{
		IndexName: "conversation_detail_infos_*",
		Conditions: map[string]string{
			"conversationId": conversationId,
			"userId.keyword": req.Identity.UserId,
		},
	}); err != nil {
		log.Errorf("从ES逻辑删除对话详情失败，conversationId: %s, error: %v", req.ConversationId, err)
	}

	return &emptypb.Empty{}, nil
}

// ClearConversationES 清空对话ES数据（不删除会话ID），支持按detailId删除单条
func (s *Service) ClearConversationES(ctx context.Context, req *assistant_service.ClearConversationESReq) (*emptypb.Empty, error) {
	conversation, status := s.cli.GetConversation(ctx, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	conversationId := buildConversationId(conversation)
	conditions := map[string]string{
		"conversationId": conversationId,
		"userId.keyword": req.Identity.UserId,
	}
	if req.DetailId != "" {
		conditions["id"] = req.DetailId
	}

	if _, err := s.LogicalDeleteFromES(ctx, &assistant_service.DeleteFromESReq{
		IndexName:  "conversation_detail_infos_*",
		Conditions: conditions,
	}); err != nil {
		if req.DetailId != "" {
			log.Errorf("从ES逻辑删除单条对话详情失败，detailId: %s, conversationId: %s, error: %v", req.DetailId, req.ConversationId, err)
		} else {
			log.Errorf("从ES逻辑删除对话详情失败，conversationId: %s, error: %v", req.ConversationId, err)
		}
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetConversationIdByAssistantId 获取对话记录id
func (s *Service) GetConversationIdByAssistantId(ctx context.Context, req *assistant_service.GetConversationIdByAssistantIdReq) (*assistant_service.ConversationIdResp, error) {
	// 通过 UUID 获取自增 ID（带权限校验）
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	// 调用client方法获取对话
	conversation, status := s.cli.GetConversationByAssistantID(ctx, assistant.ID, req.ConversationType, req.Identity.GetUserId(), req.Identity.GetOrgId())
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	return &assistant_service.ConversationIdResp{
		ConversationId: conversation.ConversationId,
	}, nil
}

// GetConversationList 对话列表
func (s *Service) GetConversationList(ctx context.Context, req *assistant_service.GetConversationListReq) (*assistant_service.GetConversationListResp, error) {
	// 计算offset
	offset := (req.PageNo - 1) * req.PageSize

	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	// 调用client方法获取对话列表
	conversations, total, status := s.cli.GetConversationList(ctx, assistant.ID, req.ConversationType, req.Identity.UserId, req.Identity.OrgId, req.SearchText, offset, req.PageSize)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	// 转换为响应格式
	var conversationInfos []*assistant_service.ConversationInfo
	for _, conversation := range conversations {
		conversationInfos = append(conversationInfos, &assistant_service.ConversationInfo{
			ConversationId: conversation.ConversationId,
			AssistantId:    assistant.UUID,
			Title:          conversation.Title,
			CreatedAt:      conversation.CreatedAt,
			UpdatedAt:      conversation.UpdatedAt,
		})
	}

	return &assistant_service.GetConversationListResp{
		Data:     conversationInfos,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// GetConversationDetailList 对话详情历史列表
func (s *Service) GetConversationDetailList(ctx context.Context, req *assistant_service.GetConversationDetailListReq) (*assistant_service.GetConversationDetailListResp, error) {
	conversation, status := s.cli.GetConversation(ctx, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	assistant, status := s.cli.GetAssistant(ctx, conversation.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	req.ConversationId = buildConversationId(conversation)
	total, details, err := s.GetConversationDetailInfoList(ctx, req)
	if err != nil {
		return nil, err
	}
	// 转换查询结果为响应格式
	var conversationDetails []*assistant_service.ConversionDetailInfo
	for _, detail := range details {
		conversationDetails = append(conversationDetails, &assistant_service.ConversionDetailInfo{
			Id:                   detail.Id,
			AssistantId:          assistant.UUID,
			ConversationId:       detail.ConversationId,
			Prompt:               detail.Prompt,
			SysPrompt:            detail.SysPrompt,
			Response:             detail.Response,
			SearchList:           detail.SearchList,
			CreatedBy:            detail.UserId, // 使用CreatedBy字段映射UserId
			CreatedAt:            detail.CreatedAt,
			UpdatedAt:            detail.UpdatedAt,
			RequestFiles:         transRequestFiles(detail.FileInfo),
			FileSize:             detail.FileSize,
			FileName:             detail.FileName,
			SubConversationList:  buildSubConversationList(detail.SubConversationDetailList, len(detail.ResponseList) == 0),
			ConversationResponse: buildConversationResponse(detail.Response, detail.ResponseList, len(detail.SubConversationDetailList)),
			ResponseFiles:        transAgentFiles(detail.ResponseFiles),
			Feedback:             detail.Feedback,
			FeedbackContent:      detail.FeedbackContent,
		})
	}

	log.Infof("成功从ES查询对话详情，conversationId: %s, userId: %s, 总数: %d, 返回: %d",
		req.ConversationId, req.Identity.UserId, total, len(conversationDetails))

	return &assistant_service.GetConversationDetailListResp{
		Data:     conversationDetails,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// InternalGetConversationDetailList 对话详情历史列表， 内部使用不校验权限
func (s *Service) InternalGetConversationDetailList(ctx context.Context, req *assistant_service.GetConversationDetailListReq) (*assistant_service.GetConversationDetailListResp, error) {
	total, details, err := s.GetConversationDetailInfoList(ctx, req)
	if err != nil {
		return nil, err
	}
	// 转换查询结果为响应格式
	var conversationDetails []*assistant_service.ConversionDetailInfo
	for _, detail := range details {
		conversationDetails = append(conversationDetails, &assistant_service.ConversionDetailInfo{
			Id:                   detail.Id,
			ConversationId:       detail.ConversationId,
			Prompt:               detail.Prompt,
			SysPrompt:            detail.SysPrompt,
			Response:             detail.Response,
			SearchList:           detail.SearchList,
			CreatedBy:            detail.UserId, // 使用CreatedBy字段映射UserId
			CreatedAt:            detail.CreatedAt,
			UpdatedAt:            detail.UpdatedAt,
			RequestFiles:         transRequestFiles(detail.FileInfo),
			FileSize:             detail.FileSize,
			FileName:             detail.FileName,
			SubConversationList:  buildSubConversationList(detail.SubConversationDetailList, len(detail.ResponseList) == 0),
			ConversationResponse: buildConversationResponse(detail.Response, detail.ResponseList, len(detail.SubConversationDetailList)),
			ResponseFiles:        transAgentFiles(detail.ResponseFiles),
			Feedback:             detail.Feedback,
			FeedbackContent:      detail.FeedbackContent,
		})
	}

	log.Infof("成功从ES查询对话详情，conversationId: %s, userId: %s, 总数: %d, 返回: %d",
		req.ConversationId, req.Identity.UserId, total, len(conversationDetails))

	return &assistant_service.GetConversationDetailListResp{
		Data:     conversationDetails,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

func (s *Service) AssistantConversionStream(req *assistant_service.AssistantConversionStreamReq, stream assistant_service.AssistantService_AssistantConversionStreamServer) error {
	// 刷新会话 updated_at，使会话列表按最近聊天排序
	var saveConversationId = req.ConversationId
	if req.ConversationId != "" {
		// 带归属查会话：查不到即不属于调用者，不能往他人会话里写消息
		conversation, status := s.cli.GetConversation(stream.Context(), req.ConversationId, req.Identity.GetUserId(), req.Identity.GetOrgId())
		if status != nil {
			return errStatus(errs.Code_AssistantConversationErr, status)
		}
		if status := s.cli.UpdateConversation(stream.Context(), &model.Conversation{ConversationId: req.ConversationId}); status != nil {
			log.Errorf("[Conversation] touch conversation updated_at failed, id: %s, err: %v", req.ConversationId, status)
		}
		if len(conversation.ConversationId) > 0 {
			saveConversationId = buildConversationId(conversation)
		}
	}
	//会话处理
	conversationProcessor := &service.ConversationProcessor{
		SSEWriter: sse_util.NewGrpcSSEWriter(stream, "AssistantConversionStreamNew", nil),
	}
	err := conversationProcessor.Process(stream.Context(), buildConversationParams(req, saveConversationId), buildAgentSendRequest(req))
	if err != nil {
		log.Errorf("Assistant服务处理智能体流式对话失败，assistantId: %s, error: %v", req.AssistantId, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "agent服务异常")
	}
	return nil
}

func (s *Service) GetConversationLog(ctx context.Context, req *assistant_service.ConversationLogReq) (*common.ConversationLog, error) {
	conversation, status := s.cli.GetConversation(ctx, req.ConversationId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	assistant, status := s.cli.GetAssistant(ctx, conversation.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	_, list, err := s.GetConversationDetailInfoList(ctx, &assistant_service.GetConversationDetailListReq{
		ConversationId: buildConversationId(conversation),
		PageSize:       1000,
		PageNo:         1,
		Identity:       req.Identity,
		ExcludeDeleted: false,
	})
	if err != nil {
		return nil, err
	}

	convLog := &common.ConversationLog{
		AppId:          assistant.UUID,
		AppType:        constant.AppTypeAgent,
		Title:          conversation.Title,
		ConversationId: req.ConversationId,
		Source:         req.SourceFrom,
		UserId:         req.Identity.UserId,
		OrgId:          req.Identity.OrgId,
		MessageCount:   int64(len(list)),
	}

	// 记录会话新旧标记到 Ext，供后续兼容旧版本 conversationId 查询
	if extBytes, err := json.Marshal(map[string]any{
		"conversation_id_mark": conversation.ID,               // 旧版本对话id（自增主键）
		"conversation_mark":    conversation.ConversationMark, // 0:新会话 1:旧版补充的conversationID
	}); err == nil {
		convLog.Ext = string(extBytes)
	}

	// 填充统计数据：错误数、平均耗时、平均首token时延
	fillConversationStatistic(convLog, list)

	// 填充已发布版本号（未发布时为空）；查询失败不阻断会话日志记录。
	if snapshot, verStatus := s.cli.GetAssistantSnapshot(ctx, conversation.AssistantId, ""); verStatus != nil {
		log.Errorf("get assistant published version failed, assistantId: %d, err: %v", conversation.AssistantId, verStatus)
	} else if snapshot != nil {
		convLog.Version = snapshot.Version
	}

	return convLog, nil
}

// GetConversationOwner 按 conversationId 查询会话归属（assistantId/userId/orgId），用于 bff 层归属校验
func (s *Service) GetConversationOwner(ctx context.Context, req *assistant_service.GetConversationOwnerReq) (*assistant_service.GetConversationOwnerResp, error) {
	conversation, status := s.cli.GetConversation(ctx, req.ConversationId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	assistant, status := s.cli.GetAssistant(ctx, conversation.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	return &assistant_service.GetConversationOwnerResp{
		AssistantId: assistant.UUID,
		UserId:      conversation.UserId,
		OrgId:       conversation.OrgId,
	}, nil
}

// GetConversationDetailInfoList 对话详情历史列表
func (s *Service) GetConversationDetailInfoList(ctx context.Context, req *assistant_service.GetConversationDetailListReq) (int64, []*model.ConversationDetails, error) {
	offset := int((req.PageNo - 1) * req.PageSize)
	pageParams := service_model.NewDetailPageParamsBuilder().
		WithConversationID(req.ConversationId).
		WithIdentity(req.Identity).
		WithPageParam(int32(offset), req.PageSize).
		WithHasDeleted(!req.ExcludeDeleted).
		WithOpenMinioUrl(true).
		Build()
	// 复用 SearchFromES 查询ES数据
	total, conversationDetails, err := es_service.SearchDetailPageList(ctx, pageParams)
	if err != nil {
		log.Errorf("从ES查询对话详情失败，conversationId: %s, userId: %s, error: %v", req.ConversationId, req.Identity.UserId, err)
		return 0, nil, fmt.Errorf("查询对话详情失败: %v", err)
	}
	return total, conversationDetails, nil
}

// extractFileInfos 从proto FileInfo中提取所有文件信息到model FileInfo
func extractFileInfos(fileInfos []*assistant_service.ConversionStreamFile) []model.FileInfo {
	if len(fileInfos) == 0 {
		return nil
	}
	var result []model.FileInfo
	for _, file := range fileInfos {
		if file != nil {
			result = append(result, model.FileInfo{
				FileName: file.FileName,
				FileSize: file.FileSize,
				FileUrl:  file.FileUrl,
			})
		}
	}
	return result
}

// extractFileUrls 从proto FileInfo中提取所有文件URL
func extractFileUrls(fileInfos []*assistant_service.ConversionStreamFile) []string {
	if len(fileInfos) == 0 {
		return nil
	}
	var fileUrls []string
	for _, file := range fileInfos {
		if file != nil && file.FileUrl != "" {
			fileUrls = append(fileUrls, file.FileUrl)
		}
	}
	return fileUrls
}

// transRequestFiles 将 model.FileInfo 转换为 assistant_service.RequestFile，并替换 fileUrl 为 minio 对外下载 url
func transRequestFiles(files []model.FileInfo) []*assistant_service.RequestFile {
	if files == nil {
		return nil
	}

	downloadURL := os.Getenv("MINIO_DOWNLOAD_URL")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")

	var result []*assistant_service.RequestFile
	for _, file := range files {
		// 替换 fileUrl 为 minio 对外下载 url
		replacedUrl := strings.Replace(file.FileUrl, "http://"+minioEndpoint+"/", downloadURL, 1)

		result = append(result, &assistant_service.RequestFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  replacedUrl,
		})
	}
	return result
}

// transAgentFiles 将 model.AgentFile 转换为 assistant_service.AgentFile
func transAgentFiles(files []*model.AgentFile) []*assistant_service.AgentFile {
	if files == nil {
		return nil
	}

	var result []*assistant_service.AgentFile
	for _, file := range files {
		if file == nil {
			continue
		}

		var metadata *assistant_service.AgentMeta
		if file.Metadata != nil {
			metadata = &assistant_service.AgentMeta{
				Desc:     file.Metadata.Desc,
				CreateAt: file.Metadata.CreateAt,
				Name:     file.Metadata.Name,
			}
		}

		result = append(result, &assistant_service.AgentFile{
			Name:     file.Name,
			Size:     int32(file.Size),
			FileUrl:  file.FileUrl,
			FileType: file.FileType,
			Metadata: metadata,
		})
	}
	return result
}

func buildConversationParams(req *assistant_service.AssistantConversionStreamReq, saveConversationId string) *service.ConversationParams {
	return &service.ConversationParams{
		AssistantId:        req.AssistantId,
		ConversationId:     req.ConversationId,
		SaveConversationId: saveConversationId,
		FileInfo:           extractFileInfos(req.FileInfo),
		OrgId:              req.Identity.OrgId,
		Query:              req.Prompt,
		UserId:             req.Identity.UserId,
		DetailId:           req.DetailId,
		SourceFrom:         req.SourceFrom,
	}
}

// buildAgentSendRequest 构建底层智能体能力接口请求体
func buildAgentSendRequest(req *assistant_service.AssistantConversionStreamReq) func(ctx context.Context) (string, *http.Response, context.CancelFunc, error) {
	var conversationID string
	// 历史聊天记录配置
	if req.ConversationId != "" {
		conversationID = req.ConversationId
	}
	// 底层智能体能力接口请求体
	chatReq := service.BuildAgentChatReq(&service.AgentUserInputParams{
		Input:          req.Prompt,
		Stream:         true,
		UploadFile:     extractFileUrls(req.FileInfo),
		ConversationId: conversationID,
		UserId:         req.Identity.UserId,
		OrgId:          req.Identity.OrgId,
		Draft:          req.Draft,
		DetailId:       req.DetailId,
	}, req.AssistantId)

	var monitorKey = "agent_chat_service"

	return func(ctx context.Context) (string, *http.Response, context.CancelFunc, error) {
		paramsBytes, err := json.Marshal(chatReq)
		if err != nil {
			return monitorKey, nil, nil, err
		}
		// 获取Assistant配置
		assistantConfig := config.Cfg().Assistant
		if assistantConfig.NewSseUrl == "" {
			return monitorKey, nil, nil, errors.New("智能体SSE URL配置错误")
		}
		params := &http_client.HttpRequestParams{
			Body:       paramsBytes,
			Timeout:    15 * time.Minute,
			Url:        assistantConfig.NewSseUrl,
			MonitorKey: monitorKey,
			LogLevel:   http_client.LogAll,
		}
		ctx, cancelFunction := context.WithTimeout(ctx, params.Timeout)
		result, err := http_client.Default().PostJsonOriResp(ctx, params)
		if err == nil {
			err = readHttpErr(result)
		}
		return monitorKey, result, cancelFunction, err
	}
}

func buildConversationResponse(response string, conversation []*model.ConversationResponse, startOrder int) []*assistant_service.ConversationResponse {
	if len(conversation) == 0 {
		return []*assistant_service.ConversationResponse{{Response: response, Order: int32(startOrder)}}
	}
	var retList []*assistant_service.ConversationResponse
	for _, resp := range conversation {
		retList = append(retList, &assistant_service.ConversationResponse{
			Response:    resp.Response,
			Order:       int32(resp.Order),
			ErrMessage:  resp.ErrMessage,
			ErrResponse: resp.ErrResponse,
		})
	}
	return retList
}

func buildSubConversationList(subConversationDetailList []*model.SubConversationDetail, oldData bool) []*assistant_service.SubConversation {
	if len(subConversationDetailList) == 0 {
		return make([]*assistant_service.SubConversation, 0)
	}
	var retList []*assistant_service.SubConversation
	for idx, detail := range subConversationDetailList {
		retList = append(retList, buildSubConversation(detail, idx, oldData))
	}
	return retList
}

func buildSubConversation(detail *model.SubConversationDetail, index int, oldData bool) *assistant_service.SubConversation {
	data := detail.EventData
	if data == nil {
		data = &model.SubEventData{}
	}
	var order = detail.Order
	if oldData {
		order = index
	}
	return &assistant_service.SubConversation{
		Response:         detail.Content,
		SearchList:       detail.SearchList,
		ParentId:         data.ParentId,
		Id:               data.Id,
		Name:             data.Name,
		Profile:          data.Profile,
		TimeCost:         data.TimeCost,
		Status:           int32(data.Status),
		ConversationType: string(detail.ConversationType),
		Order:            int32(order),
		ErrMessage:       data.ErrMessage,
	}
}

func readHttpErr(resp *http.Response) error {
	if resp != nil && resp.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Errorf("read http err body failed: %v", err)
			}
		}(resp.Body)

		// 读取响应体中的所有数据
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if len(body) > 0 {
			return errors.New(string(body))
		}
		return errors.New("http status code: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

func buildConversationId(conversation *model.Conversation) string {
	if conversation.ConversationMark == model.ConversationMarkOld {
		return util.Int2Str(conversation.ID)
	}
	return conversation.ConversationId
}

// fillConversationStatistic 从对话详情列表聚合统计数据写入 convLog
func fillConversationStatistic(convLog *common.ConversationLog, list []*model.ConversationDetails) {
	var totalCost, totalFirstTokenLatency, costCount, tokenCount int64
	for _, detail := range list {
		if detail == nil || detail.Statistic == nil {
			continue
		}
		switch detail.Feedback {
		case model.FeedBackLike:
			convLog.LikeCount++
		case model.FeedBackDislike:
			convLog.DislikeCount++
		}
		statistic := detail.Statistic
		// 统计主会话响应中的错误数量
		if len(statistic.ErrMessage) > 0 {
			convLog.ErrorCount++
		}
		// 累加耗时与首token时延，用于求平均
		if statistic.TotalCostTime > 0 {
			totalCost += statistic.TotalCostTime
			costCount++
		}
		if statistic.FirstTokenLatency > 0 {
			totalFirstTokenLatency += statistic.FirstTokenLatency
			tokenCount++
		}
	}
	if costCount > 0 {
		convLog.Costs = totalCost / costCount
	}
	if tokenCount > 0 {
		convLog.FirstTokenLatency = totalFirstTokenLatency / tokenCount
	}
}
