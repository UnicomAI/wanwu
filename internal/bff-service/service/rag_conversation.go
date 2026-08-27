package service

import (
	"sort"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RagConversationCreate(ctx *gin.Context, userId, orgId string, req request.RagConversationCreateReq, conversationType string) (*response.RagConversationCreateResp, error) {
	resp, err := rag.CreateRagConversation(ctx.Request.Context(), &rag_service.CreateRagConversationReq{
		RagId:            req.RagID,
		Prompt:           req.Prompt,
		ConversationType: conversationType,
		Identity:         &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	return &response.RagConversationCreateResp{ConversationId: resp.ConversationId}, nil
}

// GetDraftRagConversationId 取草稿会话id；草稿尚未发起过对话时返回 nil，不算错误
func GetDraftRagConversationId(ctx *gin.Context, userId, orgId, ragId string) (string, error) {
	resp, err := rag.GetRagDraftConversation(ctx.Request.Context(), &rag_service.GetRagDraftConversationReq{
		RagId:    ragId,
		Identity: &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		if isRagConversationNotExistErr(err) {
			return "", nil
		}
		return "", err
	}
	return resp.ConversationId, nil
}

// isRagConversationNotExistErr 按错误码判定会话不存在
func isRagConversationNotExistErr(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.Code(err_code.Code_RagConversationNotExist)
}

// GetOrCreateDraftRagConversation 草稿态每个知识问答仅维护一条会话，没有则用本次提问建一条
func GetOrCreateDraftRagConversation(ctx *gin.Context, userId, orgId, ragId, prompt string) (string, error) {
	conversationId, err := GetDraftRagConversationId(ctx, userId, orgId, ragId)
	if err != nil {
		return "", err
	}
	if conversationId != "" {
		return conversationId, nil
	}
	resp, err := RagConversationCreate(ctx, userId, orgId, request.RagConversationCreateReq{
		RagID:  ragId,
		Prompt: prompt,
	}, constant.ConversationTypeDraft)
	if err != nil {
		return "", err
	}
	return resp.ConversationId, nil
}

func RagConversationDelete(ctx *gin.Context, userId, orgId string, req request.RagConversationIDReq) (interface{}, error) {
	// 传了 detailId 只删单条明细，会话保留
	if req.DetailID != "" {
		_, err := rag.ClearRagConversationDetail(ctx.Request.Context(), &rag_service.ClearRagConversationDetailReq{
			RagId:          req.RagID,
			ConversationId: req.ConversationID,
			DetailId:       req.DetailID,
			Identity:       &rag_service.Identity{UserId: userId, OrgId: orgId},
		})
		return nil, err
	}
	_, err := rag.DeleteRagConversation(ctx.Request.Context(), &rag_service.DeleteRagConversationReq{
		RagId:          req.RagID,
		ConversationId: req.ConversationID,
		Identity:       &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	return nil, err
}

func RagConversationClear(ctx *gin.Context, userId, orgId string, req request.RagConversationIDReq) (interface{}, error) {
	_, err := rag.ClearRagConversationDetail(ctx.Request.Context(), &rag_service.ClearRagConversationDetailReq{
		RagId:          req.RagID,
		ConversationId: req.ConversationID,
		DetailId:       req.DetailID,
		Identity:       &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	return nil, err
}

func RagConversationList(ctx *gin.Context, userId, orgId string, req request.RagConversationListReq, conversationType string) (response.PageResult, error) {
	resp, err := rag.ListRagConversation(ctx.Request.Context(), &rag_service.ListRagConversationReq{
		RagId:            req.RagID,
		ConversationType: conversationType,
		PageSize:         int32(req.PageSize),
		PageNo:           int32(req.PageNo),
		SearchText:       strings.TrimSpace(req.SearchText),
		Identity:         &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return response.PageResult{}, err
	}
	var list []response.RagConversationInfo
	for _, item := range resp.Data {
		list = append(list, response.RagConversationInfo{
			ConversationId: item.ConversationId,
			RagId:          item.RagId,
			Title:          item.Title,
			CreatedAt:      util.Time2Str(item.CreatedAt),
			UpdatedAt:      util.Time2Str(item.UpdatedAt),
		})
	}
	return response.PageResult{Total: resp.Total, List: list, PageNo: req.PageNo, PageSize: req.PageSize}, nil
}

func RagConversationDetailList(ctx *gin.Context, userId, orgId string, req request.RagConversationDetailListReq) (response.PageResult, error) {
	resp, err := rag.ListRagConversationDetail(ctx.Request.Context(), &rag_service.ListRagConversationDetailReq{
		RagId:          req.RagID,
		ConversationId: req.ConversationID,
		PageSize:       int32(req.PageSize),
		PageNo:         int32(req.PageNo),
		Identity:       &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return response.PageResult{}, err
	}
	var list []response.RagConversationDetailInfo
	for _, item := range resp.Data {
		list = append(list, response.RagConversationDetailInfo{
			DetailId:         item.DetailId,
			RagId:            item.RagId,
			ConversationId:   item.ConversationId,
			Prompt:           item.Prompt,
			Response:         item.Response,
			ReasoningContent: item.ReasoningContent,
			SearchList:       item.SearchList,
			QaSearchList:     item.QaSearchList,
			ErrMessage:       item.ErrMessage,
			QaErrMessage:     item.QaErrMessage,
			TraceId:          item.TraceId,
			RequestFiles:     buildRagRequestFiles(item.FileInfoList),
			CreatedAt:        item.CreatedAt,

			ReasoningTimeCost:    buildRagTimeCost(item.ReasoningTimeCost),
			SearchListTimeCost:   buildRagTimeCost(item.SearchListTimeCost),
			QaSearchListTimeCost: buildRagTimeCost(item.QaSearchListTimeCost),
			Feedback:             item.Feedback,
			FeedbackContent:      item.FeedbackContent,
		})
	}
	// ES 按时间倒序取分页，页内再翻正，从早到晚渲染
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt < list[j].CreatedAt })
	return response.PageResult{Total: resp.Total, List: list, PageNo: req.PageNo, PageSize: req.PageSize}, nil
}

func RagMessageFeedback(ctx *gin.Context, userId, orgId string, req request.RagMessageFeedbackReq) (*response.RagMessageFeedbackResp, error) {
	resp, err := rag.RagMessageFeedback(ctx.Request.Context(), &rag_service.RagMessageFeedbackReq{
		RagId:           req.RagID,
		ConversationId:  req.ConversationID,
		DetailId:        req.DetailID,
		FeedbackType:    req.FeedbackType,
		FeedbackContent: req.FeedbackContent,
		Identity:        &rag_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	return &response.RagMessageFeedbackResp{FeedbackType: resp.FeedbackType}, nil
}

// buildRagTimeCost proto 里是指针，阶段未发生时下游不返回，取零值
func buildRagTimeCost(timeCost *rag_service.RagTimeCost) response.RagTimeCost {
	return response.RagTimeCost{Start: timeCost.GetStart(), End: timeCost.GetEnd()}
}

func buildRagRequestFiles(fileInfoList []*rag_service.FileInfo) []response.RagRequestFile {
	retList := make([]response.RagRequestFile, 0, len(fileInfoList))
	for _, fileInfo := range fileInfoList {
		if fileInfo == nil {
			continue
		}
		retList = append(retList, response.RagRequestFile{
			FileName: fileInfo.FileName,
			FileSize: fileInfo.FileSize,
			FileUrl:  fileInfo.FileUrl,
		})
	}
	return retList
}
