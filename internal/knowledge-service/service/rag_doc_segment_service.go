package service

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/http"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type RagCreateDocSegmentParams struct {
	UserId           string            `json:"userId"`             // 发起请求的用户ID
	KnowledgeBase    string            `json:"knowledgeBase"`      // 知识库的名称
	KnowledgeId      string            `json:"kb_id"`              // 知识库的唯一ID
	FileName         string            `json:"fileName"`           // 与chunk关联的文件名
	MaxSentenceSize  int               `json:"max_sentence_size"`  // 最大分段长度限制
	Chunks           []*NewChunkItem   `json:"chunks"`             // 分段数据列表
	SplitType        string            `json:"split_type"`         // 分段类型 //parent_child|common
	ChildChunkConfig *ChildChunkConfig `json:"child_chunk_config"` //子分段配置
}

type RagCreateDocChildSegmentParams struct {
	UserId        string   `json:"userId"`         // 发起请求的用户ID
	KnowledgeBase string   `json:"knowledgeBase"`  // 知识库的名称
	KnowledgeId   string   `json:"kb_id"`          // 知识库的唯一ID
	FileName      string   `json:"fileName"`       // 与chunk关联的文件名
	ChunkId       string   `json:"chunk_id"`       // 父分段id
	ChildContents []string `json:"child_contents"` // 子分段内容
}

type RagDeleteDocChildSegmentParams struct {
	UserId                string  `json:"userId"`                   // 发起请求的用户ID
	KnowledgeBase         string  `json:"knowledgeBase"`            // 知识库的名称
	KnowledgeId           string  `json:"kb_id"`                    // 知识库的唯一ID
	FileName              string  `json:"fileName"`                 // 与chunk关联的文件名
	ChunkId               string  `json:"chunk_id"`                 // 父分段id
	ChunkCurrentNum       int32   `json:"chunk_current_num"`        // 父分段序列号
	ChildChunkCurrentNums []int32 `json:"child_chunk_current_nums"` // 子分段序列号列表
}

type RagUpdateDocChildSegmentParams struct {
	UserId          string      `json:"userId"`            // 发起请求的用户ID
	KnowledgeBase   string      `json:"knowledgeBase"`     // 知识库的名称
	KnowledgeId     string      `json:"kb_id"`             // 知识库的唯一ID
	FileName        string      `json:"fileName"`          // 与chunk关联的文件名
	ChunkId         string      `json:"chunk_id"`          // 父分段id
	ChunkCurrentNum int32       `json:"chunk_current_num"` // 父分段序列号
	ChildChunk      *ChildChunk `json:"child_chunk"`       // 子分段内容

}

type ChildChunk struct {
	ChildContent string `json:"child_content"`           //子分段内容
	ChildChunkNo int32  `json:"child_chunk_current_num"` // 子分段序列号
}

type RagUpdateDocSegmentParams struct {
	UserId           string            `json:"userId"`             // 发起请求的用户ID
	KnowledgeBase    string            `json:"knowledgeBase"`      // 知识库的名称
	KnowledgeId      string            `json:"kb_id"`              // 知识库的唯一ID
	FileName         string            `json:"fileName"`           // 与chunk关联的文件名
	MaxSentenceSize  int               `json:"max_sentence_size"`  // 最大分段长度限制
	Chunk            *UpdateChunkItem  `json:"chunk"`              // 分段数据列表
	SplitType        string            `json:"split_type"`         // 分段类型 //parent_child|common
	ChildChunkConfig *ChildChunkConfig `json:"child_chunk_config"` //子分段配置
}

type ChildChunkConfig struct {
	Separators []string `json:"separators"` // 分隔符
	ChunkSize  int32    `json:"chunk_size"` // 子分段大小
}

type RagDeleteDocSegmentParams struct {
	UserId        string   `json:"userId"`        // 发起请求的用户ID
	KnowledgeBase string   `json:"knowledgeBase"` // 知识库的名称
	KnowledgeId   string   `json:"kb_id"`         // 知识库的唯一ID
	FileName      string   `json:"fileName"`      // 与chunk关联的文件名
	ChunkIds      []string `json:"chunk_ids"`     // 分段数据列表
}

