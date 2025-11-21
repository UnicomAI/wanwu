package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	net_url "net/url"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const (
	redisGlobalBrowseKey             = "globalBrowse"
	redisWorkflowTemplateDownloadKey = "workflowTemplateDownloadCount"
)

func GetWorkflowTemplateList(ctx *gin.Context, clientId, category, name string) (*response.GetWorkflowTemplateListResp, error) {
	// Record platform board views
	if category == "" || category == "all" {
		if err := recordGlobalBrowse(ctx.Request.Context()); err != nil {
			log.Errorf("record template browse count error: %v", err)
		}
	}
	// Record client data
	if _, err := operate.AddClientRecord(ctx.Request.Context(), &operate_service.AddClientRecordReq{
		ClientId: clientId,
	}); err != nil {
		log.Errorf("record client err:%v", err)
	}

	switch config.Cfg().WorkflowTemplate.ServerMode {
	case "remote":
		return getRemoteWorkflowTemplateList(ctx, category, name)
	case "local":
		return getLocalWorkflowTemplateList(ctx.Request.Context(), category, name)
	default:
		// Use local mode by default
		return getLocalWorkflowTemplateList(ctx.Request.Context(), category, name)
	}
}

func GetWorkflowTemplateDetail(ctx *gin.Context, clientId, templateId string) (*response.WorkflowTemplateDetail, error) {
	switch config.Cfg().WorkflowTemplate.ServerMode {
	case "remote":
		return getRemoteWorkflowTemplateDetail(ctx, templateId)
	case "local":
		return getLocalWorkflowTemplateDetail(ctx.Request.Context(), templateId)
	default:
		// Use local mode by default
		return getLocalWorkflowTemplateDetail(ctx.Request.Context(), templateId)
	}
}

