package service

import (
	"fmt"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

// --- internal wga assistant ---

// checkWgaAssistantConfig 校验wga智能体配置（用于更新配置）
// 通用智能体配置只支持单智能体
func checkWgaAssistantConfig(ctx *gin.Context, userId, orgId string, assistantList []*assistant_service.WgaConfigAssistant) error {
	if len(assistantList) == 0 {
		return nil
	}
	assistantIds := make([]string, 0, len(assistantList))
	for _, a := range assistantList {
		assistantIds = append(assistantIds, a.AssistantId)
	}
	validIds, assistantInfos, err := getValidAssistantIds(ctx, userId, orgId, assistantIds)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "assistant not found")
	}

	// 校验所有智能体
	for _, a := range assistantList {
		// 校验智能体是否存在
		if !validIds[a.AssistantId] {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("assistant not found: %s", a.AssistantId))
		}

		// 校验智能体是否已发布
		appInfo, err := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{
			AppId:   a.AssistantId,
			AppType: constant.AppTypeAgent,
		})
		if err != nil || appInfo.PublishType == "" {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("assistant not published: %s", a.AssistantId))
		}

		// 校验智能体类型：通用智能体只支持单智能体
		info := assistantInfos[a.AssistantId]
		if info != nil && info.Category != constant.AgentCategorySingle {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("assistant must be single agent: %s", a.AssistantId))
		}
	}
	return nil
}

func buildWgaAssistantOptions(ctx *gin.Context, userId, orgId string, assistantList []*assistant_service.WgaConfigAssistant) ([]wga_option.Option, error) {
	if len(assistantList) == 0 {
		return nil, nil
	}

	var assistantIds []string
	for _, a := range assistantList {
		assistantIds = append(assistantIds, a.AssistantId)
	}
	resp, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{
		AssistantIdList: assistantIds,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	var opts []wga_option.Option
	for _, a := range resp.AssistantInfos {
		if a.Info == nil {
			continue
		}
		schemaData, err := renderAgentChatProxySchema(a.Info.AppId, a.Info.Name, a.Info.Desc)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("render assistant(%s) openapi schema err: %v", a.Info.AppId, err))
		}
		doc, err := openapi3_util.LoadFromData(ctx.Request.Context(), schemaData)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("load assistant(%s) openapi schema err: %v", a.Info.AppId, err))
		}
		opts = append(opts, wga_option.WithExtraTool(wga_option.ExtraTool{
			OpenAPI3Schema: doc,
		}))
	}

	return opts, nil
}

// getValidAssistantIds 批量获取有效的智能体ID映射
// 返回: validIds - 有效ID映射, assistantInfos - 智能体信息映射, error
func getValidAssistantIds(ctx *gin.Context, userId, orgId string, assistantIds []string) (map[string]bool, map[string]*assistant_service.AssistantBrief, error) {
	if len(assistantIds) == 0 {
		return make(map[string]bool), make(map[string]*assistant_service.AssistantBrief), nil
	}
	assistantResp, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{
		AssistantIdList: assistantIds,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	validIds := make(map[string]bool)
	assistantInfos := make(map[string]*assistant_service.AssistantBrief)
	for _, info := range assistantResp.AssistantInfos {
		validIds[info.Info.AppId] = true
		assistantInfos[info.Info.AppId] = info
	}
	return validIds, assistantInfos, nil
}
