package service

import (
	"sort"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func DeleteAppSpaceApp(ctx *gin.Context, userId, orgId, appId, appType string) error {
	// delete publish app
	_, err := app.DeleteApp(ctx.Request.Context(), &app_service.DeleteAppReq{
		AppId:   appId,
		AppType: appType,
	})
	if err != nil {
		return err
	}
	// delete app
	switch appType {
	case constant.AppTypeRag:
		_, err = rag.DeleteRag(ctx.Request.Context(), &rag_service.RagDeleteReq{
			RagId: appId,
		})
	case constant.AppTypeAgent:
		_, err = assistant.AssistantDelete(ctx.Request.Context(), &assistant_service.AssistantDeleteReq{
			AssistantId: appId,
		})
	case constant.AppTypeWorkflow:
		_, err = assistant.AssistantWorkFlowDeleteByWorkflowId(ctx.Request.Context(), &assistant_service.AssistantWorkFlowDeleteByWorkflowIdReq{
			WorkflowId: appId,
		})
		if err != nil {
			return err
		}
		err = DeleteWorkflow(ctx, orgId, appId)
	case constant.AppTypeChatflow:
		_, err = assistant.AssistantWorkFlowDeleteByWorkflowId(ctx.Request.Context(), &assistant_service.AssistantWorkFlowDeleteByWorkflowIdReq{
			WorkflowId: appId,
		})
		if err != nil {
			return err
		}
		// 复用工作流的删除接口
		err = DeleteWorkflow(ctx, orgId, appId)
	}
	return err
}

func GetAppSpaceAppList(ctx *gin.Context, userId, orgId, name, appType string) (*response.ListResult, error) {
	var ret []response.AppBriefInfo
	if appType == constant.AppTypeRag {
		resp, err := rag.ListRag(ctx.Request.Context(), &rag_service.RagListReq{
			Name: name,
			Identity: &rag_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, ragInfo := range resp.RagInfos {
			ret = append(ret, appBriefProto2Model(ctx, ragInfo, 0))
		}
	}
	if appType == constant.AppTypeAgent {
		resp, err := assistant.GetAssistantListMyAll(ctx.Request.Context(), &assistant_service.GetAssistantListMyAllReq{
			Name: name,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, assistantInfo := range resp.AssistantInfos {
			ret = append(ret, appBriefProto2Model(ctx, assistantInfo.Info, assistantInfo.Category))
		}
	}
	return fillAppPublishInfo(ctx, userId, orgId, ret)
}

// GetWorkflowAndChatflowList 获取工作流和对话流列表
// appType: 空=全部, workflow=工作流, chatflow=对话流
func GetWorkflowAndChatflowList(ctx *gin.Context, userId, orgId, name, appType string) (*response.ListResult, error) {
	var ret []response.AppBriefInfo

	// 根据 appType 决定获取哪些列表
	if appType == "" || appType == constant.AppTypeWorkflow {
		// 获取工作流列表
		workflowResp, err := ListWorkflow(ctx, orgId, name, constant.AppTypeWorkflow)
		if err != nil {
			return nil, err
		}
		for _, workflowInfo := range workflowResp.Workflows {
			ret = append(ret, cozeWorkflowInfo2Model(workflowInfo))
		}
	}

	if appType == "" || appType == constant.AppTypeChatflow {
		// 获取对话流列表
		chatflowResp, err := ListWorkflow(ctx, orgId, name, constant.AppTypeChatflow)
		if err != nil {
			return nil, err
		}
		for _, chatflowInfo := range chatflowResp.Workflows {
			ret = append(ret, cozeChatflowInfo2Model(chatflowInfo))
		}
	}
	return fillAppPublishInfo(ctx, userId, orgId, ret)
}

func PublishApp(ctx *gin.Context, userId, orgId string, req request.PublishAppRequest) error {
	var rollbackPublishPackage func()
	publishPackageCommitted := true
	defer func() {
		if !publishPackageCommitted && rollbackPublishPackage != nil {
			rollbackPublishPackage()
		}
	}()

	if req.AppType == constant.AppTypeWorkflow || req.AppType == constant.AppTypeChatflow {
		if err := PublishWorkflow(ctx, orgId, req.AppId, req.Version, req.Desc); err != nil {
			return err
		}
	}
	if req.AppType == constant.AppTypeAgent {
		resp, _ := assistant.AssistantSnapshotLatest(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
			AssistantId: req.AppId,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if resp != nil {
			if err := util.IsVersionGreaterThan(req.Version, resp.Version); err != nil {
				return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_app_publish_version", resp.Version, req.Version, err.Error())
			}
		}
		_, err := assistant.AssistantSnapshotCreate(ctx.Request.Context(), &assistant_service.AssistantSnapshotReq{
			AssistantId: req.AppId,
			Version:     req.Version,
			Desc:        req.Desc,
			Extra:       req.Extra,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			return err
		}
	}
	if req.AppType == constant.AppTypeRag {
		resp, _ := rag.GetPublishRagDesc(ctx.Request.Context(), &rag_service.GetPublishRagDescReq{
			RagId: req.AppId,
			Identity: &rag_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if resp != nil {
			if err := util.IsVersionGreaterThan(req.Version, resp.Version); err != nil {
				return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_app_publish_version", resp.Version, req.Version, err.Error())
			}
		}
		_, err := rag.PublishRag(ctx.Request.Context(), &rag_service.PublishRagReq{
			RagId:   req.AppId,
			Version: req.Version,
			Desc:    req.Desc,
			Identity: &rag_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			return err
		}
	}
	if req.AppType == constant.AppTypeSkill {
		resp, err := mcp.GetPublishCustomSkillByLatest(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByLatestReq{SkillId: req.AppId})
		if err == nil && resp != nil && resp.GetVersion() != "" {
			if err := util.IsVersionGreaterThan(req.Version, resp.GetVersion()); err != nil {
				return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_app_publish_version", resp.GetVersion(), req.Version, err.Error())
			}
		}
		objectPath, markdown, rollback, err := buildCustomSkillPublishPackage(ctx, userId, orgId, req.AppId, req.Version)
		if err != nil {
			return err
		}
		rollbackPublishPackage = rollback
		publishPackageCommitted = false
		_, err = mcp.CreatePublishCustomSkill(ctx.Request.Context(), &mcp_service.PublishCustomSkillReq{
			SkillId:     req.AppId,
			Version:     req.Version,
			VersionDesc: req.Desc,
			Identity:    &mcp_service.Identity{UserId: userId, OrgId: orgId},
			ObjectPath:  objectPath,
			Markdown:    markdown,
		})
		if err != nil {
			return err
		}
	}
	_, err := app.PublishApp(ctx.Request.Context(), &app_service.PublishAppReq{
		AppId:       req.AppId,
		AppType:     req.AppType,
		PublishType: req.PublishType,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		return err
	}
	publishPackageCommitted = true
	return nil
}

func UnPublishApp(ctx *gin.Context, userId, orgId string, req request.UnPublishAppRequest) error {
	if req.AppType == constant.AppTypeWorkflow {
		_, err := assistant.AssistantWorkFlowDeleteByWorkflowId(ctx.Request.Context(), &assistant_service.AssistantWorkFlowDeleteByWorkflowIdReq{
			WorkflowId: req.AppId,
		})
		if err != nil {
			return err
		}
	}
	_, err := app.UnPublishApp(ctx.Request.Context(), &app_service.UnPublishAppReq{
		AppId:   req.AppId,
		AppType: req.AppType,
		UserId:  userId,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetAppList(ctx *gin.Context, userId, orgId, appType string) (*response.ListResult, error) {
	resp, err := app.GetAppList(ctx.Request.Context(), &app_service.GetAppListReq{
		OrgIds:  []string{orgId},
		UserIds: []string{userId},
		AppType: appType,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  resp.Infos,
		Total: int64(len(resp.Infos)),
	}, nil
}

// getAppVersionBatch 为已发布应用拉取各域最新版本号，供列表填充 Version。
// publishAppMap 的 key 为 appId，value 为 app-service 侧应用信息；仅 PublishType 非空时参与版本查询。
// 任一批次下游失败仅打日志、不中断列表；对应 app 的 version 可能为空（降级），与列表接口仍返回成功一致。
func getAppVersionBatch(ctx *gin.Context, userId, orgId string, publishAppMap map[string]*app_service.AppInfo) map[string]string {
	versionMap := make(map[string]string)

	var ragIds, assistantIds, workflowIds, chatflowIds, skillIds []string
	for appId, appInfo := range publishAppMap {
		switch appInfo.AppType {
		case constant.AppTypeRag:
			ragIds = append(ragIds, appId)
		case constant.AppTypeAgent:
			assistantIds = append(assistantIds, appId)
		case constant.AppTypeWorkflow:
			workflowIds = append(workflowIds, appId)
		case constant.AppTypeChatflow:
			chatflowIds = append(chatflowIds, appId)
		case constant.AppTypeSkill:
			skillIds = append(skillIds, appId)
		}

	}
	if len(ragIds) > 0 {
		resp, err := rag.GetPublishRagDescBatch(ctx.Request.Context(), &rag_service.GetPublishRagDescBatchReq{
			RagIdList: ragIds,
			Identity:  &rag_service.Identity{UserId: userId, OrgId: orgId},
		})
		if err != nil {
			log.Errorf("getAppVersionBatch rag batch query failed, userId=%s orgId=%s ragCount=%d err=%v", userId, orgId, len(ragIds), err)
		} else if resp == nil {
			log.Errorf("getAppVersionBatch rag batch query got nil response, userId=%s orgId=%s ragCount=%d", userId, orgId, len(ragIds))
		} else {
			for _, item := range resp.List {
				versionMap[item.RagId] = item.Version
			}
		}
	}

	if len(assistantIds) > 0 {
		resp, err := assistant.AssistantSnapshotLatestBatch(ctx.Request.Context(), &assistant_service.AssistantSnapshotLatestBatchReq{
			AssistantIdList: assistantIds,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			log.Errorf("getAppVersionBatch assistant batch query failed, userId=%s orgId=%s assistantCount=%d err=%v", userId, orgId, len(assistantIds), err)
		} else if resp == nil {
			log.Errorf("getAppVersionBatch assistant batch query got nil response, userId=%s orgId=%s assistantCount=%d", userId, orgId, len(assistantIds))
		} else {
			for _, snapshot := range resp.List {
				versionMap[snapshot.AssistantId] = snapshot.Version
			}
		}
	}

	if len(workflowIds) > 0 {
		resp, err := MultiGetWorkflowVersionList(ctx, workflowIds)
		if err != nil {
			log.Errorf("getAppVersionBatch workflow batch query failed, userId=%s orgId=%s workflowCount=%d err=%v", userId, orgId, len(workflowIds), err)
		} else if resp == nil {
			log.Errorf("getAppVersionBatch workflow batch query got nil response, userId=%s orgId=%s workflowCount=%d", userId, orgId, len(workflowIds))
		} else {
			for _, item := range resp {
				versionMap[item.WorkflowID] = item.Version
			}
		}
	}
	if len(chatflowIds) > 0 {
		resp, err := MultiGetWorkflowVersionList(ctx, chatflowIds)
		if err != nil {
			log.Errorf("getAppVersionBatch chatflow batch query failed, userId=%s orgId=%s chatflowCount=%d err=%v", userId, orgId, len(chatflowIds), err)
		} else if resp == nil {
			log.Errorf("getAppVersionBatch chatflow batch query got nil response, userId=%s orgId=%s chatflowCount=%d", userId, orgId, len(chatflowIds))
		} else {
			for _, item := range resp {
				versionMap[item.WorkflowID] = item.Version
			}
		}
	}
	if len(skillIds) > 0 {
		publishBySkill, err := getCustomSkillPublishMap(ctx, skillIds)
		if err != nil {
			log.Errorf("getAppVersionBatch skill batch query failed, userId=%s orgId=%s skillCount=%d err=%v", userId, orgId, len(skillIds), err)
		} else {
			for skillId, item := range publishBySkill {
				versionMap[skillId] = item.GetVersion()
			}
		}
	}

	return versionMap
}

// fillAppPublishInfo 批量填充应用发布信息（发布类型和版本）
func fillAppPublishInfo(ctx *gin.Context, userId, orgId string, apps []response.AppBriefInfo) (*response.ListResult, error) {
	if len(apps) == 0 {
		return &response.ListResult{List: apps, Total: 0}, nil
	}

	// 收集 appIds
	var appIds []string
	for _, appInfo := range apps {
		appIds = append(appIds, appInfo.AppId)
	}

	// 批量获取 app 信息
	appInfos, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: appIds,
		AppType:    "",
	})
	if err != nil {
		return nil, err
	}

	// 构建 map
	publishAppMap := make(map[string]*app_service.AppInfo, len(appInfos.Infos))
	for _, appInfo := range appInfos.Infos {
		publishAppMap[appInfo.AppId] = appInfo
	}

	// 批量获取版本信息
	versionMap := getAppVersionBatch(ctx, userId, orgId, publishAppMap)

	// 填充发布类型和版本
	for idx := range apps {
		if publishAppInfo, ok := publishAppMap[apps[idx].AppId]; ok {
			apps[idx].PublishType = publishAppInfo.PublishType
			if version, ok := versionMap[apps[idx].AppId]; ok {
				apps[idx].Version = version
			}
		}
	}

	// 按更新时间降序排序
	sort.SliceStable(apps, func(i, j int) bool {
		return apps[i].UpdatedAt > apps[j].UpdatedAt
	})

	return &response.ListResult{
		List:  apps,
		Total: int64(len(apps)),
	}, nil
}
