package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	net_url "net/url"
	"sort"
	"time"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func ListLlmModelsByWorkflow(ctx *gin.Context, userId, orgId, modelT string) (*response.ListResult, error) {
	modelResp, err := ListTypeModels(ctx, userId, orgId, &request.ListTypeModelsRequest{ModelType: mp.ModelTypeLLM})
	if err != nil {
		return nil, err
	}
	var rets []*response.CozeWorkflowModelInfo
	for _, modelInfo := range modelResp.List.([]*response.ModelInfo) {
		ret, err := toModelInfo4Workflow(modelInfo)
		if err != nil {
			return nil, err
		}
		rets = append(rets, ret)
	}
	return &response.ListResult{
		List:  rets,
		Total: modelResp.Total,
	}, nil
}

// ListWorkflow userID/orgID数据隔离，用于【工作流】
func ListWorkflow(ctx *gin.Context, orgID, name, appType string) (*response.CozeWorkflowListData, error) {
	switch appType {
	case constant.AppTypeWorkflow:
		appType = "0"
	case constant.AppTypeChatflow:
		appType = "3"
	}
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ListUri)
	ret := &response.CozeWorkflowListResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"login_user_create": "true",
			"space_id":          orgID,
			"name":              name,
			"page":              "1",
			"size":              "99999",
			"flow_mode":         appType,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

// ListWorkflowByIDs 无userID或orgID隔离，用于【智能体选工作流】【应用广场】业务流程中
func ListWorkflowByIDs(ctx *gin.Context, name string, workflowIDs []string) (*response.CozeWorkflowListData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ListUri)
	ret := &response.CozeWorkflowListResp{}
	request := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"name": name,
			"page": "1",
			"size": "99999",
		})
	if len(workflowIDs) > 0 {
		request = request.SetBody(map[string]interface{}{
			"workflow_ids": workflowIDs,
		})
	}
	if resp, err := request.SetResult(ret).Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

// CreateWorkflowOrChatflow 统一创建工作流或对话流，包含名称唯一性校验
func CreateWorkflowOrChatflow(ctx *gin.Context, orgID, appType, name, desc, iconUri string) (*response.CozeWorkflowIDData, error) {
	if appType == constant.AppTypeChatflow {
		return CreateChatflow(ctx, orgID, name, desc, iconUri)
	}
	return CreateWorkflow(ctx, orgID, name, desc, iconUri)
}

func CreateWorkflow(ctx *gin.Context, orgID, name, desc, iconUri string) (*response.CozeWorkflowIDData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.CreateUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"space_id": orgID,
			"name":     name,
			"desc":     desc,
			"icon_uri": iconUri,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func CopyWorkflow(ctx *gin.Context, orgID, workflowID string, needPublished bool) (*response.CozeWorkflowIDData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.CopyUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"space_id":    orgID,
			"workflow_id": workflowID,
			"qType":       util.IfElse(needPublished, uint8(2), uint8(0)), //（workflow中FromDraft:0,FromLatestVersion:2）
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_copy", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_copy", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_copy", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func DeleteWorkflow(ctx *gin.Context, orgID, workflowID string) error {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.DeleteUri)
	ret := &response.CozeWorkflowDeleteResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"workflow_id": workflowID,
			"space_id":    orgID,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", err.Error())
	} else if resp.StatusCode() >= 300 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 || (ret.Data != nil && ret.Data.Status != 0) {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", fmt.Sprintf("code %v msg %v status %v", ret.Code, ret.Msg, ret.Data.GetStatus()))
	}
	return nil
}

func ExportWorkFlow(ctx *gin.Context, orgID, workflowID, version string, needPublished bool) ([]byte, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ExportUri)
	ret := &response.CozeWorkflowExportResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"workflow_id": workflowID,
			"version":     version,
			"space_id":    orgID,
			"qType":       util.IfElse(!needPublished, uint8(0), util.IfElse(needPublished && version == "", uint8(2), uint8(1))),
			//适配从草稿导出 从最新版本导出 导出指定版本（workflow中FromDraft:0,FromSpecificVersion:1,FromLatestVersion:2）
		}).
		SetResult(&ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_export", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_export", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	}
	exportData := response.CozeWorkflowExportData{
		WorkflowName: ret.Data.WorkflowName,
		WorkflowDesc: ret.Data.WorkflowDesc,
		Schema:       ret.Data.Schema,
	}
	// 将结构体序列化为 JSON 字节
	jsonData, err := json.Marshal(exportData)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_export", fmt.Sprintf("export workflow unmarshal err:%v", err.Error()))
	}
	return jsonData, nil
}

