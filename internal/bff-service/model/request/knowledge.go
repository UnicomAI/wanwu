package request

import (
	"errors"

	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	CategoryKnowledge           = 0
	CategoryQA                  = 1
	CategoryMultimodalKnowledge = 2 //多模态知识库

	MetaOptionAdd    = "add"
	MetaOptionUpdate = "update"
	MetaOptionDelete = "delete"

	MetaValueTypeString = "string"
	MetaValueTypeNumber = "number"
	MetaValueTypeTime   = "time"

	// MetaOperationMaxNum knowledge-service 对超出部分是静默截断，openapi 侧直接报错
	MetaOperationMaxNum = 100
)

func isValidMetaValueType(valueType string) bool {
	return valueType == MetaValueTypeString || valueType == MetaValueTypeNumber || valueType == MetaValueTypeTime
}

type KnowledgeSelectReq struct {
	Name      string   `json:"name" form:"name" `
	TagIdList []string `json:"tagId" form:"tagId" `
	Category  int32    `json:"category" form:"category"`  // 0:知识库，1:问答库
	External  int32    `json:"external" form:"external" ` // -1:全部，0:内部知识库，1:外部知识库
	CommonCheck
}

type KnowledgeBatchSelectReq struct {
	KnowledgeIdList []string `json:"knowledgeIdList" form:"knowledgeIdList" `
	UserId          string   `json:"userId" form:"userId" `
	CommonCheck
}

type CreateKnowledgeReq struct {
	Avatar         Avatar          `json:"avatar"` // 图标
	Name           string          `json:"name"  validate:"required"`
	Description    string          `json:"description"`
	EmbeddingModel *EmbeddingModel `json:"embeddingModelInfo" validate:"required"`
	KnowledgeGraph *KnowledgeGraph `json:"knowledgeGraph"`
	Category       int32           `json:"category" form:"category"` // 0:知识库，1:问答库,2:多模态知识库
}

type UpdateKnowledgeReq struct {
	KnowledgeId string `json:"knowledgeId"   validate:"required"`
	Name        string `json:"name"   validate:"required"`
	Description string `json:"description"`
	Avatar      Avatar `json:"avatar"` // 图标
	CommonCheck
}

type KnowledgeHitReq struct {
	KnowledgeList        []*AppKnowledgeBase   `json:"knowledgeList"`
	KnowledgeMatchParams *KnowledgeMatchParams `json:"knowledgeMatchParams"   validate:"required"`
	Question             string                `json:"question"`
	DocInfo              []*DocInfo            `json:"docInfoList"` //上传文档列表
}

type KnowledgeMatchParams struct {
	MatchType         string  `json:"matchType"  validate:"required"` //matchType：vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本）
	RerankModelId     string  `json:"rerankModelId"`                  //rerank模型id
	PriorityMatch     int32   `json:"priorityMatch"`                  // 权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1
	SemanticsPriority float32 `json:"semanticsPriority"`              // 语义权重
	KeywordPriority   float32 `json:"keywordPriority"`                // 关键词权重
	TopK              int32   `json:"topK"`                           //topK 获取最高的几行
	Threshold         float32 `json:"threshold"`                      //threshold 过滤分数阈值
	TermWeight        float32 `json:"termWeight"`                     // 关键词系数
	TermWeightEnable  bool    `json:"termWeightEnable"`               // 关键词系数开关
	UseGraph          bool    `json:"useGraph"`                       // 是否使用知识图谱
	CommonCheck
}

type KnowledgeAttachment struct {
	FileUrl  string `json:"fileUrl"`  // 文件原始url地址
	FileType string `json:"fileType"` // 文件类型 image: 图片
}

type EmbeddingModel struct {
	ModelId string `json:"modelId"  validate:"required"`
}

// KnowledgeGraph 知识图谱信息
type KnowledgeGraph struct {
	Switch     bool   `json:"switch"`     //知识图谱开关
	LLMModelId string `json:"llmModelId"` //大模型id，开关为true必填
	SchemaUrl  string `json:"schemaUrl"`  //模型schema文件地址，可以为空
}

type DeleteKnowledge struct {
	KnowledgeId string `json:"knowledgeId" validate:"required"`
	CommonCheck
}

type GetKnowledgeReq struct {
	KnowledgeId string `json:"knowledgeId" validate:"required"`
	CommonCheck
}

