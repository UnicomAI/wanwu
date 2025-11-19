package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/http"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type RagGetReportParams struct {
	UserId            string `json:"userId"`
	KnowledgeBaseName string `json:"knowledgeBase"`
	KnowledgeId       string `json:"kb_id"`
	PageSize          int32  `json:"page_size"`
	SearchAfter       int32  `json:"search_after"`
}

type RagDeleteReportParams struct {
	UserId            string   `json:"userId"`
	KnowledgeBaseName string   `json:"knowledgeBase"`
	KnowledgeId       string   `json:"kb_id"`
	ReportIds         []string `json:"report_ids"`
}

type RagUpdateReportParams struct {
	UserId            string               `json:"userId"`
	KnowledgeBaseName string               `json:"knowledgeBase"`
	KnowledgeId       string               `json:"kb_id"`
	ReportItem        *RagUpdateReportItem `json:"reports"`
}

type RagUpdateReportItem struct {
	ReportId string `json:"report_id"`
	Content  string `json:"content"`
	Title    string `json:"title"`
}

type RagAddReportParams struct {
	UserId            string              `json:"userId"`
	KnowledgeBaseName string              `json:"knowledgeBase"`
	KnowledgeId       string              `json:"kb_id"`
	ReportItem        []*RagAddReportItem `json:"reports"`
}

type RagAddReportItem struct {
	Content string `json:"content"`
	Title   string `json:"title"`
}

type RagGetReportResp struct {
	RagCommonResp
	Data *RagReportListResp `json:"data"`
}

type RagBatchReportResp struct {
	RagCommonResp
	Data *RagBatchReportCount `json:"data"`
}

type RagBatchReportCount struct {
	SuccessCount int `json:"success_count"` // 成功的社区报告数量 [EN] Number of successful community reports
}

type RagReportListResp struct {
	List          []RagReportInfo `json:"content_list"`
	ChunkTotalNum int             `json:"chunk_total_num"` // 社区报告数量 [EN] Number of community reports
}

type RagReportInfo struct {
	ReportTitle string      `json:"report_title"` // 社区报告标题 [EN] Community report title
	Content     string      `json:"content"`      // 社区报告内容 [EN] Community report content
	ChunkId     string      `json:"chunk_id"`     // 社区报告id [EN] community report id
	MetaData    interface{} `json:"meta_data"`    // 元数据 [EN] Metadata
	Status      bool        `json:"status"`       // 社区报告状态 [EN] Community report status
	ContentId   string      `json:"content_id"`   // 社区报告id [EN] community report id
	KbName      string      `json:"kb_name"`      // 知识库名称 [EN] Knowledge base name
	CreateTime  string      `json:"create_time"`  // 生成时间 [EN] Generation time
}

// RagGetReport rag获取社区报告 [EN] RagGetReport rag gets community report
func RagGetReport(ctx context.Context, ragGetReportParams *RagGetReportParams) (*RagReportListResp, error) {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.GetCommunityReportListUri
	paramsByte, err := json.Marshal(ragGetReportParams)
	if err != nil {
		return nil, err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_get_report_list",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagGetReportResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}

// RagDeleteReport rag删除社区报告 [EN] RagDeleteReport rag delete community report
func RagDeleteReport(ctx context.Context, ragDeleteReportParams *RagDeleteReportParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.BatchDeleteReportsUri
	paramsByte, err := json.Marshal(ragDeleteReportParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_delete_report",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagBatchReportResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagUpdateReport rag更新社区报告 [EN] RagUpdateReport rag update community report
func RagUpdateReport(ctx context.Context, ragUpdateReportParams *RagUpdateReportParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.UpdateReportUri
	paramsByte, err := json.Marshal(ragUpdateReportParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_update_report",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagAddReport rag更新社区报告 [EN] RagAddReport rag update community report
func RagAddReport(ctx context.Context, ragAddReportParams *RagAddReportParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.BatchAddReportsUri
	paramsByte, err := json.Marshal(ragAddReportParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_add_report",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagBatchReportResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}
