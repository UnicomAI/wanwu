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
	Prompt     string             `json:"prompt"`     //Prompt word list
	SearchList []*ChunkSearchList `json:"searchList"` //Various results
	Score      []float64          `json:"score"`      //Rating information
	UseGraph   bool               `json:"useGraph"`   //Whether to use knowledge graph
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
	KnowledgeId        string              `json:"knowledgeId"`        //knowledge base id
	Name               string              `json:"name"`               //Knowledge base name
	OrgName            string              `json:"orgName"`            //Knowledge base name
	Description        string              `json:"description"`        //Knowledge base description
	DocCount           int                 `json:"docCount"`           //Number of documents
	EmbeddingModelInfo *EmbeddingModelInfo `json:"embeddingModelInfo"` //embedding model information
	KnowledgeTagList   []*KnowledgeTag     `json:"knowledgeTagList"`   //Knowledge base tag list
	CreateUserId       string              `json:"createUserId"`
	CreateAt           string              `json:"createAt"`       //creation time
	PermissionType     int32               `json:"permissionType"` //Permission type: 0: View permission; 10: Edit permission; 20: Authorization permission. The reason for discontinuous values ​​prevents intermediate permissions in the future. The current logic is Authorization permission > Edit permission > View permission.
	Share              bool                `json:"share"`          //Is it shared or private?
	RagName            string              `json:"ragName"`        //rag name
	GraphSwitch        int32               `json:"graphSwitch"`    //Map switch
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
	ContentType      string          `json:"contentType"` // graph: knowledge graph (text), text: document segmentation (text), community_report: community report (markdown)
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
	MetaValue     string `json:"metaValue"` // Determine value
}

type KnowledgeMetaValueListResp struct {
	KnowledgeMetaValues []*KnowledgeMetaValues `json:"knowledgeMetaValues"`
}

type KnowledgeMetaValues struct {
	MetaId        string   `json:"metaId"`
	MetaKey       string   `json:"metaKey"`
	MetaValue     []string `json:"metaValue"` // Determine value
	MetaValueType string   `json:"metaValueType"`
}

type KnowledgeGraphResp struct {
	ProcessingCount int32                 `json:"processingCount"` //Processing
	SuccessCount    int32                 `json:"successCount"`    //number of successes
	FailCount       int32                 `json:"failCount"`       //Number of failures
	Total           int32                 `json:"total"`           //total
	Graph           *KnowledgeGraphSchema `json:"graph"`           //Knowledge graph nodes and edges
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
