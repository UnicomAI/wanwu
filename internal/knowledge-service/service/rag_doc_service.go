package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/http"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/mq"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	SplitByDesign   string = "split_by_design"
	SplitByDefault  string = "split_by_default"
	SplitTypeCommon string = "common"
	SplitTypeParent string = "parent_child"

	RagCommunityReport string = "community_report"
)

type RagOperationParams struct {
	Operation string              `json:"operation"`
	Type      string              `json:"type"`
	Doc       *RagImportDocParams `json:"doc"`
}

type RagMetaDataParams struct {
	MetaId    string      `json:"meta_id"`    // 元数据id
	Key       string      `json:"key"`        // key
	Value     interface{} `json:"value"`      // 常量
	ValueType string      `json:"value_type"` // 常量类型
	Rule      string      `json:"rule"`       // 正则表达式
}

type RagChunkConfig struct {
	SegmentSize int      `json:"chunk_size"`
	Separators  []string `json:"separators"`
}

type RagImportDocParams struct {
	DocId                 string               `json:"id"`         //文档id
	KnowledgeName         string               `json:"categoryId"` //知识库名称
	CategoryId            string               `json:"kb_id"`      //知识库id
	IsEnhanced            string               `json:"is_enhanced"`
	UserId                string               `json:"userId"`
	Overlap               float32              `json:"overlap" `
	ObjectName            string               `json:"objectName"`
	SegmentSize           int                  `json:"chunk_size"`
	OriginalName          string               `json:"originalName"`
	SegmentType           string               `json:"chunk_type"`
	SplitType             string               `json:"split_type"` //parent_child|common
	Separators            []string             `json:"separators"`
	ParserChoices         []string             `json:"parser_choices"`
	OcrModelId            string               `json:"ocr_model_id"`
	PreProcess            []string             `json:"pre_process"`
	RagMetaDataParams     []*RagMetaDataParams `json:"meta_data"`
	RagChildChunkConfig   *RagChunkConfig      `json:"child_chunk_config"`
	KnowledgeGraphSwitch  bool                 `json:"enable_knowledge_graph"`
	GraphModelId          string               `json:"graph_model_id"`
	GraphSchemaObjectName string               `json:"graph_schema_objectname"`
	GraphSchemaFileName   string               `json:"graph_schema_filename"`
	MessageType           string               `json:"message_type,omitempty"`
}

type RagImportUrlDocParams struct {
	Url               string               `json:"url"`
	FileName          string               `json:"file_name"`
	Overlap           float32              `json:"overlap_size" `
	SegmentSize       int                  `json:"sentence_size"`
	SegmentType       string               `json:"chunk_type"`
	SplitType         string               `json:"split_type"` //parent_child|common
	UserId            string               `json:"userId"`
	KnowledgeBaseName string               `json:"knowledgeBase"`
	IsEnhanced        bool                 `json:"is_enhanced"`
	Separators        []string             `json:"separators"`
	TaskId            string               `json:"task_id"`
	OcrModelId        string               `json:"ocr_model_id"`
	PreProcess        []string             `json:"pre_process"`
	RagMetaDataParams []*RagMetaDataParams `json:"meta_data"`
}

type RagDeleteDocParams struct {
	UserId        string `json:"userId"`
	KnowledgeBase string `json:"knowledgeBase"`
	FileName      string `json:"fileName"`
}

type RagDocMetaParams struct {
	UserId        string      `json:"userId"`
	KnowledgeBase string      `json:"knowledgeBase"`
	FileName      string      `json:"fileName"`
	MetaList      []*MetaData `json:"tags"`
}

type BatchRagDocMetaParams struct {
	UserId        string         `json:"userId"`
	KnowledgeBase string         `json:"knowledgeBase"`
	KnowledgeId   string         `json:"kb_id"`
	MetaList      []*DocMetaInfo `json:"metas"`
}

type DocMetaInfo struct {
	FileName     string      `json:"file_name"`
	MetaDataList []*MetaData `json:"metadata_list"`
}

type MetaData struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ValueType string      `json:"value_type"`
}