func ImportWorkflow(ctx *gin.Context, orgID, appType string) (*response.CozeWorkflowIDData, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("get file err: %v", err))
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("open file err: %v", err))
	}
	defer func() { _ = file.Close() }()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("read file err: %v", err))
	}
	var rawData workflowImportData
	if err := json.Unmarshal(fileBytes, &rawData); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("schema unmarshal failed: %v", err))
	}
	// 校验name和desc
	if rawData.Name == "" || rawData.Desc == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", "name or desc is empty")
	}
	switch appType {
	case constant.AppTypeChatflow:
		appType = "3"
	// 默认工作流模式
	default:
		appType = "0"
	}
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ImportUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]string{
			"space_id":  orgID,
			"name":      rawData.Name,
			"desc":      rawData.Desc,
			"schema":    rawData.Schema,
			"flow_mode": appType,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func WorkflowConvert(ctx *gin.Context, orgId, workflowId, flowMode string) error {
	switch flowMode {
	case constant.AppTypeChatflow:
		flowMode = "3"
	case constant.AppTypeWorkflow:
		flowMode = "0"
	default:
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_convert", "invalid flow mode")
	}
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ConvertUri)
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"workflow_id": workflowId,
			"flow_mode":   util.MustI64(flowMode),
			"space_id":    orgId,
		}).
		Post(url)
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_convert", err.Error())
	} else if resp.StatusCode() >= 300 {
		b, err := io.ReadAll(resp.RawResponse.Body)
		if err != nil {
			return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_convert", fmt.Sprintf("[%v] %v", resp.StatusCode(), err))
		}
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_convert", fmt.Sprintf("[%v] %v", resp.StatusCode(), string(b)))
	}
	var oldAppType, newAppType string
	if flowMode == "3" {
		oldAppType = constant.AppTypeWorkflow
		newAppType = constant.AppTypeChatflow
	} else {
		oldAppType = constant.AppTypeChatflow
		newAppType = constant.AppTypeWorkflow
	}
	_, err = app.ConvertAppType(ctx.Request.Context(), &app_service.ConvertAppTypeReq{AppId: workflowId, OldAppType: oldAppType, NewAppType: newAppType})
	return err
}

func PublishedWorkflowRun(ctx *gin.Context, userId, orgId string, req request.WorkflowRunReq) (result *response.CozeNodeResult, err error) {
	// Step 1: 触发异步执行（使用web的test_run接口），获取executeId
	startTime := time.Now()
	isSuccess := false
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		costs := time.Since(startTime).Milliseconds()
		go func() {
			defer util.PrintPanicStack()
			RecordAppStatistic(detachedCtx, userId, orgId, req.WorkflowID, constant.AppTypeWorkflow, isSuccess, false, 0, int64(costs), constant.AppStatisticSourceWeb)
		}()
	}()

	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.WorkflowRunLatestVersionUri)
	testRunRet := &response.CozeWorkflowTestRunResponse{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetResult(testRunRet).
		SetBody(map[string]any{
			"workflow_id": req.WorkflowID,
			"space_id":    orgId,
			"input":       req.Input,
		}).
		Post(url)

	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run", err.Error())
	}
	if resp.StatusCode() >= 300 {
		b, _ := io.ReadAll(resp.RawResponse.Body)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
			fmt.Sprintf("[%d] %s", resp.StatusCode(), string(b)))
	}

	if testRunRet.Data == nil || testRunRet.Data.ExecuteID == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run", "execute_id is empty")
	}
	executeID := testRunRet.Data.ExecuteID

	// Step 2: 轮询执行状态(将coze的getprocess方法弄为同步，返回结果给应用广场的工作流使用)
	gpUri, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.GetProcessUri)

	// 创建一个带 30 分钟超时的子 context
	pollCtx, cancel := context.WithTimeout(ctx.Request.Context(), 30*time.Minute)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			// 超时 or 上层取消
			if errors.Is(pollCtx.Err(), context.DeadlineExceeded) {
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
					"workflow execution timeout")
			}
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
				"request canceled")

		case <-ticker.C:
			// 执行一次状态查询
			statusResp, err := trace_util.NewResty(ctx).
				R().
				SetContext(pollCtx). // 使用带超时的 context
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/json").
				SetHeaders(workflowHttpReqHeader(ctx)).
				SetQueryParams(map[string]string{
					"execute_id":  executeID,
					"workflow_id": req.WorkflowID,
					"space_id":    orgId,
				}).
				Get(gpUri)

			if err != nil {
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run", err.Error())
			}
			if statusResp.StatusCode() >= 300 {
				b, _ := io.ReadAll(statusResp.RawResponse.Body)
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
					fmt.Sprintf("[%d] %s", statusResp.StatusCode(), string(b)))
			}

			var processResp response.CozeGetWorkflowProcessResponse
			if err := json.Unmarshal(statusResp.Body(), &processResp); err != nil {
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
					fmt.Sprintf("failed to unmarshal status response: %v", err))
			}
			data := processResp.Data
			if data.ExecuteId == "" {
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run", "executeId is empty in status response")
			}

			switch data.ExecuteStatus {
			case 2: // Success
				for _, nodeResult := range data.NodeResults {
					if nodeResult.NodeType == "End" && nodeResult.NodeId == "900001" {
						isSuccess = true
						return nodeResult, nil
					}
				}
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run", "workflow execution succeeded but End node result not found")

			case 3: // Failed
				reason := "unknown"
				if data.Reason != nil {
					reason = *data.Reason
				}
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
					fmt.Sprintf("workflow execution failed: %s", reason))

			case 4: // Canceled
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
					"workflow execution was canceled")

			case 1: // Running — 继续下一次轮询
				continue

			default:
				return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_exploration_run",
					fmt.Sprintf("unexpected executeStatus: %d", data.ExecuteStatus))
			}
		}
	}

}