type CallbackUpdateDocStatusReq struct {
	DocId        string              `json:"id" validate:"required"`
	Status       int32               `json:"status" validate:"required"`
	MetaDataList []*CallbackMetaData `json:"metaDataList"`
	CommonCheck
}

type CallbackUpdateKnowledgeStatusReq struct {
	KnowledgeId  string `json:"knowledgeId" validate:"required"`
	ReportStatus int32  `json:"reportStatus" validate:"required"` //此状态不会是0
	CommonCheck
}

type CallbackMetaData struct {
	Key    string `json:"key"`
	MetaId string `json:"metaId" validate:"required"`
	Value  string `json:"value" validate:"required"`
}

type DocMetaData struct {
	MetaId        string `json:"metaId"`        // 元数据id
	MetaKey       string `json:"metaKey"`       // key
	MetaValue     string `json:"metaValue"`     // 确定值
	MetaValueType string `json:"metaValueType"` // string，number，time
	MetaRule      string `json:"metaRule"`      // 正则表达式
	Option        string `json:"option"`        // option:add(新增)、update(更新)、delete(删除),update 和delete 的时候metaId 不能为空
}

type GetKnowledgeMetaSelectReq struct {
	KnowledgeId string `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	CommonCheck
}

type KnowledgeMetaValueListReq struct {
	KnowledgeId string   `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	DocIdList   []string `json:"docIdList" form:"docIdList" validate:"required" `
	CommonCheck
}

