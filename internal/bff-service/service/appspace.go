package service

import (
	"sort"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/gin-gonic/gin"
)

const (
	successCode = 0
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
		// AgentScope Workflow
		// err = DeleteAgentScopeWorkFlow(ctx, userId, orgId, appId)

		// Coze Workflow
		err = DeleteWorkflow(ctx, orgId, appId)
	case constant.AppTypeChatflow:
		_, err = assistant.AssistantWorkFlowDeleteByWorkflowId(ctx.Request.Context(), &assistant_service.AssistantWorkFlowDeleteByWorkflowIdReq{
			WorkflowId: appId,
		})
		if err != nil {
			return err
		}
		// Coze Chatflow reuse workflow deletion interface
		err = DeleteWorkflow(ctx, orgId, appId)
	}
	return err
}

func GetAppSpaceAppList(ctx *gin.Context, userId, orgId, name, appType string) (*response.ListResult, error) {
	var ret []response.AppBriefInfo
	if appType == "" || appType == constant.AppTypeRag {
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
			ret = append(ret, appBriefProto2Model(ctx, ragInfo))
		}
	}
	if appType == "" || appType == constant.AppTypeAgent {
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
			ret = append(ret, appBriefProto2Model(ctx, assistantInfo))
		}
	}
	if appType == "" || appType == constant.AppTypeWorkflow {
		// AgentScope Workflow
		// resp, err := ListAgentScopeWorkFlow(ctx, userId, orgId, name)
		// if err != nil {
		// 	return nil, err
		// }
		// for _, workflowInfo := range resp.List {
		// 	ret = append(ret, agentscopeWorkflowInfo2Model(workflowInfo))
		// }

		// Coze Workflow
		resp, err := ListWorkflow(ctx, orgId, name, constant.AppTypeWorkflow)
		if err != nil {
			return nil, err
		}
		for _, workflowInfo := range resp.Workflows {
			ret = append(ret, cozeWorkflowInfo2Model(workflowInfo))
		}
	}
	if appType == "" || appType == constant.AppTypeChatflow {
		resp, err := ListWorkflow(ctx, orgId, name, constant.AppTypeChatflow)
		if err != nil {
			return nil, err
		}
		for _, chatflowInfo := range resp.Workflows {
			ret = append(ret, cozeChatflowInfo2Model(chatflowInfo))
		}
	}
	var appIds []string
	for _, appInfo := range ret {
		appIds = append(appIds, appInfo.AppId)
	}
	AppInfos, err := app.GetAppListByIds(ctx, &app_service.GetAppListByIdsReq{
		AppIdsList: appIds,
	})
	if err != nil {
		return nil, err
	}
	publishTypeMap := make(map[string]string, len(AppInfos.Infos))
	for _, appInfo := range AppInfos.Infos {
		publishTypeMap[appInfo.AppId] = appInfo.PublishType
	}
	for idx, appInfo := range ret {
		if publishType, ok := publishTypeMap[appInfo.AppId]; ok {
			ret[idx].PublishType = publishType
		}
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].UpdatedAt > ret[j].UpdatedAt
	})
	return &response.ListResult{
		List:  ret,
		Total: int64(len(ret)),
	}, nil
}

func PublishApp(ctx *gin.Context, userId, orgId string, req request.PublishAppRequest) error {
	// Special handling of release of AgentScope workflow
	// if req.AppType == constant.AppTypeWorkflow {
	// 	if err := PublishAgentScopeWorkFlow(ctx, userId, orgId, req.AppId); err != nil {
	// 		return err
	// 	}
	// }
	_, err := app.PublishApp(ctx.Request.Context(), &app_service.PublishAppReq{
		AppId:       req.AppId,
		AppType:     req.AppType,
		PublishType: req.PublishType,
		UserId:      userId,
		OrgId:       orgId,
	})
	return err
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
	// Special handling of unpublishing of AgentScope workflow
	// if req.AppType == constant.AppTypeWorkflow {
	// 	err = UnPublishAgentScopeWorkFlow(ctx, userId, orgId, req.AppId)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func GetAppList(ctx *gin.Context, userId, orgId, appType string) (*response.ListResult, error) {
	resp, err := app.GetAppList(ctx.Request.Context(), &app_service.GetAppListReq{
		OrgId:   orgId,
		AppType: appType,
		UserId:  userId,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  resp.Infos,
		Total: int64(len(resp.Infos)),
	}, nil
}
