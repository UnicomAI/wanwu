package service

import (
	"encoding/json"
	"errors"
	net_url "net/url"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/gin-gonic/gin"
)

func ListAgentScopeWorkFlow(ctx *gin.Context, userId, orgId, name string) (*response.AgentScopeWorkFlowPageResult, error) {
	workflowService := config.Cfg().AgentScopeWorkFlow
	url, _ := net_url.JoinPath(workflowService.Endpoint, workflowService.WorkFlowListUri)
	result, err := http_client.Default().Get(ctx.Request.Context(), &http_client.HttpRequestParams{
		Url: url,
		Headers: map[string]string{
			"x-org-id":      orgId,
			"x-user-id":     userId,
			"Authorization": ctx.GetHeader("Authorization"),
		},
		Params: map[string]string{
			"keyword": net_url.QueryEscape(name),
		},
		Timeout:    60 * time.Second,
		MonitorKey: "workflow_list",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", err.Error())
	}
	var resp = &response.AgentScopeWorkFlowListResp{}
	if err = json.Unmarshal(result, resp); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", err.Error())
	}
	if resp.Code != successCode {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", errors.New(resp.Message).Error())
	}
	return resp.Data, nil
}

func DeleteAgentScopeWorkFlow(ctx *gin.Context, userId, orgId, id string) error {
	workflowService := config.Cfg().AgentScopeWorkFlow
	url, _ := net_url.JoinPath(workflowService.Endpoint, workflowService.DeleteWorkFlowUri)
	params := &request.DeleteAgentScopeWorkFlowRequest{
		AppId: id,
	}
	body, err := json.Marshal(params)
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", err.Error())
	}
	result, err := http_client.Default().Delete(ctx.Request.Context(), &http_client.HttpRequestParams{
		Url: url,
		Headers: map[string]string{
			"x-org-id":      orgId,
			"x-user-id":     userId,
			"Authorization": ctx.GetHeader("Authorization"),
			"Content-Type":  "application/json",
		},
		Body:       body,
		Timeout:    60 * time.Second,
		MonitorKey: "workflow_delete",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", err.Error())
	}
	var resp = &response.AgentScopeDeleteWorkFlowResp{}
	if err = json.Unmarshal(result, resp); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", err.Error())
	}
	if resp.Code != successCode {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", errors.New(resp.Message).Error())
	}
	return nil
}

func PublishAgentScopeWorkFlow(ctx *gin.Context, userId, orgId, workflowID string) error {
	workflowService := config.Cfg().AgentScopeWorkFlow
	url, _ := net_url.JoinPath(workflowService.Endpoint, workflowService.PublishWorkFlowUri)
	params := &request.PublishAgentScopeWorkFlowRequest{
		AppId: workflowID,
	}
	body, err := json.Marshal(params)
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_publish", err.Error())
	}
	result, err := http_client.Default().PostJson(ctx.Request.Context(), &http_client.HttpRequestParams{
		Url: url,
		Headers: map[string]string{
			"x-org-id":      orgId,
			"x-user-id":     userId,
			"Authorization": ctx.GetHeader("Authorization"),
			"Content-Type":  "application/json",
		},
		Body:       body,
		Timeout:    60 * time.Second,
		MonitorKey: "workflow_publish",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_publish", err.Error())
	}
	type publishWorkFlowResp struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	var resp = &publishWorkFlowResp{}
	if err = json.Unmarshal(result, resp); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_publish", err.Error())
	}
	if resp.Code != successCode {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_publish", errors.New(resp.Message).Error())
	}
	return nil
}

func ListAgentScopeWorkFlowInternal(ctx *gin.Context) (*response.AgentScopeWorkFlowPageResult, error) {
	workflowService := config.Cfg().AgentScopeWorkFlow
	url, _ := net_url.JoinPath(workflowService.Endpoint, workflowService.WorkFlowListUriInternal)
	result, err := http_client.Default().Get(ctx.Request.Context(), &http_client.HttpRequestParams{
		Url: url,
		Headers: map[string]string{
			"Authorization": ctx.GetHeader("Authorization"),
		},
		Timeout:    60 * time.Second,
		MonitorKey: "workflow_list_internal",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list_internal", err.Error())
	}
	var resp = &response.AgentScopeWorkFlowListResp{}
	if err = json.Unmarshal(result, resp); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list_internal", err.Error())
	}
	if resp.Code != successCode {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list_internal", errors.New(resp.Message).Error())
	}
	return resp.Data, nil
}

func UnPublishAgentScopeWorkFlow(ctx *gin.Context, userId, orgId, workflowID string) error {
	workflowService := config.Cfg().AgentScopeWorkFlow
	url, _ := net_url.JoinPath(workflowService.Endpoint, workflowService.UnPublishWorkFlowUri)
	params := &request.UnPublishAgentScopeWorkFlowRequest{
		AppId: workflowID,
	}
	body, err := json.Marshal(params)
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_unpublish", err.Error())
	}
	result, err := http_client.Default().PostJson(ctx.Request.Context(), &http_client.HttpRequestParams{
		Url: url,
		Headers: map[string]string{
			"x-org-id":      orgId,
			"x-user-id":     userId,
			"Authorization": ctx.GetHeader("Authorization"),
			"Content-Type":  "application/json",
		},
		Body:       body,
		Timeout:    60 * time.Second,
		MonitorKey: "workflow_unpublish",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_unpublish", err.Error())
	}
	// 声明局部结构体类型 [EN] Declare local structure type
	type unpublishWorkFlowResp struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	var resp = &unpublishWorkFlowResp{}
	if err = json.Unmarshal(result, resp); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_unpublish", err.Error())
	}
	if resp.Code != successCode {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_unpublish", errors.New(resp.Message).Error())
	}
	return nil
}

// --- internal ---

// func agentscopeWorkflowInfo2Model(workflowInfo response.AgentScopeWorkFlowInfo) response.AppBriefInfo {
// 	return response.AppBriefInfo{
// 		AppId:   workflowInfo.Id,
// 		AppType: constant.AppTypeWorkflow,
// 		//Avatar:    CacheAvatar(ctx, workflowInfo.AvatarPath),
// 		Name:      workflowInfo.ConfigName,
// 		Desc:      workflowInfo.ConfigDesc,
// 		CreatedAt: workflowInfo.UpdatedTime,
// 		UpdatedAt: workflowInfo.UpdatedTime,
// 	}
// }
