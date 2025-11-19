package service

import (
	"encoding/json"
	"fmt"
	"io"
	net_url "net/url"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
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

// ListWorkflow userID/orgID数据隔离，用于【工作流】 [EN] ListWorkflow userID/orgID data isolation, used for [workflow]
func ListWorkflow(ctx *gin.Context, orgID, name, appType string) (*response.CozeWorkflowListData, error) {
	switch appType {
	case constant.AppTypeWorkflow:
		appType = "0"
	case constant.AppTypeChatflow:
		appType = "3"
	}
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ListUri)
	ret := &response.CozeWorkflowListResp{}
	if resp, err := resty.New().
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
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

// ListWorkflowByIDs 无userID或orgID隔离，用于【智能体选工作流】【应用广场】业务流程中 [EN] ListWorkflowByIDs has no userID or orgID isolation and is used in the [Intelligent Body Selection Workflow] [Application Square] business process
func ListWorkflowByIDs(ctx *gin.Context, name string, workflowIDs []string) (*response.CozeWorkflowListData, error) {
	var ids []string
	for _, workflowID := range workflowIDs {
		if _, err := util.I64(workflowID); err == nil {
			// AgentScope Workflow ID为uuid，将这部分脏数据过滤掉；Coze Workflow ID可转为数字 [EN] AgentScope Workflow ID is uuid, filter out this part of dirty data; Coze Workflow ID can be converted to numbers
			ids = append(ids, workflowID)
		}
	}
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ListUri)
	ret := &response.CozeWorkflowListResp{}
	request := resty.New().
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
	if len(ids) > 0 {
		request = request.SetBody(map[string]interface{}{
			"workflow_ids": ids,
		})
	}
	if resp, err := request.SetResult(ret).Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_apps_list", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func CreateWorkflow(ctx *gin.Context, orgID, name, desc, iconUri string) (*response.CozeWorkflowIDData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.CreateUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := resty.New().
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
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_create", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func CopyWorkflow(ctx *gin.Context, orgID, workflowID string) (*response.CozeWorkflowIDData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.CopyUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := resty.New().
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"space_id":    orgID,
			"workflow_id": workflowID,
		}).
		SetResult(ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_copy", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_copy", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_copy", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

func DeleteWorkflow(ctx *gin.Context, orgID, workflowID string) error {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.DeleteUri)
	ret := &response.CozeWorkflowDeleteResp{}
	if resp, err := resty.New().
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
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	} else if ret.Code != 0 || (ret.Data != nil && ret.Data.Status != 0) {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_app_delete", fmt.Sprintf("code %v msg %v status %v", ret.Code, ret.Msg, ret.Data.GetStatus()))
	}
	return nil
}

func ExportWorkflow(ctx *gin.Context, orgID, workflowID string) ([]byte, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ExportUri)
	ret := &response.CozeWorkflowExportResp{}
	if resp, err := resty.New().
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]string{
			"space_id":    orgID,
			"workflow_id": workflowID,
		}).
		SetResult(&ret).
		Post(url); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_export", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_export", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	}
	exportData := response.CozeWorkflowExportData{
		WorkflowName: ret.Data.WorkflowName,
		WorkflowDesc: ret.Data.WorkflowDesc,
		Schema:       ret.Data.Schema,
	}
	// 将结构体序列化为 JSON 字节 [EN] Serialize struct to JSON bytes
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
	defer file.Close()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("read file err: %v", err))
	}
	var rawData workflowImportData
	if err := json.Unmarshal(fileBytes, &rawData); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("schema unmarshal failed: %v", err))
	}
	// 校验name和desc [EN] Verify name and desc
	if rawData.Name == "" || rawData.Desc == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", "name or desc is empty")
	}
	switch appType {
	case constant.AppTypeChatflow:
		appType = "3"
	// 默认工作流模式 [EN] Default workflow mode
	default:
		appType = "0"
	}
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ImportUri)
	ret := &response.CozeWorkflowIDResp{}
	if resp, err := resty.New().
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
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
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	} else if ret.Code != 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("code %v msg %v", ret.Code, ret.Msg))
	}
	return ret.Data, nil
}

// --- internal ---

type workflowImportData struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Schema string `json:"schema"` // 存储为JSON字符串 [EN] Store as JSON string
}

func workflowHttpReqHeader(ctx *gin.Context) map[string]string {
	return map[string]string{
		"Authorization": ctx.GetHeader("Authorization"),
		"X-Org-Id":      ctx.GetHeader(gin_util.X_ORG_ID),
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
	ret := &response.CozeWorkflowModelInfo{
		ModelInfo:   *modelInfo,
		ModelParams: config.Cfg().Workflow.ModelParams,
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

			}
		}
	}
	return ret, nil
}