type UpdateMetaValueReq struct {
	KnowledgeId     string         `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	DocIdList       []string       `json:"docIdList"  validate:"required"`
	MetaValueList   []*DocMetaData `json:"metaValueList"`
	ApplyToSelected bool           `json:"applyToSelected"`
}

// KnowledgeMetaKeyOpenapiReq openapi 定义知识库元数据key
type KnowledgeMetaKeyOpenapiReq struct {
	KnowledgeId string                  `json:"knowledgeId" validate:"required"` // 知识库id
	MetaKeyList []*KnowledgeMetaKeyItem `json:"metaKeyList" validate:"required"` // 元数据key操作列表
	CommonCheck
}

type KnowledgeMetaKeyItem struct {
	MetaId        string `json:"metaId"`                     // 元数据id，option 为 update/delete 时必填
	MetaKey       string `json:"metaKey"`                    // key，只能包含小写字母、数字和下划线，且以小写字母开头，option 为 add/update 时必填
	MetaValueType string `json:"metaValueType"`              // 值类型：string/number/time，option 为 add/update 时必填
	Option        string `json:"option" validate:"required"` // 操作类型：add(新增)、update(修改)、delete(删除)
}

// KnowledgeDocMetaValueOpenapiReq openapi 更新文档元数据值
type KnowledgeDocMetaValueOpenapiReq struct {
	KnowledgeId   string                    `json:"knowledgeId" validate:"required"` // 知识库id
	DocIdList     []string                  `json:"docIdList" validate:"required"`   // 文档id列表
	MetaValueList []*KnowledgeMetaValueItem `json:"metaValueList" validate:"required"`
	CommonCheck
}

type KnowledgeMetaValueItem struct {
	MetaKey       string `json:"metaKey" validate:"required"` // key，需已在知识库元数据中定义
	MetaValue     string `json:"metaValue"`                   // 值，option 为 delete 时可不填
	MetaValueType string `json:"metaValueType"`               // 值类型：string/number/time，需与key的定义一致
	Option        string `json:"option" validate:"required"`  // 操作类型：add(不存在则新增、已有则覆盖)、update(仅已存在时更新)、delete(删除)
}

// RagSearchQABaseReq rag知识库查询请求
type RagSearchQABaseReq struct {
	QAUser                      map[string][]*RagQaInfo `json:"QABaseInfo"`
	UserId                      string                  `json:"userId" validate:"required"`
	KnowledgeIdList             []string                `json:"knowledgeIdList,omitempty" validate:"required"`
	Question                    string                  `json:"question"`
	ReturnMeta                  bool                    `json:"returnMeta"`
	Threshold                   float64                 `json:"threshold"`
	TopK                        int64                   `json:"topK"`
	RetrieveMethod              string                  `json:"retrieveMethod"`
	RerankMod                   string                  `json:"rerankMod"`
	RerankModelId               string                  `json:"rerankModelId"`
	MetadataFiltering           bool                    `json:"metadataFiltering"`
	MetadataFilteringConditions []*QAMetadataFilterItem `json:"metadataFilteringConditions"`
	Weight                      *WeightParams           `json:"weights"`
	CommonCheck
}

// RagSearchKnowledgeBaseReq rag知识库查询请求
type RagSearchKnowledgeBaseReq struct {
	KnowledgeUser        map[string][]*RagKnowledgeInfo `json:"knowledge_base_info"`
	UseGraph             bool                           `json:"use_graph"` // 是否启动知识图谱查询
	UserId               string                         `json:"userId" validate:"required"`
	Question             string                         `json:"question" `
	KnowledgeIdList      []string                       `json:"knowledgeIdList,omitempty" validate:"required"`
	Threshold            float64                        `json:"threshold"`
	TopK                 int32                          `json:"topK"`
	RerankModelId        string                         `json:"rerank_model_id"`               // rerankId
	RerankMod            string                         `json:"rerank_mod"`                    // rerank_model:重排序模式，weighted_score：权重搜索
	RetrieveMethod       string                         `json:"retrieve_method"`               // hybrid_search:混合搜索， semantic_search:向量搜索， full_text_search：文本搜索
	Weight               *WeightParams                  `json:"weights"`                       // 权重搜索下的权重配置
	TermWeight           float32                        `json:"term_weight_coefficient"`       // 关键词系数
	MetaFilter           bool                           `json:"metadata_filtering"`            // 元数据过滤开关
	MetaFilterConditions []*MetadataFilterItem          `json:"metadata_filtering_conditions"` // 元数据过滤条件
	AutoCitation         bool                           `json:"auto_citation"`                 // 是否启动知识图谱查询
	RewriteQuery         bool                           `json:"rewrite_query"`                 // 是否query改写
	EnableVision         bool                           `json:"enable_vision"`                 // 召回结果是否包含多模态文件(只在callback层做最后的赋值)
	AttachmentFiles      []*RagKnowledgeAttachment      `json:"attachment_files"`              // 上传的多模态文件
	ReturnMeta           bool                           `json:"return_meta,omitempty"`         // 是否返回元数据
}

type RagKnowledgeAttachment struct {
	FileType string `json:"file_type"`
	FileUrl  string `json:"file_url"`
}

type RagKnowledgeChatReq struct {
	UserId               string                         `json:"userId"`
	KnowledgeUser        map[string][]*RagKnowledgeInfo `json:"knowledge_base_info"`
	KnowledgeIdList      []string                       `json:"knowledgeIdList"` // 知识库id列表
	Question             string                         `json:"question"`
	Threshold            float32                        `json:"threshold"` // Score阈值
	TopK                 int32                          `json:"topK"`
	Stream               bool                           `json:"stream"`
	Chichat              bool                           `json:"chichat"` // 当知识库召回结果为空时是否使用默认话术（兜底），默认为false
	RerankModelId        string                         `json:"rerank_model_id"`
	CustomModelInfo      *CustomModelInfo               `json:"custom_model_info"`
	History              []*HistoryItem                 `json:"history"`
	MaxHistory           int32                          `json:"max_history"`
	RewriteQuery         bool                           `json:"rewrite_query"`   // 是否query改写
	RerankMod            string                         `json:"rerank_mod"`      // rerank_model:重排序模式，weighted_score：权重搜索
	RetrieveMethod       string                         `json:"retrieve_method"` // hybrid_search:混合搜索， semantic_search:向量搜索， full_text_search：文本搜索
	Weight               *WeightParams                  `json:"weights"`         // 权重搜索下的权重配置
	Temperature          float32                        `json:"temperature,omitempty"`
	TopP                 float32                        `json:"top_p,omitempty"`               // 多样性
	RepetitionPenalty    float32                        `json:"repetition_penalty,omitempty"`  // 重复惩罚/频率惩罚
	ReturnMeta           bool                           `json:"return_meta,omitempty"`         // 是否返回元数据
	AutoCitation         bool                           `json:"auto_citation"`                 // 是否自动角标
	TermWeight           float32                        `json:"term_weight_coefficient"`       // 关键词系数
	MetaFilter           bool                           `json:"metadata_filtering"`            // 元数据过滤开关
	MetaFilterConditions []*MetadataFilterItem          `json:"metadata_filtering_conditions"` // 元数据过滤条件
	UseGraph             bool                           `json:"use_graph"`                     // 是否启动知识图谱查询
	EnableVision         bool                           `json:"enable_vision"`                 // 召回结果是否包含多模态文件
	AttachmentFiles      []*RagKnowledgeAttachment      `json:"attachment_files"`              // 上传的多模态文件
	EnableThinking       bool                           `json:"enable_thinking"`               // 是否打开思考过程
	CommonCheck
}

type KnowledgeGraphReq struct {
	KnowledgeId string `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	CommonCheck
}

