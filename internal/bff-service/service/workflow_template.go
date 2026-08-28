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

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

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
	// 记录平台板浏览量
	if category == "" || category == "all" {
		if err := recordGlobalBrowse(ctx.Request.Context()); err != nil {
			log.Errorf("record template browse count error: %v", err)
		}
	}
	// 记录client数据
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
		// 默认使用本地模式
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
		// 默认使用本地模式
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
		// 默认使用本地模式
		res, err := getLocalWorkflowTemplateList(ctx.Request.Context(), "", "")
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func DownloadWorkflowTemplate(ctx *gin.Context, clientId, templateId string) ([]byte, error) {
	// 记录工作流模板下载数据
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
		// 默认使用本地模式
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
		// 默认使用本地模式
		res, err := getLocalCreateWorkflowByTemplate(ctx, orgId, req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

// --- 获取工作流模板列表 ---

func getRemoteWorkflowTemplateList(ctx *gin.Context, category, name string) (*response.GetWorkflowTemplateListResp, error) {
	client := resty.NewWithClient(&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second, // 连接超时时间
				KeepAlive: time.Minute,      // 连接保持活跃的时间
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
		// 远程调用失败，返回默认下载链接
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
		// status not ok,  返回默认下载链接
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
		if category != "" && category != "all" && !strings.Contains(wtfCfg.Category, category) {
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

// --- 获取工作流模板详情 ---

func getRemoteWorkflowTemplateDetail(ctx *gin.Context, templateId string) (*response.WorkflowTemplateDetail, error) {
	var res response.Response
	var ret response.WorkflowTemplateDetail
	resp, err := trace_util.NewResty(ctx).R().
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
	// 远程调用成功，返回远程结果
	return &ret, nil
}

func getLocalWorkflowTemplateDetail(ctx context.Context, templateId string) (*response.WorkflowTemplateDetail, error) {
	wtfCfg, exist := config.Cfg().WorkflowTemp(templateId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_detail", "get local workflow template detail empty")
	}
	return buildWorkflowTempDetail(ctx, wtfCfg), nil
}

// --- 下载工作流模板 ---

func getRemoteDownloadWorkflowTemplate(ctx *gin.Context, templateId string) ([]byte, error) {
	resp, err := trace_util.NewResty(ctx).R().
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
	// 远程调用成功，返回远程结果
	return resp.Body(), nil
}

func getLocalDownloadWorkflowTemplate(templateId string) ([]byte, error) {
	wtfCfg, exist := config.Cfg().WorkflowTemp(templateId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_workflow_template_download", "get local workflow template download empty")
	}
	return []byte(wtfCfg.Schema), nil
}

// --- 复制工作流模板 ---

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

// 工作流文件解析结构体
type workflowTemplateSchema struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Schema string `json:"schema"`
}

// 提取工作流创建的公共函数
func createWorkflowByTemplate(ctx *gin.Context, orgId string, req request.CreateWorkflowByTemplateReq, schema []byte) (*response.CozeWorkflowIDData, error) {
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ImportUri)
	ret := &response.CozeWorkflowIDResp{}
	// 解析外层结构
	var templateSchema workflowTemplateSchema
	if err := json.Unmarshal(schema, &templateSchema); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", err.Error())
	}
	if resp, err := trace_util.NewResty(ctx).
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
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_import_file", fmt.Sprintf("[%v] code %v msg %v", resp.StatusCode(), ret.Code, ret.Msg))
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

// 记录模板下载量到单独的Redis Key
func recordTemplateDownloadCount(ctx context.Context, templateID string) error {
	// 使用HINCRBY原子性增加模板下载量
	err := redis.OP().Cli().HIncrBy(ctx, redisWorkflowTemplateDownloadKey, templateID, 1).Err()
	if err != nil {
		return fmt.Errorf("redis HIncrBy key %v field %v err: %v", redisWorkflowTemplateDownloadKey, templateID, err)
	}
	return nil
}

// 根据templateId获取下载量
func getTemplateDownloadCount(ctx context.Context, templateID string) int32 {
	// 使用HGet获取指定模板的下载量
	countStr, err := redis.OP().Cli().HGet(ctx, redisWorkflowTemplateDownloadKey, templateID).Result()
	if err != nil {
		// 键或字段不存在，返回0
		return 0
	}
	return util.MustI32(countStr)
}
