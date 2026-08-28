package service

import (
	"encoding/json"
	"fmt"
	"io"
	net_url "net/url"
	"path/filepath"
	"time"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func OpenAPIWorkflowRun(ctx *gin.Context, userId, orgId string, req request.OpenAPIWorkflowRunReq) (result []byte, err error) {
	// 生成调用工作流的url
	// 将用户输入的 parameters 透传
	startTime := time.Now()
	isSuccess := false
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		costs := time.Since(startTime).Milliseconds()
		statusCode, failureReason := GrpcErrorToHTTPStatus(err)
		respBody := ""
		if isSuccess && len(result) > 0 {
			respBody = string(result)
		}
		go func() {
			defer util.PrintPanicStack()
			question := MarshalStatisticBody(req.Parameters)
			RecordAppStatistic(detachedCtx, userId, orgId, req.UUID, constant.AppTypeWorkflow, "",
				statusCode, failureReason, false, 0, int64(costs), constant.BizSourceOpenAPI, MarshalStatisticBody(req), respBody, question, respBody)
		}()
	}()

	input, err := json.Marshal(req.Parameters)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_openapi_workflow_run", err.Error())
	}

	testRunUrl, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, fmt.Sprintf(config.Cfg().Workflow.WorkflowRunByOpenapiUri, req.UUID))
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(input).
		SetDoNotParseResponse(true).
		Post(testRunUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_test_run", err.Error())
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_test_run", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_test_run", fmt.Sprintf("[%v] %v", resp.StatusCode(), string(b)))
	}
	isSuccess = true
	return b, nil
}

func OpenAPIWorkflowFileUpload(ctx *gin.Context) (string, error) {
	// 从context中获取file
	fh, err := ctx.FormFile("file")
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("read file err: %v", err))
	}
	file, err := fh.Open()
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("open file err: %v", err))
	}
	defer func() { _ = file.Close() }()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("io read all file err: %v", err))
	}
	fileExtension := filepath.Ext(fh.Filename)
	// 生成文件在tos上的storeUri
	uploadActionUri, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.UploadActionUri)
	uploadActionRet := &cozeApplyUploadActionResponse{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{
			"Action":        "ApplyImageUpload",
			"FileExtension": fileExtension,
		}).
		SetResult(uploadActionRet).
		Get(uploadActionUri); err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", err.Error())
	} else if resp.StatusCode() >= 300 {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v]", resp.StatusCode()))
	}
	var storeUri string
	if len(uploadActionRet.Result.UploadAddress.StoreInfos) > 0 {
		storeUri = uploadActionRet.Result.UploadAddress.StoreInfos[0].StoreUri
	} else {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", "invalid response format: missing StoreUri")
	}
	// 使用storeUri+fileBytes上传文件
	uploadCommonUrl, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.UploadCommonUri, storeUri)
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/octet-stream").
		SetBody(fileBytes).
		Post(uploadCommonUrl); err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", err.Error())
	} else if resp.StatusCode() >= 300 {
		b, err := io.ReadAll(resp.RawResponse.Body)
		if err != nil {
			return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), err))
		}
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), string(b)))
	}
	// 生成签名，并返回可访问文件的url
	signImgUrl, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.SignImgUri)
	ret := &cozeWorkflowSignImgUrlResp{}
	if resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(map[string]any{
			"uri": storeUri,
		}).
		SetResult(ret).
		Post(signImgUrl); err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", err.Error())
	} else if resp.StatusCode() >= 300 {
		b, err := io.ReadAll(resp.RawResponse.Body)
		if err != nil {
			return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), err))
		}
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), string(b)))
	}
	return ret.Url, nil
}

// --- internal ---
type cozeWorkflowSignImgUrlResp struct {
	Url string `json:"url"`
}

type cozeApplyUploadActionResponse struct {
	ResponseMetadata cozeResponseMetadata        `json:"ResponseMetadata"`
	Result           cozeApplyUploadActionResult `json:"Result"`
}

type cozeResponseMetadata struct {
	RequestId string `json:"RequestId"`
	Action    string `json:"Action"`
	Version   string `json:"Version"`
	Service   string `json:"Service"`
	Region    string `json:"Region"`
}

type cozeApplyUploadActionResult struct {
	UploadAddress cozeUploadAddress `json:"UploadAddress"`
}

type cozeUploadAddress struct {
	StoreInfos []cozeStoreInfo `json:"StoreInfos"`
}

type cozeStoreInfo struct {
	StoreUri string `json:"StoreUri"`
	Auth     string `json:"Auth"`
	UploadID string `json:"UploadID"`
}
