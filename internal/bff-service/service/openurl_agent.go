package service

import (
	"fmt"
	"time"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func UrlConversationCreate(ctx *gin.Context, req request.UrlConversationCreateRequest, xCId, suffix string) (*response.ConversationCreateResp, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	resp, err := assistant.ConversationCreate(ctx, &assistant_service.ConversationCreateReq{
		AssistantId:      appUrlInfo.AppId,
		Prompt:           req.Prompt,
		ConversationType: constant.ConversationTypeWebURL,
		Identity: &assistant_service.Identity{
			UserId: xCId,
			OrgId:  appUrlInfo.OrgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.ConversationCreateResp{
		ConversationId: resp.ConversationId,
	}, nil
}

func UrlConversationDelete(ctx *gin.Context, userId, suffix string, req request.UrlConversationIdRequest) error {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return err
	}
	_, err = assistant.ConversationDelete(ctx, &assistant_service.ConversationDeleteReq{
		ConversationId: req.ConversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  appUrlInfo.OrgId,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func UrlConversationClear(ctx *gin.Context, userId, suffix string, req request.UrlConversationIdRequest) error {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return err
	}
	// 清空已发布智能体的对话ES数据（不删除会话ID）
	_, err = assistant.ClearConversationES(ctx.Request.Context(), &assistant_service.ClearConversationESReq{
		ConversationId: req.ConversationId,
		DetailId:       req.DetailId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  appUrlInfo.OrgId,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func GetAppUrlInfo(ctx *gin.Context, suffix string) (*response.AppUrlConfig, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	assistantInfo, err := assistant.AssistantSnapshotInfo(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
		AssistantId: appUrlInfo.AppId,
	})
	if err != nil {
		return nil, err
	}
	assistantResp, err := transAssistantResp2Model(ctx, assistantInfo)
	if err != nil {
		return nil, err
	}
	return &response.AppUrlConfig{
		Assistant:  assistantResp,
		AppUrlInfo: transAppUrlInfo(appUrlInfo),
	}, nil
}

func GetAppUrlModel(ctx *gin.Context, suffix string) (*response.ModelInfo, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	assistantInfo, err := assistant.AssistantSnapshotInfo(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
		AssistantId: appUrlInfo.AppId,
	})
	if err != nil {
		return nil, err
	}
	return GetModelById(ctx, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{
			ModelId: assistantInfo.ModelConfig.ModelId,
		},
	})
}

func GetUrlConversationList(ctx *gin.Context, xCId, suffix string, req request.GetUrlConversationListReq) (*response.ListResult, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	resp, err := assistant.GetConversationList(ctx.Request.Context(), &assistant_service.GetConversationListReq{
		PageSize:         int32(req.PageSize),
		PageNo:           int32(req.PageNo),
		ConversationType: constant.ConversationTypeWebURL,
		Identity: &assistant_service.Identity{
			UserId: xCId,
			OrgId:  appUrlInfo.OrgId,
		},
		AssistantId: appUrlInfo.AppId,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{Total: resp.Total, List: resp.Data}, nil
}

func GetUrlConversationDetailList(ctx *gin.Context, req request.UrlConversationIdRequest, xCId, suffix string) (*response.ListResult, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	resp, err := GetConversationDetailList(ctx, xCId, appUrlInfo.OrgId, request.ConversationGetDetailListRequest{
		ConversationId: req.ConversationId,
		PageSize:       1000,
		PageNo:         1,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{Total: resp.Total, List: resp.List}, nil
}

func AppUrlConversionStream(ctx *gin.Context, req request.UrlConversionStreamRequest, xCid, suffix string) error {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return err
	}
	streamParams := &agentChatStreamParams{startTime: time.Now()}
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		go func() {
			defer util.PrintPanicStack()
			RecordAppStatistic(detachedCtx, appUrlInfo.UserId, appUrlInfo.OrgId, appUrlInfo.AppId, constant.AppTypeAgent, !streamParams.hasErr, true, streamParams.firstTokenLatency, 0, constant.AppStatisticSourceWebUrl)
		}()
	}()

	chatCh, err := CallAssistantConversationStream(ctx, xCid, appUrlInfo.OrgId, xCid, request.ConversionStreamRequest{
		AssistantId:    appUrlInfo.AppId,
		ConversationId: req.ConversationId,
		FileInfo:       req.FileInfo,
		Prompt:         req.Prompt,
		SseHold:        true,
	}, true)
	if err != nil {
		streamParams.hasErr = true
		return err
	}
	// 2. 流式返回结果
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Agent] %v conversation %v recv", appUrlInfo.AppId, req.ConversationId), sse_util.DONE_MSG).
		WriteStream(chatCh, streamParams, buildAgentChatRespLineProcessor(), nil)
	return nil
}

func AppUrlGetPendingConversation(ctx *gin.Context, req request.UrlPendingConversionRequest, xCid, suffix string) (*response.PendingConversationResp, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	return GetPendingConversation(ctx, appUrlInfo.UserId, appUrlInfo.OrgId, xCid, request.PendingConversionRequest{
		ConversationId: req.ConversationId,
		Draft:          false,
		AssistantId:    appUrlInfo.AppId,
	})
}

func AppUrlConversionStreamConnect(ctx *gin.Context, req request.UrlConversionStreamConnectRequest, xCid, suffix string) error {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return err
	}
	return AssistantConversionStreamConnect(ctx, appUrlInfo.UserId, appUrlInfo.OrgId, xCid, request.ConversionStreamConnectRequest{
		AssistantId:    appUrlInfo.AppId,
		ConversationId: req.ConversationId,
	})
}

func AppUrlConversionStreamCancel(ctx *gin.Context, req request.UrlConversionStreamCancelRequest, xCid, suffix string) error {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return err
	}
	return AssistantConversionStreamCancel(ctx, appUrlInfo.UserId, appUrlInfo.OrgId, xCid, request.ConversionStreamCancelRequest{
		PendingConversionRequest: request.PendingConversionRequest{
			AssistantId:    appUrlInfo.AppId,
			ConversationId: req.ConversationId,
		},
	})
}

func getAppUrlInfoAndCheck(ctx *gin.Context, suffix string) (*app_service.AppUrlInfo, error) {
	appUrlInfo, err := app.GetAppUrlInfoBySuffix(ctx, &app_service.GetAppUrlInfoBySuffixReq{
		Suffix: suffix,
	})
	if err != nil {
		return nil, err
	}
	// 验证 Status 开关
	if !appUrlInfo.Status {
		return nil, grpc_util.ErrorStatus(err_code.Code_AppUrlStatus)
	}
	// 验证 expiredAt 是否已过期
	if appUrlInfo.ExpiredAt != 0 && time.Now().After(time.Unix(appUrlInfo.ExpiredAt/1000, 0)) {
		return nil, grpc_util.ErrorStatus(err_code.Code_AppUrlExpired)
	}
	// 设置UserID、OrgID（通过http调用工作流接口header需要传递）
	ctx.Set(gin_util.USER_ID, appUrlInfo.UserId)
	ctx.Set(gin_util.X_ORG_ID, appUrlInfo.OrgId)
	return appUrlInfo, nil
}

func AppUrlQuestionRecommend(ctx *gin.Context, req request.UrlQuestionRecommendRequest, xCid, suffix string) error {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return err
	}
	AssistantQuestionRecommend(ctx, xCid, appUrlInfo.OrgId, &request.QuestionRecommendRequest{
		Query:          req.Query,
		AssistantId:    appUrlInfo.AppId,
		ConversationId: req.ConversationId,
		Trial:          false,
	})
	return nil
}