func GetWorkflowTemplateRecommend(ctx *gin.Context, clientId, templateId string) (*response.GetWorkflowTemplateListResp, error) {
	switch config.Cfg().WorkflowTemplate.ServerMode {
	case "remote":
		res, err := getRemoteWorkflowTemplateList(ctx, "", "")
		if err != nil {
			return nil, err
		}
		return res, nil
	case "local":
		res, err := getLocalWorkflowTemplateList(ctx.Request.Context(), "", "")
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		// Use local mode by default
		res, err := getLocalWorkflowTemplateList(ctx.Request.Context(), "", "")
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func DownloadWorkflowTemplate(ctx *gin.Context, clientId, templateId string) ([]byte, error) {
	// Record workflow template download data
	if err := recordTemplateDownloadCount(ctx.Request.Context(), templateId); err != nil {
		log.Errorf("record template download count error: %v", err)
	}
	switch config.Cfg().WorkflowTemplate.ServerMode {
	case "remote":
		res, err := getRemoteDownloadWorkflowTemplate(ctx, templateId)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "local":
		res, err := getLocalDownloadWorkflowTemplate(templateId)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		// Use local mode by default
		res, err := getLocalDownloadWorkflowTemplate(templateId)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func CreateWorkflowByTemplate(ctx *gin.Context, orgId, clientId string, req request.CreateWorkflowByTemplateReq) (*response.CozeWorkflowIDData, error) {
	switch config.Cfg().WorkflowTemplate.ServerMode {
	case "remote":
		res, err := getRemoteCreateWorkflowByTemplate(ctx, orgId, req)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "local":
		res, err := getLocalCreateWorkflowByTemplate(ctx, orgId, req)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		// Use local mode by default
		res, err := getLocalCreateWorkflowByTemplate(ctx, orgId, req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

// --- Get the list of workflow templates ---

func getRemoteWorkflowTemplateList(ctx *gin.Context, category, name string) (*response.GetWorkflowTemplateListResp, error) {
	client := resty.NewWithClient(&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second, // Connection timeout
				KeepAlive: time.Minute,      // How long the connection remains active
			}).DialContext,
			ResponseHeaderTimeout: time.Minute,
		},
		Timeout: time.Minute,
	})
	var res response.Response
	var ret response.GetWorkflowTemplateListResp
	resp, err := client.R().
		SetContext(ctx.Request.Context()).
		SetQueryParams(map[string]string{
			"category": category,
			"name":     name,
		}).
		SetHeader("Accept", "application/json").
		SetResult(&res).
		Get(config.Cfg().WorkflowTemplate.ListUrl)
	if err != nil {
		// The remote call fails and the default download link is returned.
		log.Errorf("request remote workflow template err: %v", err)
		return &response.GetWorkflowTemplateListResp{
			Total: 0,
			List:  make([]*response.WorkflowTemplateInfo, 0),
			DownloadLink: response.WorkflowTemplateURL{
				Url: config.Cfg().WorkflowTemplate.GlobalWebListUrl,
			},
		}, nil
	}

	if resp.StatusCode() != http.StatusOK {
		// status not ok, return to default download link
		log.Errorf("request remote workflow template http code: %v, resp: %v", resp.StatusCode(), resp.String())
		return &response.GetWorkflowTemplateListResp{
			Total: 0,
			List:  make([]*response.WorkflowTemplateInfo, 0),
			DownloadLink: response.WorkflowTemplateURL{
				Url: config.Cfg().WorkflowTemplate.GlobalWebListUrl,
			},
		}, nil
	}
	marshal, err := json.Marshal(res.Data)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_template_list", fmt.Sprintf("request  marshal response body: %v", err))
	}
	if err = json.Unmarshal(marshal, &ret); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_template_list", fmt.Sprintf("request  unmarshal response body: %v", err))
	}
	return &ret, nil
}

func getLocalWorkflowTemplateList(ctx context.Context, category, name string) (*response.GetWorkflowTemplateListResp, error) {
	var resWorkflowTemp []*response.WorkflowTemplateInfo
	for _, wtfCfg := range config.Cfg().WorkflowTemplates {
		if name != "" && !strings.Contains(wtfCfg.Name, name) {
			continue
		}
		if !(category == "" || category == "all") && !strings.Contains(wtfCfg.Category, category) {
			continue
		}
		resWorkflowTemp = append(resWorkflowTemp, buildWorkflowTempInfo(ctx, *wtfCfg))
	}
	return &response.GetWorkflowTemplateListResp{
		Total:        int64(len(resWorkflowTemp)),
		List:         resWorkflowTemp,
		DownloadLink: response.WorkflowTemplateURL{},
	}, nil
}

// --- Get workflow template details ---

func getRemoteWorkflowTemplateDetail(ctx *gin.Context, templateId string) (*response.WorkflowTemplateDetail, error) {
	var res response.Response
	var ret response.WorkflowTemplateDetail
	resp, err := resty.New().R().
		SetContext(ctx.Request.Context()).
		SetQueryParams(map[string]string{
			"templateId": templateId,
		}).
		SetHeader("Accept", "application/json").
		SetResult(&res).
		Get(config.Cfg().WorkflowTemplate.DetailUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_detail", fmt.Sprintf("failed to call remote workflow template API: %v", err))
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_detail", fmt.Sprintf("request remote workflow template http code: %v", resp.StatusCode()))
	}
	marshal, err := json.Marshal(res.Data)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_template_detail", fmt.Sprintf("request marshal response body: %v", err))
	}
	if err = json.Unmarshal(marshal, &ret); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_template_detail", fmt.Sprintf("request unmarshal response body: %v", err))
	}
	// The remote call is successful and the remote result is returned.
	return &ret, nil
}

func getLocalWorkflowTemplateDetail(ctx context.Context, templateId string) (*response.WorkflowTemplateDetail, error) {
	wtfCfg, exist := config.Cfg().WorkflowTemp(templateId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_detail", "get local workflow template detail empty")
	}
	return buildWorkflowTempDetail(ctx, wtfCfg), nil
}

// --- Download workflow template ---

func getRemoteDownloadWorkflowTemplate(ctx *gin.Context, templateId string) ([]byte, error) {
	resp, err := resty.New().R().
		SetContext(ctx.Request.Context()).
		SetQueryParams(map[string]string{
			"templateId": templateId,
		}).
		SetHeader("Accept", "application/json").
		Get(config.Cfg().WorkflowTemplate.DownloadUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_download", fmt.Sprintf("failed to call remote workflow template API: %v", err))
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_download", fmt.Sprintf("request remote workflow template http code: %v", resp.StatusCode()))
	}
	// The remote call is successful and the remote result is returned.
	return convertToBytes(resp.Body())
}

func getLocalDownloadWorkflowTemplate(templateId string) ([]byte, error) {
	wtfCfg, exist := config.Cfg().WorkflowTemp(templateId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_download", "get local workflow template download empty")
	}
	return []byte(wtfCfg.Schema), nil
}

// --- Copy workflow template ---

