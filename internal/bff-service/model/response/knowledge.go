package response

import (
	"encoding/json"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

type KnowledgeListResp struct {
	KnowledgeList []*KnowledgeInfo `json:"knowledgeList"`
}

type CreateKnowledgeResp struct {
	KnowledgeId string `json:"knowledgeId"`
}

type KnowledgeHitResp struct {
	Prompt     string             `json:"prompt"`     //提示词列表
	SearchList []*ChunkSearchList `json:"searchList"` //种种结果
	Score      []float64          `json:"score"`      //打分信息
	UseGraph   bool               `json:"useGraph"`   //是否使用知识图谱
}

type RagKnowledgeResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func CommonRagKnowledgeError(err error) ([]byte, int) {
	resp := RagKnowledgeResp{Code: 1, Message: err.Error()}
	marshal, err := json.Marshal(resp)
	if err != nil {
		return []byte(err.Error()), http.StatusBadRequest
	}
	return marshal, http.StatusBadRequest
}

type EmbeddingModelInfo struct {
	ModelId string `json:"modelId"`
}

type KnowledgeInfo struct {
	KnowledgeId           string                 `json:"knowledgeId"`        //知识库id
	Name                  string                 `json:"name"`               //知识库名称
	OrgName               string                 `json:"orgName"`            //知识库所属名称
	Description           string                 `json:"description"`        //知识库描述
	DocCount              int                    `json:"docCount"`           //文档数量
	EmbeddingModelInfo    *EmbeddingModelInfo    `json:"embeddingModelInfo"` //embedding模型信息
	KnowledgeTagList      []*KnowledgeTag        `json:"knowledgeTagList"`   //知识库标签列表
	CreateUserId          string                 `json:"createUserId"`
	CreateAt              string                 `json:"createAt"`              //创建时间
	PermissionType        int32                  `json:"permissionType"`        //权限类型:0: 查看权限; 10: 编辑权限; 20: 授权权限,数值不连续的原因防止后续有中间权限，目前逻辑 授权权限>编辑权限>查看权限
	Share                 bool                   `json:"share"`                 //是分享，还是私有
	RagName               string                 `json:"ragName"`               //rag名称
	GraphSwitch           int32                  `json:"graphSwitch"`           //图谱开关
	Category              int32                  `json:"category"`              // 0: 知识库; 1: 问答库
	LlmModelId            string                 `json:"llmModelId"`            // 知识图谱模型id
	UpdatedAt             string                 `json:"updatedAt"`             // 更新时间
	External              int32                  `json:"external"`              // 0: 内部知识库 1：外部知识库
	ExternalKnowledgeInfo *KnowledgeExternalInfo `json:"externalKnowledgeInfo"` //外部知识库信息
	Avatar                request.Avatar         `json:"avatar"`                // 头像
}

type KnowledgeExternalInfo struct {
	ExternalKnowledgeId   string `json:"externalKnowledgeId"`   // 外部知识库id
	ExternalKnowledgeName string `json:"externalKnowledgeName"` // 外部知识库名称
	ExternalSource        string `json:"externalSource"`        // 外部知识库来源
	ExternalApiId         string `json:"externalApiId"`         // 外部知识库API id
	ExternalApiName       string `json:"externalApiName"`       // 外部知识库API名称
}

type KnowledgeMetaData struct {
	Key  string `json:"key"`  // key
	Type string `json:"type"` // type(time, string, number)
}

type ChunkSearchList struct {
	Title            string          `json:"title"`
	Snippet          string          `json:"snippet"`
	KnowledgeName    string          `json:"knowledgeName"`
	ChildContentList []*ChildContent `json:"childContentList"`
	ChildScore       []float64       `json:"childScore"`
	ContentType      string          `json:"contentType"` // graph：知识图谱（文本）, text：文档分段（文本）, community_report：社区报告（markdown），qa：问答库（文本）
	Score            float64         `json:"score"`
	RerankInfo       []*RerankInfo   `json:"rerankInfo"`
	// MetaData 命中片段元信息，对象内容原样透传 rag 返回（rag 侧字段名为 meta_data），如 file_name/download_link/page_num/chunk_total_num 等
	MetaData interface{} `json:"metaData,omitempty"`
}

type RerankInfo struct {
	Type    string  `json:"type"`
	FileUrl string  `json:"fileUrl"`
	Score   float64 `json:"score"`
}

type ChildContent struct {
	ChildSnippet string  `json:"childSnippet"`
	Score        float64 `json:"score"`
}

type GetKnowledgeMetaSelectResp struct {
	MetaList []*KnowledgeMetaItem `json:"knowledgeMetaList"`
}

type KnowledgeMetaItem struct {
	MetaId        string `json:"metaId"`
	MetaKey       string `json:"metaKey"`
	MetaValueType string `json:"metaValueType"`
	MetaValue     string `json:"metaValue"` // 确定值
}

type KnowledgeMetaValueListResp struct {
	KnowledgeMetaValues []*KnowledgeMetaValues `json:"knowledgeMetaValues"`
}

type KnowledgeMetaValues struct {
	MetaId        string   `json:"metaId"`
	MetaKey       string   `json:"metaKey"`
	MetaValue     []string `json:"metaValue"` // 确定值
	MetaValueType string   `json:"metaValueType"`
}

type KnowledgeGraphResp struct {
	ProcessingCount int32                 `json:"processingCount"` //处理中
	SuccessCount    int32                 `json:"successCount"`    //成功数量
	FailCount       int32                 `json:"failCount"`       //失败数量
	Total           int32                 `json:"total"`           //总数
	Graph           *KnowledgeGraphSchema `json:"graph"`           //知识图谱节点、边
}

type KnowledgeGraphSchema struct {
	Directed  bool                        `json:"directed"`
	MutiGraph bool                        `json:"mutigraph"`
	Graph     *KnowledgeGraphSourceIdList `json:"graph"`
	Nodes     []*KnowledgeGraphNode       `json:"nodes"`
	Edges     []*KnowledgeGraphEdge       `json:"edges"`
}

type KnowledgeGraphSourceIdList struct {
	SourceIdList []string `json:"source_id"`
}

type KnowledgeGraphNode struct {
	EntityName  string   `json:"entity_name"`
	EntityType  string   `json:"entity_type"`
	Description string   `json:"description"`
	SourceId    []string `json:"source_id"`
	Rank        int32    `json:"rank"`
	PageRank    float64  `json:"pagerank"`
}

type KnowledgeGraphEdge struct {
	SourceEntity string   `json:"source_entity"`
	TargetEntity string   `json:"target_entity"`
	Description  string   `json:"description"`
	Weight       float64  `json:"weight"`
	SourceId     []string `json:"source_id"`
}

type KnowledgeExportRecordPageResult struct {
	List     []*ListKnowledgeExportRecordResp `json:"list"`
	Total    int64                            `json:"total"`
	PageNo   int                              `json:"pageNo"`
	PageSize int                              `json:"pageSize"`
}

type ListKnowledgeExportRecordResp struct {
	ExportRecordId string `json:"exportRecordId"` //知识库导出记录id
	Author         string `json:"author"`         //导出人
	ExportTime     string `json:"exportTime"`     //导出时间
	FilePath       string `json:"filePath"`       //导出文件路径
	Status         int    `json:"status"`         //状态
	ErrorMsg       string `json:"errorMsg"`       //导出状态错误信息
	KnowledgeName  string `json:"knowledgeName"`
}

type CreateKnowledgeExternalAPIResp struct {
	ExternalAPIId string `json:"externalApiId"` // 外部知识库API id
}

type KnowledgeExternalAPIInfo struct {
	ExternalAPIId string `json:"externalApiId"` // 外部知识库API id
	Name          string `json:"name" `         // 外部知识库API 名称
	Description   string `json:"description"`   // 外部知识库API 描述
	BaseUrl       string `json:"baseUrl"`       // 外部知识库API endpoint
	ApiKey        string `json:"apiKey"`        // 外部知识库API Key
}

type KnowledgeExternalAPIListResp struct {
	ExternalAPIList []*KnowledgeExternalAPIInfo `json:"externalApiList"` // 外部知识库API列表
}

type KnowledgeExternalListResp struct {
	ExternalKnowledgeList []*KnowledgeExternalBriefInfo `json:"externalKnowledgeList"` // 外部知识库列表
}

type KnowledgeExternalBriefInfo struct {
	ExternalKnowledgeId   string `json:"externalKnowledgeId"`   // 外部知识库id
	ExternalKnowledgeName string `json:"externalKnowledgeName"` // 外部知识库名称
	ExternalApiId         string `json:"externalApiId"`         // 外部知识库API id
}

type CreateKnowledgeExternalResp struct {
	KnowledgeId string `json:"knowledgeId"` //知识库id
}

// GetDocByKnowledgeNameAndDocNameResp 根据知识库名称和文档名称获取文档信息响应
type GetDocByKnowledgeNameAndDocNameResp struct {
	KnowledgeId   string `json:"knowledgeId"`   // 知识库ID
	KnowledgeName string `json:"knowledgeName"` // 知识库名称
	DocId         string `json:"id"`            // 文档ID
	DocType       string `json:"type"`          // 文档类型
	DocName       string `json:"name"`          // 文档名称
	Disable       bool   `json:"disable"`       // 禁用模式，默认false
}
