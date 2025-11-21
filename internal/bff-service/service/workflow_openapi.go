package service

import (
	"fmt"
	"io"
	net_url "net/url"
	"path/filepath"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func OpenAPIWorkflowRun(ctx *gin.Context, workflowID string, input []byte) ([]byte, error) {
	// Generate the url for calling the workflow
	// Transparently transmit the input entered by the user
	testRunUrl, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, fmt.Sprintf(config.Cfg().Workflow.TestRunUri, workflowID))
	resp, err := resty.New().
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(input).
		SetDoNotParseResponse(true).
		Post(testRunUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_test_run", err.Error())
	} else if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_test_run", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_test_run", err.Error())
	}
	return b, nil
}

func OpenAPIWorkflowFileUpload(ctx *gin.Context) (string, error) {
	// Get file from context
	fh, err := ctx.FormFile("file")
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("read file err: %v", err))
	}
	file, err := fh.Open()
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("open file err: %v", err))
	}
	defer file.Close()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("io read all file err: %v", err))
	}
	fileExtension := filepath.Ext(fh.Filename)
	// Generate the storeUri of the file on tos
	uploadActionUri, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.UploadActionUri)
	uploadActionRet := &cozeApplyUploadActionResponse{}
	if resp, err := resty.New().
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
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	}
	var storeUri string
	if len(uploadActionRet.Result.UploadAddress.StoreInfos) > 0 {
		storeUri = uploadActionRet.Result.UploadAddress.StoreInfos[0].StoreUri
	} else {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", "invalid response format: missing StoreUri")
	}
	// Use storeUri+fileBytes to upload files
	uploadCommonUrl, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.UploadCommonUri, storeUri)
	if resp, err := resty.New().
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/octet-stream").
		SetBody(fileBytes).
		Post(uploadCommonUrl); err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", err.Error())
	} else if resp.StatusCode() >= 300 {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
	}
	// Generate a signature and return the url of the accessible file
	signImgUrl, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.SignImgUri)
	ret := &cozeWorkflowSignImgUrlResp{}
	if resp, err := resty.New().
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
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_file_upload", fmt.Sprintf("[%v] %v", resp.StatusCode(), resp.String()))
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