func getRemoteCreateWorkflowByTemplate(ctx *gin.Context, orgId string, req request.CreateWorkflowByTemplateReq) (*response.CozeWorkflowIDData, error) {
	resp, err := getRemoteWorkflowTemplateList(ctx, "", "")
	if err != nil {
		return nil, err
	}
	var schema []byte
	for _, i := range resp.List {
		if i.TemplateId == req.TemplateId {
			schemaJson, err := getRemoteDownloadWorkflowTemplate(ctx, i.TemplateId)
			if err != nil {
				return nil, err
			}
			schema = schemaJson
			break
		}
	}
	return createWorkflowByTemplate(ctx, orgId, req, schema)
}

func getLocalCreateWorkflowByTemplate(ctx *gin.Context, orgId string, req request.CreateWorkflowByTemplateReq) (*response.CozeWorkflowIDData, error) {
	wtfCfg, exist := config.Cfg().WorkflowTemp(req.TemplateId)
	if !exist {
		return nil, fmt.Errorf("template not found: %s", req.TemplateId)
	}
	return createWorkflowByTemplate(ctx, orgId, req, wtfCfg.Schema)
}

// Workflow file parsing structure
type workflowTemplateSchema struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Schema string `json:"schema"`
}

// Extract public functions created by the workflow
func createWorkflowByTemplate(ctx *gin.Context, orgId string, req request.CreateWorkflowByTemplateReq, schema []byte) (*response.CozeWorkflowIDData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ImportUri)
	ret := &response.CozeWorkflowIDResp{}
	// Analyze outer structure
	var templateSchema workflowTemplateSchema
	if err := json.Unmarshal(schema, &templateSchema); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", err.Error())
	}
	if resp, err := resty.New().
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(map[string]string{
			"space_id": orgId,
			"name":     req.Name,
			"desc":     req.Desc,
			"schema":   templateSchema.Schema,
			"icon_url": req.Avatar.Key,
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

func buildWorkflowTempInfo(ctx context.Context, wtfCfg config.WorkflowTemplateConfig) *response.WorkflowTemplateInfo {
	iconUrl, _ := net_url.JoinPath(config.Cfg().Server.ApiBaseUrl, config.Cfg().DefaultIcon.WorkflowIcon)
	return &response.WorkflowTemplateInfo{
		TemplateId: wtfCfg.TemplateId,
		Avatar: request.Avatar{
			Path: iconUrl,
		},
		Name:          wtfCfg.Name,
		Author:        wtfCfg.Author,
		Desc:          wtfCfg.Desc,
		Category:      wtfCfg.Category,
		DownloadCount: getTemplateDownloadCount(ctx, wtfCfg.TemplateId),
	}
}

func buildWorkflowTempDetail(ctx context.Context, wtfCfg config.WorkflowTemplateConfig) *response.WorkflowTemplateDetail {
	iconUrl, _ := net_url.JoinPath(config.Cfg().Server.ApiBaseUrl, config.Cfg().DefaultIcon.WorkflowIcon)
	return &response.WorkflowTemplateDetail{
		WorkflowTemplateInfo: response.WorkflowTemplateInfo{
			TemplateId: wtfCfg.TemplateId,
			Avatar: request.Avatar{
				Path: iconUrl,
			},
			Name:          wtfCfg.Name,
			Desc:          wtfCfg.Desc,
			Category:      wtfCfg.Category,
			Author:        wtfCfg.Author,
			DownloadCount: getTemplateDownloadCount(ctx, wtfCfg.TemplateId),
		},
		Summary:  wtfCfg.Summary,
		Feature:  wtfCfg.Feature,
		Scenario: wtfCfg.Scenario,
		Note:     wtfCfg.Note,
	}
}

func convertToBytes(data any) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	switch v := data.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_download", "unsupported data type for conversion to bytes")
	}
}

// Record template downloads to a separate Redis Key
func recordTemplateDownloadCount(ctx context.Context, templateID string) error {
	// Use HINCRBY atomicity to increase template downloads
	err := redis.OP().Cli().HIncrBy(ctx, redisWorkflowTemplateDownloadKey, templateID, 1).Err()
	if err != nil {
		return fmt.Errorf("redis HIncrBy key %v field %v err: %v", redisWorkflowTemplateDownloadKey, templateID, err)
	}
	return nil
}

// Get downloads based on templateId
func getTemplateDownloadCount(ctx context.Context, templateID string) int32 {
	// Use HGet to get the download count of a specified template
	countStr, err := redis.OP().Cli().HGet(ctx, redisWorkflowTemplateDownloadKey, templateID).Result()
	if err != nil {
		// Key or field does not exist, return 0
		return 0
	}
	return util.MustI32(countStr)
}