func PublishWorkflow(ctx *gin.Context, orgID, workflowID, version, versionDesc string) error {
	body := map[string]any{
		"space_id":            orgID,
		"workflow_id":         workflowID,
		"has_collaborator":    false,
		"force":               true,
		"workflow_version":    version,
		"version_description": versionDesc,
	}
	url := config.Cfg().Workflow.Endpoint + config.Cfg().Workflow.PublishUri
	ret := &response.CozeCommonResp{}
	resp, err := trace_util.NewResty(ctx).R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(body).
		SetResult(ret).
		Post(url)
	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_publish", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_publish", fmt.Sprintf("[%d] http error", resp.StatusCode()))
	}
	if ret.Code != 0 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_publish", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return nil
}

func GetWorkflowVersionList(ctx *gin.Context, workflowID string) (*response.CozeWorkflowVersionListData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.VersionListUri)
	ret := &response.CozeWorkflowVersionListResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]string{
			"workflow_id": workflowID,
		}).
		SetResult(ret).
		Post(url)

	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("[%d] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	}
	if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func MultiGetWorkflowVersionList(ctx *gin.Context, workflowIDs []string) ([]*response.CozeWorkflowVersionInfo, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.MultiVersionListUri)
	ret := &response.CozeMGetWorkflowLatestVersionResponse{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"workflow_ids": workflowIDs,
		}).
		SetResult(ret).
		Post(url)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("[%d] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	}
	if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func GetWorkflowVersion(ctx *gin.Context, appID string) (string, string, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.VersionListUri)
	ret := &response.CozeWorkflowVersionListResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]string{
			"workflow_id": appID,
		}).
		SetResult(ret).
		Post(url)

	if err != nil {
		return "", "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return "", "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("[%d] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	}
	if ret.Code != 0 {
		return "", "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	sort.SliceStable(ret.Data.VersionList, func(i, j int) bool {
		return ret.Data.VersionList[i].CreatedAt > ret.Data.VersionList[j].CreatedAt
	})
	if len(ret.Data.VersionList) == 0 {
		return "", "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_version_list",
			fmt.Sprintf("workflow %s no version", appID))
	}
	return ret.Data.VersionList[0].Version, ret.Data.VersionList[0].Desc, nil
}

func UpdateWorkflowVersionDesc(ctx *gin.Context, workflowID, description string) error {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.UpdateVersionDescUri)
	ret := &response.CozeCommonResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]interface{}{
			"workflow_id":         workflowID,
			"version_description": description,
		}).
		SetResult(ret).
		Put(url)

	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_update_desc", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_update_desc",
			fmt.Sprintf("[%d] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	}
	if ret.Code != 0 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_update_desc",
			fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return nil
}

func RollbackWorkflowVersion(ctx *gin.Context, workflowID, version string) error {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.RollbackUri)
	ret := &response.CozeCommonResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]string{
			"workflow_id": workflowID,
			"version":     version,
		}).
		SetResult(ret).
		Post(url)

	if err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_rollback", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_rollback",
			fmt.Sprintf("[%d] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
	}
	if ret.Code != 0 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_rollback",
			fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return nil
}

