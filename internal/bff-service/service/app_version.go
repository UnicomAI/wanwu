package service

import (
	"sort"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetAppVersionList(ctx *gin.Context, userID, orgID, appType, appID string) (*response.ListResult, error) {
	var list []response.AppVersionInfo
	switch appType {
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		data, err := GetWorkflowVersionList(ctx, appID)
		if err != nil {
			return nil, err
		}
		for _, v := range data.VersionList {
			list = append(list, response.AppVersionInfo{
				Version:   v.Version,
				Desc:      v.Desc,
				CreatedAt: util.Time2Str(v.CreatedAt),
			})
		}

	case constant.AppTypeAgent:
		resp, err := assistant.AssistantSnapshotList(ctx.Request.Context(), &assistant_service.AssistantSnapshotListReq{
			AssistantId: appID,
			Identity: &assistant_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, snapshot := range resp.List {
			list = append(list, response.AppVersionInfo{
				Version:   snapshot.Version,
				Desc:      snapshot.Desc,
				CreatedAt: util.Time2Str(snapshot.CreateAt),
			})
		}

	case constant.AppTypeRag:
		resp, err := rag.ListPublishRagHistory(ctx.Request.Context(), &rag_service.ListPublishRagHistoryReq{
			RagId: appID,
			Identity: &rag_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, history := range resp.HistoryList {
			list = append(list, response.AppVersionInfo{
				Version:   history.Version,
				Desc:      history.Desc,
				CreatedAt: util.Time2Str(history.CreateAt),
			})
		}
	case constant.AppTypeSkill:
		resp, err := mcp.GetPublishCustomSkillHistoryList(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillHistoryListReq{
			SkillId: appID,
		})
		if err != nil {
			return nil, err
		}
		for _, history := range resp.HistoryList {
			list = append(list, response.AppVersionInfo{
				Version:   history.Version,
				Desc:      history.VersionDesc,
				CreatedAt: util.Time2Str(history.CreatedAt),
			})
		}
	default:
		return nil, grpc_util.ErrorStatus(errs.Code_BFFAppType)
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].CreatedAt > list[j].CreatedAt
	})
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func UpdateAppVersion(ctx *gin.Context, userID, orgID, appType, appID, description, publishType, extra string) error {
	switch appType {
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		if err := UpdateWorkflowVersionDesc(ctx, appID, description); err != nil {
			return err
		}
	case constant.AppTypeAgent:
		_, err := assistant.AssistantSnapshotUpdate(ctx.Request.Context(), &assistant_service.AssistantSnapshotUpdateReq{
			AssistantId: appID,
			Desc:        description,
			Extra:       extra,
			Identity: &assistant_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return err
		}
	case constant.AppTypeRag:
		_, err := rag.UpdatePublishRag(ctx.Request.Context(), &rag_service.UpdatePublishRagReq{
			RagId: appID,
			Desc:  description,
			Identity: &rag_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return err
		}
	case constant.AppTypeSkill:
		_, err := mcp.UpdatePublishCustomSkill(ctx.Request.Context(), &mcp_service.UpdatePublishCustomSkillReq{
			SkillId:     appID,
			VersionDesc: description,
		})
		if err != nil {
			return err
		}
	default:
		return grpc_util.ErrorStatus(errs.Code_BFFAppType)
	}
	_, err := app.PublishApp(ctx.Request.Context(), &app_service.PublishAppReq{
		AppId:       appID,
		AppType:     appType,
		PublishType: publishType,
		UserId:      userID,
		OrgId:       orgID,
	})
	return err
}

func RollbackAppVersion(ctx *gin.Context, userID, orgID, appType, appID, version string) error {
	switch appType {
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		if err := RollbackWorkflowVersion(ctx, appID, version); err != nil {
			return err
		}
		return nil
	case constant.AppTypeAgent:
		_, err := assistant.AssistantSnapshotRollback(ctx.Request.Context(), &assistant_service.AssistantSnapshotRollbackReq{
			AssistantId: appID,
			Version:     version,
			Identity: &assistant_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		return err
	case constant.AppTypeRag:
		_, err := rag.OverwriteRagDraft(ctx.Request.Context(), &rag_service.OverwriteRagDraftReq{
			RagId:   appID,
			Version: version,
		})
		return err
	case constant.AppTypeSkill:
		return rollbackCustomSkillWorkspace(ctx, appID, version)
	default:
		return grpc_util.ErrorStatus(errs.Code_BFFAppType)
	}
}

func GetAppLatestVersion(ctx *gin.Context, userID, orgID, appType, appID string) (*response.AppVersionInfo, error) {
	var ret response.AppVersionInfo
	switch appType {
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		version, desc, err := GetWorkflowVersion(ctx, appID)
		if err != nil {
			return nil, err
		}
		ret.Version = version
		ret.Desc = desc

	case constant.AppTypeAgent:
		resp, err := assistant.AssistantSnapshotLatest(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
			AssistantId: appID,
			Identity: &assistant_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return nil, err
		}
		ret.Version = resp.Version
		ret.Desc = resp.Desc
		ret.Extra = resp.Extra

	case constant.AppTypeRag:
		resp, err := rag.GetPublishRagDesc(ctx.Request.Context(), &rag_service.GetPublishRagDescReq{
			RagId: appID,
			Identity: &rag_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return nil, err
		}
		ret.Version = resp.Version
		ret.Desc = resp.Desc

	case constant.AppTypeSkill:
		resp, err := mcp.GetPublishCustomSkillByLatest(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByLatestReq{SkillId: appID})
		if err != nil {
			return nil, err
		}
		ret.Version = resp.GetVersion()
		ret.Desc = resp.GetVersionDesc()
		ret.CreatedAt = util.Time2Str(resp.GetCreatedAt())

	default:
		return nil, grpc_util.ErrorStatus(errs.Code_BFFAppType)
	}

	appInfo, _ := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{
		AppId:   appID,
		AppType: appType,
	}) // 可能没有发布过，不返回错误
	ret.PublishType = appInfo.GetPublishType()
	return &ret, nil
}