type CustomModelInfo struct {
	LlmModelID string `json:"llm_model_id"`
}

type HistoryItem struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

type RagKnowledgeInfo struct {
	KnowledgeId   string `json:"kb_id"`
	KnowledgeName string `json:"kb_name"`
}

type RagQaInfo struct {
	QaBaseId   string `json:"QAId"`
	QaBaseName string `json:"QABase"`
}

type WeightParams struct {
	VectorWeight float32 `json:"vector_weight"` //语义权重
	TextWeight   float32 `json:"text_weight"`   //关键字权重
}

type MetadataFilterItem struct {
	FilterKnowledgeName string      `json:"filtering_kb_name"`
	LogicalOperator     string      `json:"logical_operator"`
	Conditions          []*MetaItem `json:"conditions"`
}

type QAMetadataFilterItem struct {
	FilteringQaBaseName string      `json:"filtering_qa_base_name"`
	LogicalOperator     string      `json:"logical_operator"`
	Conditions          []*MetaItem `json:"conditions"`
}

type MetaItem struct {
	MetaName           string      `json:"meta_name"`           // 元数据名称
	MetaType           string      `json:"meta_type"`           // 元数据类型
	ComparisonOperator string      `json:"comparison_operator"` // 比较运算符
	Value              interface{} `json:"value,omitempty"`     // 用于过滤的条件值
}

type CreateKnowledgeExternalAPIReq struct {
	Name        string `json:"name"  validate:"required"`   //外部知识库API名称
	Description string `json:"description"`                 //外部知识库API描述
	BaseUrl     string `json:"baseUrl" validate:"required"` //外部知识库API endpoint
	ApiKey      string `json:"apiKey" validate:"required"`  //外部知识库API Key
}

type UpdateKnowledgeExternalAPIReq struct {
	ExternalAPIId string `json:"externalApiId" validate:"required"` // 外部知识库API id
	Name          string `json:"name"  validate:"required"`         // 外部知识库API名称
	Description   string `json:"description"`                       // 外部知识库API描述
	BaseUrl       string `json:"baseUrl" validate:"required"`       // 外部知识库API endpoint
	ApiKey        string `json:"apiKey" validate:"required"`        // 外部知识库API Key
	CommonCheck
}

type DeleteKnowledgeExternalAPIReq struct {
	ExternalAPIId string `json:"externalApiId" validate:"required"` // 外部知识库API id
	CommonCheck
}

type CreateKnowledgeExternalReq struct {
	Name                string `json:"name"  validate:"required"`               //外部知识库名称
	Description         string `json:"description"`                             //外部知识库描述
	ExternalSource      string `json:"externalSource" validate:"required"`      //外部来源
	ExternalAPIId       string `json:"externalApiId" validate:"required"`       //外部知识库API id
	ExternalKnowledgeId string `json:"externalKnowledgeId" validate:"required"` //外部知识库 id
}

type UpdateKnowledgeExternalReq struct {
	KnowledgeId         string `json:"knowledgeId" validate:"required"`         //知识库id
	Name                string `json:"name"  validate:"required"`               //名称
	Description         string `json:"description"`                             //描述
	ExternalSource      string `json:"externalSource" validate:"required"`      //外部知识库来源
	ExternalAPIId       string `json:"externalApiId" validate:"required"`       //外部知识库API id
	ExternalKnowledgeId string `json:"externalKnowledgeId" validate:"required"` //外部知识库 id
	CommonCheck
}

type DeleteKnowledgeExternalReq struct {
	KnowledgeId string `json:"knowledgeId" validate:"required"` //知识库id
	CommonCheck
}

type KnowledgeExternalListReq struct {
	ExternalAPIId string `json:"externalApiId" form:"externalApiId" validate:"required"` //外部知识库API id
	CommonCheck
}

