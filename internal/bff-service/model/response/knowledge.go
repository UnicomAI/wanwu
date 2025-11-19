package response

import (
	"encoding/json"
	"net/http"
)

type KnowledgeListResp struct {
	KnowledgeList []*KnowledgeInfo `json:"knowledgeList"`
}

type CreateKnowledgeResp struct {
	KnowledgeId string `json:"knowledgeId"`
}

type KnowledgeHitResp struct {
	Prompt     string             `json:"prompt"`     //提示词列表 [EN] Prompt word list
	SearchList []*ChunkSearchList `json:"searchList"` //种种结果 [EN] Various results
	Score      []float64          `json:"score"`      //打分信息 [EN] Rating information
	UseGraph   bool               `json:"useGraph"`   //是否使用知识图谱 [EN] Whether to use knowledge graph
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
	KnowledgeId        string              `json:"knowledgeId"`        //知识库id [EN] knowledge base id
	Name               string              `json:"name"`               //知识库名称 [EN] Knowledge base name
	OrgName            string              `json:"orgName"`            //知识库所属名称 [EN] Knowledge base name
	Description        string              `json:"description"`        //知识库描述 [EN] Knowledge base description
	DocCount           int                 `json:"docCount"`           //文档数量 [EN] Number of documents
	EmbeddingModelInfo *EmbeddingModelInfo `json:"embeddingModelInfo"` //embedding模型信息 [EN] embedding model information
	KnowledgeTagList   []*KnowledgeTag     `json:"knowledgeTagList"`   //知识库标签列表 [EN] Knowledge base tag list
	CreateUserId       string              `json:"createUserId"`
	CreateAt           string              `json:"createAt"`       //创建时间 [EN] creation time
	PermissionType     int32               `json:"permissionType"` //权限类型:0: 查看权限; 10: 编辑权限; 20: 授权权限,数值不连续的原因防止后续有中间权限，目前逻辑 授权权限>编辑权限>查看权限 [EN] Permission type: 0: View permission; 10: Edit permission; 20: Authorization permission. The reason for discontinuous values ​​prevents intermediate permissions in the future. The current logic is Authorization permission > Edit permission > View permission.
	Share              bool                `json:"share"`          //是分享，还是私有 [EN] Is it shared or private?
	RagName            string              `json:"ragName"`        //rag名称 [EN] rag name
	GraphSwitch        int32               `json:"graphSwitch"`    //图谱开关 [EN] Map switch
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
	ContentType      string          `json:"contentType"` // graph：知识图谱（文本）, text：文档分段（文本）, community_report：社区报告（markdown） [EN] graph: knowledge graph (text), text: document segmentation (text), community_report: community report (markdown)
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
	MetaValue     string `json:"metaValue"` // 确定值 [EN] Determine value
}

type KnowledgeMetaValueListResp struct {
	KnowledgeMetaValues []*KnowledgeMetaValues `json:"knowledgeMetaValues"`
}

type KnowledgeMetaValues struct {
	MetaId        string   `json:"metaId"`
	MetaKey       string   `json:"metaKey"`
	MetaValue     []string `json:"metaValue"` // 确定值 [EN] Determine value
	MetaValueType string   `json:"metaValueType"`
}

type KnowledgeGraphResp struct {
	ProcessingCount int32                 `json:"processingCount"` //处理中 [EN] Processing
	SuccessCount    int32                 `json:"successCount"`    //成功数量 [EN] number of successes
	FailCount       int32                 `json:"failCount"`       //失败数量 [EN] Number of failures
	Total           int32                 `json:"total"`           //总数 [EN] total
	Graph           *KnowledgeGraphSchema `json:"graph"`           //知识图谱节点、边 [EN] Knowledge graph nodes and edges
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