type RagGetDocSegmentParams struct {
	UserId            string `json:"userId"`
	KnowledgeBaseName string `json:"knowledgeBase"`
	KnowledgeBaseId   string `json:"kb_id"`
	FileName          string `json:"fileName"`
	PageSize          int32  `json:"page_size"`
	SearchAfter       int32  `json:"search_after"`
	QueryText         string `json:"query_text"`
}

type RagGetDocChildSegmentParams struct {
	UserId            string `json:"userId"`        // 用户id
	KnowledgeBaseName string `json:"knowledgeBase"` // 知识库名称
	KnowledgeId       string `json:"kb_id"`         // 知识库id
	FileName          string `json:"file_name"`     // 文件名
	ChunkId           string `json:"chunk_id"`      // 使用父分段的contentId
}

type DocSegmentStatusUpdateParams struct {
	UserId          string `json:"userId"`
	KnowledgeBaseId string `json:"kb_id"`
	KnowledgeName   string `json:"knowledgeBase"`
	FileName        string `json:"fileName"`
	ContentId       string `json:"content_id"`
	Status          bool   `json:"status"`
}

type DocSegmentStatusUpdateAllParams struct {
	DocSegmentStatusUpdateParams
	All bool `json:"on_off_switch"`
}

type RagGetDocSegmentResp struct {
	RagCommonResp
	Data *ContentListResp `json:"data"`
}

type ContentListResp struct {
	List          []FileSplitContent `json:"content_list"`
	ChunkTotalNum int                `json:"chunk_total_num"`
}

type FileSplitContent struct {
	Content            string          `json:"content"`
	Order              int             `json:"order"`
	Status             bool            `json:"status"`
	MetaData           ContentMetaData `json:"meta_data"`
	ContentId          string          `json:"content_id"`
	UserId             string          `json:"userId"`
	KbName             string          `json:"kb_name"`
	FileName           string          `json:"file_name"`
	Labels             []string        `json:"labels"`
	ChunkId            string          `json:"chunk_id"`
	OssPath            string          `json:"oss_path"`
	IsParent           bool            `json:"is_parent"`             // 区分是否是父分段，true是父分段，false是子分段,不存在这个key时说明文档分段模式不是父子分段
	ChildChunkTotalNum int             `json:"child_chunk_total_num"` // 父分段对应子分段数量
}

type ContentMetaData struct {
	FileName        string         `json:"file_name"`
	ChunkCurrentNum int            `json:"chunk_current_num"`
	ChunkTotalNum   int            `json:"chunk_total_num"`
	DownloadLink    string         `json:"download_link"`
	BucketName      string         `json:"bucket_name"`
	ObjectName      string         `json:"object_name"`
	DocMeta         []*DocMetaData `json:"doc_meta"`
}

type RagGetDocChildSegmentResp struct {
	RagCommonResp
	Data *ChildContentListResp `json:"data"`
}

type ChildContentListResp struct {
	ParentChunkId      string                  `json:"parent_chunk_id"`
	ChildChunkTotalNum int                     `json:"child_chunk_total_num"` // 以这个字段为准
	ChildContentList   []ChildFileSplitContent `json:"child_content_list"`
}

type ChildFileSplitContent struct {
	Content         string           `json:"content"`
	ChunkId         string           `json:"chunk_id"` // 尽量不用
	FileName        string           `json:"file_name"`
	OssPath         string           `json:"oss_path"`
	MetaData        ChildContentMeta `json:"meta_data"`
	Status          bool             `json:"status"`
	ContentId       string           `json:"content_id"`
	ParentContentId string           `json:"parent_content_id"`
	KnowledgeName   string           `json:"kb_name"`
	IsParent        bool             `json:"is_parent"` // false是子分段
}

type ChildContentMeta struct {
	FileName             string         `json:"file_name"`
	ChildChunkCurrentNum int            `json:"child_chunk_current_num"`
	ChildChunkTotalNum   int            `json:"child_chunk_total_num"`
	DownloadLink         string         `json:"download_link"`
	BucketName           string         `json:"bucket_name"`
	ObjectName           string         `json:"object_name"`
	DocMeta              []*DocMetaData `json:"doc_meta"`
}

