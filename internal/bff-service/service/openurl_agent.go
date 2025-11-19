package service

import (
	"fmt"
	"time"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/gin-gonic/gin"
)

func UrlConversationCreate(ctx *gin.Context, req request.UrlConversationCreateRequest, xCId, suffix string) (*response.ConversationCreateResp, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	resp, err := assistant.ConversationCreate(ctx, &assistant_service.ConversationCreateReq{
		AssistantId: appUrlInfo.AppId,
		Prompt:      req.Prompt,
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

func GetAppUrlInfo(ctx *gin.Context, suffix string) (*response.AppUrlConfig, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	assistantInfo, err := assistant.GetAssistantInfo(ctx, &assistant_service.GetAssistantInfoReq{
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

func GetUrlConversationList(ctx *gin.Context, xCId, suffix string) (*response.ListResult, error) {
	appUrlInfo, err := getAppUrlInfoAndCheck(ctx, suffix)
	if err != nil {
		return nil, err
	}
	resp, err := assistant.GetConversationList(ctx, &assistant_service.GetConversationListReq{
		PageSize: 1000,
		PageNo:   1,
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
	// 1. CallAssistantConversationStream
	chatCh, err := CallAssistantConversationStream(ctx, xCid, appUrlInfo.OrgId, request.ConversionStreamRequest{
		AssistantId:    appUrlInfo.AppId,
		ConversationId: req.ConversationId,
		FileInfo:       []request.ConversionStreamFile{},
		Trial:          false,
		Prompt:         req.Prompt,
	})
	if err != nil {
		return err
	}
	// 2. 流式返回结果 [EN] 2. Streaming return results
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Agent] %v conversation %v recv", appUrlInfo.AppId, req.ConversationId), sse_util.DONE_MSG).
		WriteStream(chatCh, nil, buildAgentChatRespLineProcessor(), nil)
	return nil
}

func getAppUrlInfoAndCheck(ctx *gin.Context, suffix string) (*app_service.AppUrlInfo, error) {
	appUrlInfo, err := app.GetAppUrlInfoBySuffix(ctx, &app_service.GetAppUrlInfoBySuffixReq{
		Suffix: suffix,
	})
	if err != nil {
		return nil, err
	}
	// 验证 Status 开关 [EN] Verify Status switch
	if !appUrlInfo.Status {
		return nil, grpc_util.ErrorStatus(err_code.Code_AppUrlStatus)
	}
	// 验证 expiredAt 是否已过期 [EN] Verify expiredAt has expired
	if appUrlInfo.ExpiredAt != 0 && time.Now().After(time.Unix(appUrlInfo.ExpiredAt/1000, 0)) {
		return nil, grpc_util.ErrorStatus(err_code.Code_AppUrlExpired)
	}
	// 设置UserID、OrgID（通过http调用工作流接口header需要传递） [EN] Set UserID and OrgID (the header needs to be passed when calling the workflow interface through http)
	ctx.Set(gin_util.USER_ID, appUrlInfo.UserId)
	ctx.Set(gin_util.X_ORG_ID, appUrlInfo.OrgId)
	return appUrlInfo, nil
}