type RagDocSegmentLabelsParams struct {
	UserId        string   `json:"userId"`        // 发起请求的用户ID
	KnowledgeBase string   `json:"knowledgeBase"` // 知识库的名称
	KnowledgeId   string   `json:"kb_id"`         // 知识库的唯一ID
	FileName      string   `json:"fileName"`      // 与chunk关联的文件名
	ContentId     string   `json:"chunk_id"`      // 要更新标签的chunk的唯一ID
	Labels        []string `json:"labels"`        // 需要为该chunk关联的标签列表
}

type UpdateChunkItem struct {
	ChunkId string   `json:"chunk_id"`
	Content string   `json:"content"`
	Labels  []string `json:"labels"`
}

type NewChunkItem struct {
	Content string   `json:"content"`
	Labels  []string `json:"labels"`
}

type DocMetaData struct {
	ValueType   string `json:"value_type"`
	StringValue string `json:"string_value"`
	Key         string `json:"key"`
}

type DocUrlParams struct {
	Url string `json:"url"`
}

type DocUrlResp struct {
	Url          string        `json:"url"`
	OldName      string        `json:"old_name"`
	FileName     string        `json:"file_name"`
	FileSize     float64       `json:"file_size"`
	ResponseInfo RagCommonResp `json:"response_info"`
}

// DeepCopy 深copy一下
func (d *DocUrlResp) DeepCopy() *DocUrlResp {
	return &DocUrlResp{
		Url:      d.Url,
		OldName:  d.OldName,
		FileName: d.FileName,
		FileSize: d.FileSize,
		ResponseInfo: RagCommonResp{
			Code:    d.ResponseInfo.Code,
			Message: d.ResponseInfo.Message,
		},
	}
}

type DocUrlRespSafeArray struct {
	deduplicationMap map[string]*DocUrlResp
	data             []*DocUrlResp
	mu               sync.Mutex
}

func (sa *DocUrlRespSafeArray) Contains(url string) *DocUrlResp {
	if len(sa.deduplicationMap) == 0 {
		return nil
	}
	return sa.deduplicationMap[url]
}

func (sa *DocUrlRespSafeArray) Append(value *DocUrlResp) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	if len(sa.deduplicationMap) == 0 {
		sa.deduplicationMap = make(map[string]*DocUrlResp)
		sa.deduplicationMap[value.Url] = value
	}
	sa.data = append(sa.data, value)
}

func (sa *DocUrlRespSafeArray) Get(index int) interface{} {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	if index < 0 || index >= len(sa.data) {
		return nil
	}
	return sa.data[index]
}

func (sa *DocUrlRespSafeArray) Len() int {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	return len(sa.data)
}

// RagImportDoc 导入具体文档
func RagImportDoc(ctx context.Context, ragImportDocParams *RagImportDocParams) error {
	return mq.SendMessage(&RagOperationParams{
		Operation: "add",
		Type:      "doc",
		Doc:       ragImportDocParams,
	}, config.GetConfig().Topic.Topic)
}

// RagBuildKnowledgeGraph 构建知识库图谱
func RagBuildKnowledgeGraph(ctx context.Context, ragImportDocParams *RagImportDocParams) error {
	return mq.SendMessage(&RagOperationParams{
		Operation: "add",
		Type:      "doc",
		Doc:       ragImportDocParams,
	}, config.GetConfig().Topic.KnowledgeGraphTopic)
}

// RagImportUrlDoc 导入url文档
func RagImportUrlDoc(ctx context.Context, ragImportDocParams *RagImportUrlDocParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.UrlImportEndpoint + ragServer.DocUrlImportUri
	paramsByte, err := json.Marshal(ragImportDocParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_url_import",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err = json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		if strings.Contains(resp.Message, "文档不存在") {
			return nil
		}
		return errors.New(resp.Message)
	}
	return nil
}

// RagDeleteDoc 删除具体文档
func RagDeleteDoc(ctx context.Context, ragDeleteDocParams *RagDeleteDocParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocDeleteUri
	paramsByte, err := json.Marshal(ragDeleteDocParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_delete",
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
		if strings.Contains(resp.Message, "文档不存在") {
			return nil
		}
		return errors.New(resp.Message)
	}
	return nil
}