// RagCreateDocSegment 新增文档切片
func RagCreateDocSegment(ctx context.Context, ragCreateDocSegmentParams *RagCreateDocSegmentParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocSegmentCreateUri
	paramsByte, err := json.Marshal(ragCreateDocSegmentParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_segment_create",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagDocSegmentResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag create doc segment unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagUpdateDocSegment 更新文档切片
func RagUpdateDocSegment(ctx context.Context, ragUpdateDocSegmentParams *RagUpdateDocSegmentParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocSegmentUpdateUri
	paramsByte, err := json.Marshal(ragUpdateDocSegmentParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_segment_update",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag update doc segment unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagDeleteDocSegment 删除文档切片
func RagDeleteDocSegment(ctx context.Context, ragDeleteDocSegmentParams *RagDeleteDocSegmentParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocSegmentDeleteUri
	paramsByte, err := json.Marshal(ragDeleteDocSegmentParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_segment_delete",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagDocSegmentResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag delete doc segment unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagDocUpdateDocSegmentStatus 更新文档切片状态
func RagDocUpdateDocSegmentStatus(ctx context.Context, docSegmentStatusUpdateParams interface{}) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocSegmentUpdateStatusUri
	paramsByte, err := json.Marshal(docSegmentStatusUpdateParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_segment_update_status",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag segment update unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagGetDocSegmentList rag获取知识库文档分片
func RagGetDocSegmentList(ctx context.Context, ragGetDocSegmentParams *RagGetDocSegmentParams) (*ContentListResp, error) {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.GetDocSegmentUri
	paramsByte, err := json.Marshal(ragGetDocSegmentParams)
	if err != nil {
		return nil, err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_get_doc_segment",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagGetDocSegmentResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	// 排序
	slices.SortFunc(resp.Data.List, func(a, b FileSplitContent) int {
		return cmp.Compare(a.MetaData.ChunkCurrentNum, b.MetaData.ChunkCurrentNum)
	})
	return resp.Data, nil
}

// RagCreateDocChildSegment 新增文档子切片
func RagCreateDocChildSegment(ctx context.Context, ragCreateDocChildSegmentParams *RagCreateDocChildSegmentParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocChildSegmentCreateUri
	paramsByte, err := json.Marshal(ragCreateDocChildSegmentParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_child_segment_create",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag create doc child segment unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagUpdateDocChildSegment 更新文档子切片
func RagUpdateDocChildSegment(ctx context.Context, ragUpdateDocChildSegmentParams *RagUpdateDocChildSegmentParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocChildSegmentUpdateUri
	paramsByte, err := json.Marshal(ragUpdateDocChildSegmentParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_child_segment_update",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag update doc child segment unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagDeleteDocChildSegment 删除文档子切片
func RagDeleteDocChildSegment(ctx context.Context, ragDeleteDocChildSegmentParams *RagDeleteDocChildSegmentParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocChildSegmentDeleteUri
	paramsByte, err := json.Marshal(ragDeleteDocChildSegmentParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_child_segment_delete",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag delete doc child segment unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

func RagGetDocChildSegmentList(ctx context.Context, ragGetDocChildSegmentParams *RagGetDocChildSegmentParams) (*ChildContentListResp, error) {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.GetDocChildSegmentUri
	paramsByte, err := json.Marshal(ragGetDocChildSegmentParams)
	if err != nil {
		return nil, err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_get_doc_child_segment",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagGetDocChildSegmentResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	if resp.Data == nil {
		return nil, errors.New("doc child segment response is empty")
	}
	if len(resp.Data.ChildContentList) > 0 {
		// 按排序
		slices.SortFunc(resp.Data.ChildContentList, func(a, b ChildFileSplitContent) int {
			return cmp.Compare(a.MetaData.ChildChunkCurrentNum, b.MetaData.ChildChunkCurrentNum)
		})
	}

	return resp.Data, nil
}