func GetWorkflowSchemas(ctx *gin.Context, workflowIDs []string) ([]*openapi3.T, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ListSchemaUri)
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]any{
			"workflow_ids": workflowIDs,
		}).
		Post(url)

	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_list_schema", err.Error())
	}
	if resp.StatusCode() >= 300 {
		b, _ := io.ReadAll(resp.RawResponse.Body)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_list_schema",
			fmt.Sprintf("[%d] %s", resp.StatusCode(), string(b)))
	}
	var schemas []interface{}
	if err := json.Unmarshal(resp.Body(), &schemas); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_list_schema", fmt.Sprintf("failed to unmarshal schemas response: %v", err))
	}
	var ret []*openapi3.T
	for _, schema := range schemas {
		schemaBytes, err := json.Marshal(schema)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_list_schema", fmt.Sprintf("failed to marshal schema: %v", err))
		}
		doc, err := openapi3_util.LoadFromData(ctx.Request.Context(), schemaBytes)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_list_schema", fmt.Sprintf("failed to load schema: %v", err))
		}
		ret = append(ret, doc)
	}
	return ret, nil
}

// --- internal ---

type workflowImportData struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Schema string `json:"schema"` // 存储为JSON字符串
}

func workflowHttpReqHeader(ctx *gin.Context) map[string]string {
	// X-Org-Id 优先从 gin.Context 读取（中间件已设置），如果没有则从 header 读取
	orgID := ctx.GetString(gin_util.X_ORG_ID)
	if orgID == "" {
		orgID = ctx.GetHeader(gin_util.X_ORG_ID)
	}
	return map[string]string{
		"Authorization": ctx.GetHeader("Authorization"),
		"X-Org-Id":      orgID,
		"X-User-Id":     ctx.GetString(gin_util.USER_ID),
		"Content-Type":  "application/json",
	}
}

func cozeWorkflowInfo2Model(workflowInfo *response.CozeWorkflowListDataWorkflow) response.AppBriefInfo {
	return response.AppBriefInfo{
		AppId:     workflowInfo.WorkflowId,
		AppType:   constant.AppTypeWorkflow,
		Name:      workflowInfo.Name,
		Desc:      workflowInfo.Desc,
		Avatar:    cacheWorkflowAvatar(workflowInfo.URL, constant.AppTypeWorkflow),
		CreatedAt: util.Time2Str(workflowInfo.CreateTime * 1000),
		UpdatedAt: util.Time2Str(workflowInfo.UpdateTime * 1000),
	}
}

func toModelInfo4Workflow(modelInfo *response.ModelInfo) (*response.CozeWorkflowModelInfo, error) {
	// 默认过滤thinking_type参数
	needFilterThinkingType := true

	ret := &response.CozeWorkflowModelInfo{
		ModelInfo: *modelInfo,
		ModelAbility: response.CozeWorkflowModelInfoAbility{
			CotDisplay: true, // 固定给llm节点增加 reasoning_content 字段
		},
	}
	if modelInfo.Config != nil {
		cfg := make(map[string]interface{})
		b, err := json.Marshal(modelInfo.Config)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("model %v marshal config err: %v", modelInfo.ModelId, err))
		}
		if err = json.Unmarshal(b, &cfg); err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("model %v unmarshal config err: %v", modelInfo.ModelId, err))
		}
		for k, v := range cfg {
			switch k {
			case "functionCalling":
				if fc, ok := v.(string); ok && mp_common.FCType(fc) == mp_common.FCTypeToolCall {
					ret.ModelAbility.FunctionCall = true
				}
			case "visionSupport":
				if vs, ok := v.(string); ok && mp_common.VSType(vs) == mp_common.VSTypeSupport {
					ret.ModelAbility.ImageUnderstanding = true
				}
			case "thinkingSupport":
				if ts, ok := v.(string); ok && mp_common.ThinkingType(ts) == mp_common.ThinkingTypeSupport {
					needFilterThinkingType = false
				}
			}
		}
	}

	// 根据需要过滤modelParams
	if needFilterThinkingType {
		for _, param := range config.Cfg().Workflow.ModelParams {
			if param.Name != "thinking_type" {
				ret.ModelParams = append(ret.ModelParams, param)
			}
		}
	} else {
		ret.ModelParams = config.Cfg().Workflow.ModelParams
	}

	return ret, nil
}