// BatchRagDocMeta 更新文档元数据
func BatchRagDocMeta(ctx context.Context, batchRagDocTagParams *BatchRagDocMetaParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.UpdateFileMetasUri
	paramsByte, err := json.Marshal(batchRagDocTagParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "batch_rag_doc_meta",
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

func BatchRagDocUrlAnalysis(ctx context.Context, urlList []string) ([]*DocUrlResp, error) {
	var resultArray = DocUrlRespSafeArray{}
	urlSlice := util.ChunkSlice(urlList, 5)
	for _, urlBatch := range urlSlice {
		err := batchDocUrlAnalysis(ctx, urlBatch, &resultArray)
		if err != nil {
			log.Errorf(err.Error())
			return nil, err
		}
	}
	if resultArray.Len() == 0 {
		return nil, errors.New("解析url失败")
	}
	return resultArray.data, nil
}

// RagDocUrlAnalysis 文档url解析
func RagDocUrlAnalysis(ctx context.Context, docUrlParams *DocUrlParams) (*DocUrlResp, error) {
	ragServer := config.GetConfig().RagServer
	url := ragServer.UrlAnalysisEndpoint + ragServer.DocUrlAnalysisUri
	paramsByte, err := json.Marshal(docUrlParams)
	if err != nil {
		return nil, err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_analysis_uri",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp = &DocUrlResp{}
	if err := json.Unmarshal(result, resp); err != nil {
		log.Errorf("rag doc analysis unmarshal err: %v", err.Error())
		return nil, err
	}
	if resp.ResponseInfo.Code != successCode {
		return nil, errors.New(resp.ResponseInfo.Message)
	}
	if len(resp.FileName) == 0 {
		return nil, errors.New("解析文件失败")
	}
	resp.Url = docUrlParams.Url
	return resp, nil
}

// RagDocSegmentLabels 更新文档切片标签
func RagDocSegmentLabels(ctx context.Context, ragDocSegLabelsParams *RagDocSegmentLabelsParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DocSegmentUpdateLabelsUri
	paramsByte, err := json.Marshal(ragDocSegLabelsParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_doc_segment_tag",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("rag update doc segment labels unmarshal err: %v", err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RebuildSplitType 转换分段方法
func RebuildSplitType(segmentMethod string) string {
	if segmentMethod == model.ParentSegmentMethod {
		return SplitTypeParent
	}
	return SplitTypeCommon
}

func RebuildChildChunkConfig(segmentMethod string, subMaxSplitter int, subSplitter []string) *ChildChunkConfig {
	if segmentMethod == model.CommonSegmentMethod || len(segmentMethod) == 0 {
		return nil
	}
	return &ChildChunkConfig{
		ChunkSize:  int32(subMaxSplitter),
		Separators: subSplitter,
	}
}

// RebuildSegmentType 转换分段类型
func RebuildSegmentType(segmentType string, segmentMethod string) string {
	if segmentMethod == model.ParentSegmentMethod {
		return SplitByDesign
	}
	if segmentType == "0" {
		return SplitByDefault
	}
	return SplitByDesign
}

func RebuildFileName(docId, docType, docName string) string {
	if docType == "url" {
		return docId + ".txt"
	}
	return docName
}

func batchDocUrlAnalysis(ctx context.Context, urlList []string, resultArray *DocUrlRespSafeArray) error {
	var wg = &sync.WaitGroup{}
	for _, url := range urlList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//查询url解析结果
			docUrlResp, err := searchUrlAnalysisResult(ctx, resultArray, url)
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			resultArray.Append(docUrlResp)
		}()
	}
	wg.Wait()
	return nil
}

func searchUrlAnalysisResult(ctx context.Context, resultArray *DocUrlRespSafeArray, url string) (*DocUrlResp, error) {
	//处理同一文件重复url问题
	docUrlResp := resultArray.Contains(url)
	if docUrlResp != nil {
		return docUrlResp.DeepCopy(), nil
	}
	return RagDocUrlAnalysis(ctx, &DocUrlParams{
		Url: url,
	})
}
