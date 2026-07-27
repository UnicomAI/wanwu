package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerKnowledge(apiV1 *gin.RouterGroup) {
	// 知识库增删改查
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge", http.MethodPost, v1.CreateKnowledge, "创建知识库（文档分类）", middleware.AuthModelByModelId([]string{"embeddingModelInfo.modelId", "knowledgeGraph.llmModelId"}))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge", http.MethodPut, v1.UpdateKnowledge, "修改知识库（文档分类）", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge", http.MethodDelete, v1.DeleteKnowledge, "删除知识库（文档分类）", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))

	// 知识库命中测试，通用校验不好做改内部校验
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/hit", http.MethodPost, v1.KnowledgeHit, "知识库命中测试", middleware.AuthModelByModelId([]string{"knowledgeMatchParams.rerankModelId"}))

	// 知识库文档
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/detail", http.MethodGet, v1.GetDocKnowledgeDetail, "获取文档知识库信息", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/config", http.MethodGet, v1.GetDocConfig, "获取文档配置信息", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/list", http.MethodPost, v1.GetDocList, "获取文档列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/import", http.MethodPost, v1.ImportDoc, "上传文档", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/update/config", http.MethodPost, v1.UpdateDocConfig, "更新文档配置", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/import/tip", http.MethodGet, v1.GetDocImportTip, "获取知识库文档上传状态", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/export", http.MethodPost, v1.ExportKnowledgeDoc, "知识库文档导出", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/reimport", http.MethodPost, v1.ReImportDoc, "重试导入文档", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/upload/limit", http.MethodGet, v1.GetDocUploadLimit, "获取可上传文件类型", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))

	// 知识库文档，以下通用校验不好做改内部校验
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc", http.MethodDelete, v1.DeleteDoc, "删除文档", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/meta", http.MethodPost, v1.UpdateDocMetaData, "更新文档元数据", middleware.AuthKnowledgeIfHas("knowledgeId", middleware.KnowledgeEdit))

	// 知识库元数据,前端增加knowledgeId
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/meta/select", http.MethodGet, v1.GetKnowledgeMetaKeySelect, "获取知识库元数据key列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/meta/value/list", http.MethodPost, v1.GetKnowledgeMetaValueList, "获取知识库元数据值列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/meta/value/update", http.MethodPost, v1.UpdateKnowledgeMetaValue, "更新知识库元数据值列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// 知识库文档切片增删改查
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/list", http.MethodGet, v1.GetDocSegmentList, "获取文档切分结果", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/status/update", http.MethodPost, v1.UpdateDocSegmentStatus, "更新文档切片启用状态", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/labels", http.MethodPost, v1.UpdateDocSegmentLabels, "更新文档切片标签", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/create", http.MethodPost, v1.CreateDocSegment, "新增文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/batch/create", http.MethodPost, v1.BatchCreateDocSegment, "批量新增文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/delete", http.MethodDelete, v1.DeleteDocSegment, "删除文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/update", http.MethodPost, v1.UpdateDocSegment, "更新文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/child/list", http.MethodGet, v1.GetDocChildSegmentList, "获取子分段列表", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/child/create", http.MethodPost, v1.CreateDocChildSegment, "创建子分段", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/child/update", http.MethodPost, v1.UpdateDocChildSegment, "更新子分段", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/segment/child/delete", http.MethodDelete, v1.DeleteDocChildSegment, "删除子分段", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))

	// 知识库url文档导入(单条url上传/url文件上传)已下架
	// mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/url/analysis", http.MethodPost, v1.AnalysisDocUrl, "解析url")

	// 知识库标签增删改查
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/tag", http.MethodGet, v1.GetKnowledgeTagSelect, "查询知识库标签列表", middleware.AuthKnowledgeIfHas("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/tag", http.MethodPost, v1.CreateKnowledgeTag, "创建知识库标签")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/tag", http.MethodPut, v1.UpdateKnowledgeTag, "修改知识库标签")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/tag", http.MethodDelete, v1.DeleteKnowledgeTag, "删除知识库标签")
	// 绑定知识库标签
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/tag/bind/count", http.MethodGet, v1.SelectKnowledgeTagBindCount, "查询标签绑定的知识库数量")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/tag/bind", http.MethodPost, v1.BindKnowledgeTag, "绑定知识库标签", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// 知识库关键词管理--底层进行了权限knowledgeId 过滤，此处无需处理
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodGet, v1.GetKnowledgeKeywordsList, "查询知识库关键词列表")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodPost, v1.CreateKnowledgeKeywords, "新增知识库关键词")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/keywords/detail", http.MethodGet, v1.GetKnowledgeKeywordsDetail, "查询知识库关键词详情")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodPut, v1.UpdateKnowledgeKeywords, "编辑知识库关键词")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodDelete, v1.DeleteDocCategoryKeywords, "删除知识库关键词")

	// 知识库分隔符增删改查,和用户有关，和知识库无关，无需校验权限
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodGet, v1.GetKnowledgeSplitterSelect, "查询知识库分隔符列表")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodPost, v1.CreateKnowledgeSplitter, "创建知识库分隔符")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodPut, v1.UpdateKnowledgeSplitter, "修改知识库分隔符")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodDelete, v1.DeleteKnowledgeSplitter, "删除知识库分隔符")

	// 知识库权限
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/org", http.MethodGet, v1.SelectKnowledgeOrg, "知识库组织权限列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/user", http.MethodGet, v1.SelectKnowledgeUserPermit, "知识库用户权限列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/user/no/permit", http.MethodGet, v1.SelectKnowledgeUserNoPermit, "没有知识库用户列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/user/add", http.MethodPost, v1.AddKnowledgeUser, "增加知识库用户", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/user/edit", http.MethodPost, v1.EditKnowledgeUser, "修改知识库用户", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/user/delete", http.MethodDelete, v1.DeleteKnowledgeUser, "删除知识库用户", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/user/admin/transfer", http.MethodPost, v1.TransferKnowledgeUserAdmin, "转让管理员权限", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))

	// 知识库社区报告
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/report/list", http.MethodGet, v1.GetKnowledgeReport, "获取社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/report/generate", http.MethodPost, v1.GenerateKnowledgeReport, "生成社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/report/delete", http.MethodDelete, v1.DeleteKnowledgeReport, "删除社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/report/update", http.MethodPost, v1.UpdateKnowledgeReport, "更新社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/report/add", http.MethodPost, v1.AddKnowledgeReport, "单条新增社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/report/batch/add", http.MethodPost, v1.BatchAddKnowledgeReport, "批量新增社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// 知识图谱详情
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/graph", http.MethodGet, v1.GetKnowledgeGraph, "获取知识图谱详情", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))

	// 问答库
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair", http.MethodPost, v1.GreateKnowledgeQAPair, "新增问答对", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair", http.MethodPut, v1.UpdateKnowledgeQAPair, "编辑问答对", middleware.AuthKnowledgeQAPair("qaPairId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair", http.MethodDelete, v1.DeleteKnowledgeQAPair, "删除问答对", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair/switch", http.MethodPut, v1.UpdateKnowledgeQAPairSwitch, "启停问答对", middleware.AuthKnowledgeQAPair("qaPairId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair/list", http.MethodPost, v1.GetKnowledgeQAPairList, "获取问答对列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair/import", http.MethodPost, v1.ImportKnowledgeQAPair, "问答库文档导入", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/pair/import/tip", http.MethodGet, v1.GetKnowledgeQAPairImportTip, "获取问答库文档上传状态", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/export", http.MethodGet, v1.ExportKnowledgeQAPair, "问答库文档导出", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/qa/hit", http.MethodPost, v1.KnowledgeQAHit, "问答库命中测试", middleware.AuthModelByModelId([]string{"knowledgeMatchParams.rerankModelId"}))

	// 知识库导出记录
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/export/record/list", http.MethodGet, v1.GetKnowledgeExportRecordList, "获取知识库导出记录列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/export/record", http.MethodDelete, v1.DeleteKnowledgeExportRecord, "删除知识库库导出记录", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// 外部知识库API
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external/api/select", http.MethodGet, v1.GetKnowledgeExternalAPIList, "获取外部知识库API列表")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external/api", http.MethodPost, v1.CreateKnowledgeExternalAPI, "新建外部知识库API")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external/api", http.MethodPut, v1.UpdateKnowledgeExternalAPI, "编辑外部知识库API")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external/api", http.MethodDelete, v1.DeleteKnowledgeExternalAPI, "删除外部知识库API")
	//外部知识库
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external/select", http.MethodGet, v1.GetKnowledgeExternalList, "获取外部知识库列表")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external", http.MethodPost, v1.CreateKnowledgeExternal, "新建外部知识库")
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external", http.MethodPut, v1.UpdateKnowledgeExternal, "编辑外部知识库", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/external", http.MethodDelete, v1.DeleteKnowledgeExternal, "删除外部知识库", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// 根据知识库名称和文档名称获取文档信息
	mid.Sub("resource.knowledge").Reg(apiV1, "/knowledge/doc/by/name", http.MethodGet, v1.GetDocByKnowledgeNameAndDocName, "根据知识库名称和文档名称获取文档信息", middleware.AuthKnowledgeRagName("knowledgeName", middleware.KnowledgeView))

}
