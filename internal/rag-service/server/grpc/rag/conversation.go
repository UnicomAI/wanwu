package rag

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/orm"
	rag_conversation "github.com/UnicomAI/wanwu/internal/rag-service/service/rag-conversation"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// conversationErrStatus 归属过滤未命中与不存在统一报 RagConversationNotExist，不让调用方区分
func conversationErrStatus(code errs.Code, status *errs.Status) error {
	if status.TextKey == orm.RagConversationNotExistKey {
		code = errs.Code_RagConversationNotExist
	}
	return errStatus(code, status)
}

func (s *Service) CreateRagConversation(ctx context.Context, req *rag_service.CreateRagConversationReq) (*rag_service.CreateRagConversationResp, error) {
	userId, orgId := identityOf(req.Identity)
	conversation := &model.RagConversation{
		ConversationID:   util.NewID(),
		RagID:            req.RagId,
		Title:            req.Prompt, // 使用prompt作为初始标题
		ConversationType: req.ConversationType,
		PublicModel: model.PublicModel{
			UserID: userId,
			OrgID:  orgId,
		},
	}
	if status := s.cli.CreateRagConversation(ctx, conversation); status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	return &rag_service.CreateRagConversationResp{
		ConversationId: conversation.ConversationID,
	}, nil
}

func (s *Service) DeleteRagConversation(ctx context.Context, req *rag_service.DeleteRagConversationReq) (*emptypb.Empty, error) {
	userId, orgId := identityOf(req.Identity)
	if status := s.cli.DeleteRagConversation(ctx, req.ConversationId, req.RagId, userId, orgId); status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	// 会话行已删除，明细清理失败只记日志，不把已成功的删除报成失败
	if err := rag_conversation.DeleteDetail(ctx, req.ConversationId, "", userId); err != nil {
		log.Errorf("删除知识问答会话明细失败，conversationId: %s, err: %v", req.ConversationId, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ClearRagConversationDetail(ctx context.Context, req *rag_service.ClearRagConversationDetailReq) (*emptypb.Empty, error) {
	userId, orgId := identityOf(req.Identity)
	// 校验会话归属
	if _, status := s.cli.FetchRagConversation(ctx, req.ConversationId, req.RagId, userId, orgId); status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	if err := rag_conversation.DeleteDetail(ctx, req.ConversationId, req.DetailId, userId); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagConversationErr, "rag_conversation_clear_err", err.Error())
	}
	// 刷新 updated_at
	if status := s.cli.TouchRagConversation(ctx, req.ConversationId); status != nil {
		log.Errorf("刷新知识问答会话时间失败，conversationId: %s, err: %v", req.ConversationId, status)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetRagDraftConversation(ctx context.Context, req *rag_service.GetRagDraftConversationReq) (*rag_service.GetRagDraftConversationResp, error) {
	userId, orgId := identityOf(req.Identity)
	conversation, status := s.cli.FetchDraftRagConversation(ctx, req.RagId, userId, orgId)
	if status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	return &rag_service.GetRagDraftConversationResp{
		ConversationId: conversation.ConversationID,
	}, nil
}

// GetRagConversationOwner 只回归属事实，不做判断——判断留给 bff 中间件
func (s *Service) GetRagConversationOwner(ctx context.Context, req *rag_service.GetRagConversationOwnerReq) (*rag_service.GetRagConversationOwnerResp, error) {
	conversation, status := s.cli.FetchRagConversation(ctx, req.ConversationId, "", "", "")
	if status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	return &rag_service.GetRagConversationOwnerResp{
		RagId:  conversation.RagID,
		UserId: conversation.UserID,
		OrgId:  conversation.OrgID,
	}, nil
}

func (s *Service) ListRagConversation(ctx context.Context, req *rag_service.ListRagConversationReq) (*rag_service.ListRagConversationResp, error) {
	userId, orgId := identityOf(req.Identity)
	offset := (req.PageNo - 1) * req.PageSize
	conversations, total, status := s.cli.ListRagConversation(ctx, req.RagId, req.ConversationType, userId, orgId, req.SearchText, offset, req.PageSize)
	if status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	var data []*rag_service.RagConversationInfo
	for _, conversation := range conversations {
		data = append(data, &rag_service.RagConversationInfo{
			ConversationId: conversation.ConversationID,
			RagId:          conversation.RagID,
			Title:          conversation.Title,
			CreatedAt:      conversation.CreatedAt,
			UpdatedAt:      conversation.UpdatedAt,
		})
	}
	return &rag_service.ListRagConversationResp{
		Data:     data,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

func (s *Service) ListRagConversationDetail(ctx context.Context, req *rag_service.ListRagConversationDetailReq) (*rag_service.ListRagConversationDetailResp, error) {
	userId, orgId := identityOf(req.Identity)
	// 先校验会话归属，再查明细
	if _, status := s.cli.FetchRagConversation(ctx, req.ConversationId, req.RagId, userId, orgId); status != nil {
		return nil, conversationErrStatus(errs.Code_RagConversationErr, status)
	}
	offset := int((req.PageNo - 1) * req.PageSize)
	details, total, err := rag_conversation.SearchDetail(ctx, req.ConversationId, userId, orgId, offset, int(req.PageSize), "desc")
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_RagConversationErr, "rag_conversation_detail_err", err.Error())
	}
	var data []*rag_service.RagConversationDetailInfo
	for _, detail := range details {
		data = append(data, &rag_service.RagConversationDetailInfo{
			DetailId:         detail.Id,
			RagId:            detail.RagId,
			ConversationId:   detail.ConversationId,
			Prompt:           detail.Prompt,
			Response:         detail.Response,
			ReasoningContent: detail.ReasoningContent,
			SearchList:       detail.SearchList,
			QaSearchList:     detail.QaSearchList,
			ErrMessage:       detail.ErrMessage,
			QaErrMessage:     detail.QaErrMessage,
			FileInfoList:     buildDetailFileInfo(detail.FileInfo),
			CreatedAt:        detail.CreatedAt,

			ReasoningTimeCost:    buildDetailTimeCost(detail.Statistic.ReasoningTimeCost),
			SearchListTimeCost:   buildDetailTimeCost(detail.Statistic.SearchListTimeCost),
			QaSearchListTimeCost: buildDetailTimeCost(detail.Statistic.QaSearchListTimeCost),
		})
	}
	return &rag_service.ListRagConversationDetailResp{
		Data:     data,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

func buildDetailTimeCost(timeCost model.RagTimeCost) *rag_service.RagTimeCost {
	return &rag_service.RagTimeCost{Start: timeCost.Start, End: timeCost.End}
}

func buildDetailFileInfo(fileInfoList []*model.RagFileInfo) []*rag_service.FileInfo {
	retList := make([]*rag_service.FileInfo, 0, len(fileInfoList))
	for _, fileInfo := range fileInfoList {
		if fileInfo == nil {
			continue
		}
		retList = append(retList, &rag_service.FileInfo{
			FileName: fileInfo.FileName,
			FileSize: fileInfo.FileSize,
			FileUrl:  fileInfo.FileUrl,
		})
	}
	return retList
}