// GetDocByKnowledgeNameAndDocNameReq 根据知识库名称和文档名称获取文档信息请求
type GetDocByKnowledgeNameAndDocNameReq struct {
	KnowledgeName string `json:"knowledgeName" form:"knowledgeName" validate:"required"` // 知识库名称
	DocName       string `json:"docName" form:"docName" validate:"required"`             // 文档名称
	CommonCheck
}

func (c *KnowledgeMetaKeyOpenapiReq) Check() error {
	if len(c.MetaKeyList) > MetaOperationMaxNum {
		return errors.New("metaKeyList size can not exceed 100")
	}
	keyMap := make(map[string]bool)
	for _, v := range c.MetaKeyList {
		switch v.Option {
		case MetaOptionDelete:
			if v.MetaId == "" {
				return errors.New("metaId can not be empty when option is delete")
			}
			continue
		case MetaOptionUpdate:
			if v.MetaId == "" {
				return errors.New("metaId can not be empty when option is update")
			}
		case MetaOptionAdd:
		default:
			return errors.New("option must be add, update or delete")
		}
		if !isValidKey(v.MetaKey) {
			return errors.New("非法key")
		}
		if !isValidMetaValueType(v.MetaValueType) {
			return errors.New("metaValueType must be string, number or time")
		}
		if keyMap[v.MetaKey] {
			return errors.New("key can not be repeated")
		}
		keyMap[v.MetaKey] = true
	}
	return nil
}

func (c *KnowledgeDocMetaValueOpenapiReq) Check() error {
	if len(c.MetaValueList) > MetaOperationMaxNum {
		return errors.New("metaValueList size can not exceed 100")
	}
	keyMap := make(map[string]bool)
	for _, v := range c.MetaValueList {
		if v.Option != MetaOptionAdd && v.Option != MetaOptionUpdate && v.Option != MetaOptionDelete {
			return errors.New("option must be add, update or delete")
		}
		if !isValidKey(v.MetaKey) {
			return errors.New("非法key")
		}
		if v.Option != MetaOptionDelete && !isValidMetaValueType(v.MetaValueType) {
			return errors.New("metaValueType must be string, number or time")
		}
		if keyMap[v.MetaKey] {
			return errors.New("key can not be repeated")
		}
		keyMap[v.MetaKey] = true
	}
	return nil
}

func (c *UpdateMetaValueReq) Check() error {
	for _, v := range c.MetaValueList {
		if v.Option == "" {
			return errors.New("option为空")
		}
	}
	return nil
}

func (c *CreateKnowledgeReq) Check() error {
	if err := util.ValidateName(&c.Name, util.SubjectKnowledge); err != nil {
		return err
	}
	if err := util.ValidateDesc(&c.Description, util.SubjectKnowledge); err != nil {
		return err
	}
	if c.Category != CategoryKnowledge && c.Category != CategoryQA && c.Category != CategoryMultimodalKnowledge {
		return errors.New("invalid category")
	}
	if c.Category == CategoryKnowledge {
		if c.KnowledgeGraph == nil {
			return errors.New("knowledge graph can not be nil")
		}
		if c.KnowledgeGraph.Switch && c.KnowledgeGraph.LLMModelId == "" {
			return errors.New("knowledge graph llmModelId can not be empty")
		}
	}
	return nil
}

func (c *CreateKnowledgeExternalAPIReq) Check() error {
	if err := util.ValidateName(&c.Name, util.SubjectKnowledgeExternalAPI); err != nil {
		return err
	}
	return util.ValidateDesc(&c.Description, util.SubjectKnowledgeExternalAPI)
}

func (c *CreateKnowledgeExternalReq) Check() error {
	return util.ValidateBriefCreate(&c.Name, &c.Description, util.SubjectKnowledge)
}

func (c *RagSearchKnowledgeBaseReq) Check() error {
	if len(c.AttachmentFiles) == 0 && c.Question == "" {
		return errors.New("both question and attachmentFiles cannot be empty")
	}
	return nil
}

func (c *KnowledgeHitReq) Check() error {
	if len(c.DocInfo) == 0 && c.Question == "" {
		return errors.New("both docInfo and question cannot be empty")
	}
	// 纯图片搜索必须用多模态rerank
	if c.Question == "" {
		if c.KnowledgeMatchParams.RerankModelId == "" {
			return errors.New("只输入图片必须选择多模态reranker")
		}
		if c.KnowledgeMatchParams.MatchType == "text" {
			return errors.New("只输入图片不支持全文检索")
		}
	}
	return nil
}
