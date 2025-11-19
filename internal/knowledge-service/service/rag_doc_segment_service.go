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
	UserId           string            `json:"userId"`             // The user ID that initiated the request
	KnowledgeBase    string            `json:"knowledgeBase"`      // The name of the knowledge base
	KnowledgeId      string            `json:"kb_id"`              // The unique ID of the knowledge base
	FileName         string            `json:"fileName"`           // The file name associated with the chunk
	MaxSentenceSize  int               `json:"max_sentence_size"`  // Maximum segment length limit
	Chunks           []*NewChunkItem   `json:"chunks"`             // Segmented data list
	SplitType        string            `json:"split_type"`         // Segmentation type //parent_child|common
	ChildChunkConfig *ChildChunkConfig `json:"child_chunk_config"` //Subsegment configuration
}

type RagCreateDocChildSegmentParams struct {
	UserId        string   `json:"userId"`         // The user ID that initiated the request
	KnowledgeBase string   `json:"knowledgeBase"`  // The name of the knowledge base
	KnowledgeId   string   `json:"kb_id"`          // The unique ID of the knowledge base
	FileName      string   `json:"fileName"`       // The file name associated with the chunk
	ChunkId       string   `json:"chunk_id"`       // parent segment id
	ChildContents []string `json:"child_contents"` // subsection content
}

type RagDeleteDocChildSegmentParams struct {
	UserId                string  `json:"userId"`                   // The user ID that initiated the request
	KnowledgeBase         string  `json:"knowledgeBase"`            // The name of the knowledge base
	KnowledgeId           string  `json:"kb_id"`                    // The unique ID of the knowledge base
	FileName              string  `json:"fileName"`                 // The file name associated with the chunk
	ChunkId               string  `json:"chunk_id"`                 // parent segment id
	ChunkCurrentNum       int32   `json:"chunk_current_num"`        // Parent segment sequence number
	ChildChunkCurrentNums []int32 `json:"child_chunk_current_nums"` // Subsegment sequence number list
}

type RagUpdateDocChildSegmentParams struct {
	UserId          string      `json:"userId"`            // The user ID that initiated the request
	KnowledgeBase   string      `json:"knowledgeBase"`     // The name of the knowledge base
	KnowledgeId     string      `json:"kb_id"`             // The unique ID of the knowledge base
	FileName        string      `json:"fileName"`          // The file name associated with the chunk
	ChunkId         string      `json:"chunk_id"`          // parent segment id
	ChunkCurrentNum int32       `json:"chunk_current_num"` // Parent segment sequence number
	ChildChunk      *ChildChunk `json:"child_chunk"`       // subsection content

}

type ChildChunk struct {
	ChildContent string `json:"child_content"`           //subsection content
	ChildChunkNo int32  `json:"child_chunk_current_num"` // subsegment sequence number
}

type RagUpdateDocSegmentParams struct {
	UserId           string            `json:"userId"`             // The user ID that initiated the request
	KnowledgeBase    string            `json:"knowledgeBase"`      // The name of the knowledge base
	KnowledgeId      string            `json:"kb_id"`              // The unique ID of the knowledge base
	FileName         string            `json:"fileName"`           // The file name associated with the chunk
	MaxSentenceSize  int               `json:"max_sentence_size"`  // Maximum segment length limit
	Chunk            *UpdateChunkItem  `json:"chunk"`              // Segmented data list
	SplitType        string            `json:"split_type"`         // Segmentation type //parent_child|common
	ChildChunkConfig *ChildChunkConfig `json:"child_chunk_config"` //Subsegment configuration
}

type ChildChunkConfig struct {
	Separators []string `json:"separators"` // delimiter
	ChunkSize  int32    `json:"chunk_size"` // subsegment size
}

type RagDeleteDocSegmentParams struct {
	UserId        string   `json:"userId"`        // The user ID that initiated the request
	KnowledgeBase string   `json:"knowledgeBase"` // The name of the knowledge base
	KnowledgeId   string   `json:"kb_id"`         // The unique ID of the knowledge base
	FileName      string   `json:"fileName"`      // The file name associated with the chunk
	ChunkIds      []string `json:"chunk_ids"`     // Segmented data list
}

type RagGetDocSegmentParams struct {
	UserId            string `json:"userId"`
	KnowledgeBaseName string `json:"knowledgeBase"`
	FileName          string `json:"fileName"`
	PageSize          int32  `json:"page_size"`
	SearchAfter       int32  `json:"search_after"`
}

type RagGetDocChildSegmentParams struct {
	UserId            string `json:"userId"`        // user id
	KnowledgeBaseName string `json:"knowledgeBase"` // Knowledge base name
	KnowledgeId       string `json:"kb_id"`         // knowledge base id
	FileName          string `json:"file_name"`     // file name
	ChunkId           string `json:"chunk_id"`      // Use the contentId of the parent segment
}

type DocSegmentStatusUpdateParams struct {
	UserId        string `json:"userId"`
	KnowledgeName string `json:"knowledgeBase"`
	FileName      string `json:"fileName"`
	ContentId     string `json:"content_id"`
	Status        bool   `json:"status"`
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
	IsParent           bool            `json:"is_parent"`             // Distinguish whether it is a parent segment, true is a parent segment, false is a child segment, if this key does not exist, it means that the document segmentation mode is not a parent-child segmentation
	ChildChunkTotalNum int             `json:"child_chunk_total_num"` // Number of parent segments corresponding to child segments
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
	ChildChunkTotalNum int                     `json:"child_chunk_total_num"` // Based on this field
	ChildContentList   []ChildFileSplitContent `json:"child_content_list"`
}

type ChildFileSplitContent struct {
	Content         string           `json:"content"`
	ChunkId         string           `json:"chunk_id"` // Try not to use it
	FileName        string           `json:"file_name"`
	OssPath         string           `json:"oss_path"`
	MetaData        ChildContentMeta `json:"meta_data"`
	Status          bool             `json:"status"`
	ContentId       string           `json:"content_id"`
	ParentContentId string           `json:"parent_content_id"`
	KnowledgeName   string           `json:"kb_name"`
	IsParent        bool             `json:"is_parent"` // false is a subsegment
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

// RagCreateDocSegment adds a new document slice
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

// RagUpdateDocSegment updates document slices
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

// RagDeleteDocSegment deletes document slices
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

// RagDocUpdateDocSegmentStatus updates document slice status
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

// RagGetDocSegmentList rag gets the knowledge base document fragments
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
	// sort
	slices.SortFunc(resp.Data.List, func(a, b FileSplitContent) int {
		return cmp.Compare(a.MetaData.ChunkCurrentNum, b.MetaData.ChunkCurrentNum)
	})
	return resp.Data, nil
}

// RagCreateDocChildSegment adds a new document sub-slice
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

// RagUpdateDocChildSegment updates document subslices
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

// RagDeleteDocChildSegment deletes document subsegments
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
		// Sort by
		slices.SortFunc(resp.Data.ChildContentList, func(a, b ChildFileSplitContent) int {
			return cmp.Compare(a.MetaData.ChildChunkCurrentNum, b.MetaData.ChildChunkCurrentNum)
		})
	}

	return resp.Data, nil
}
