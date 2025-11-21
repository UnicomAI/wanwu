package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerKnowledge(apiV1 *gin.RouterGroup) {
	// Knowledge base addition, deletion, modification and query
	mid.Sub("knowledge").Reg(apiV1, "/knowledge", http.MethodPost, v1.CreateKnowledge, "创建知识库（文档分类）")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge", http.MethodPut, v1.UpdateKnowledge, "修改知识库（文档分类）", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge", http.MethodDelete, v1.DeleteKnowledge, "删除知识库（文档分类）", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))

	// Knowledge base hit test, general verification is not easy to change internal verification
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/hit", http.MethodPost, v1.KnowledgeHit, "知识库命中测试")

	// Knowledge base documentation
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/list", http.MethodGet, v1.GetDocList, "获取文档列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/import", http.MethodPost, v1.ImportDoc, "上传文档", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/import/tip", http.MethodGet, v1.GetDocImportTip, "获取知识库文档上传状态", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	// Knowledge base document, the following general verification is not easy to change the internal verification
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc", http.MethodDelete, v1.DeleteDoc, "删除文档", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/meta", http.MethodPost, v1.UpdateDocMetaData, "更新文档元数据", middleware.AuthKnowledgeIfHas("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/meta/batch", http.MethodPost, v1.BatchUpdateDocMetaData, "批量更新文档元数据", middleware.AuthKnowledgeIfHas("knowledgeId", middleware.KnowledgeEdit))

	// Knowledge base metadata, add knowledgeId to the front end
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/meta/select", http.MethodGet, v1.GetKnowledgeMetaKeySelect, "获取知识库元数据key列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/meta/value/list", http.MethodPost, v1.GetKnowledgeMetaValueList, "获取知识库元数据值列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/meta/value/update", http.MethodPost, v1.UpdateKnowledgeMetaValue, "更新知识库元数据值列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// Knowledge base document slice addition, deletion, modification and query
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/list", http.MethodGet, v1.GetDocSegmentList, "获取文档切分结果", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/status/update", http.MethodPost, v1.UpdateDocSegmentStatus, "更新文档切片启用状态", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/labels", http.MethodPost, v1.UpdateDocSegmentLabels, "更新文档切片标签", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/create", http.MethodPost, v1.CreateDocSegment, "新增文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/batch/create", http.MethodPost, v1.BatchCreateDocSegment, "批量新增文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/delete", http.MethodDelete, v1.DeleteDocSegment, "删除文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/update", http.MethodPost, v1.UpdateDocSegment, "更新文档切片", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/child/list", http.MethodGet, v1.GetDocChildSegmentList, "获取子分段列表", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/child/create", http.MethodPost, v1.CreateDocChildSegment, "创建子分段", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/child/update", http.MethodPost, v1.UpdateDocChildSegment, "更新子分段", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/segment/child/delete", http.MethodDelete, v1.DeleteDocChildSegment, "删除子分段", middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))

	// Knowledge base URL document import, this interface does not require verification
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/doc/url/analysis", http.MethodPost, v1.AnalysisDocUrl, "解析url")

	// Knowledge base tag addition, deletion, modification and query
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/tag", http.MethodGet, v1.GetKnowledgeTagSelect, "查询知识库标签列表", middleware.AuthKnowledgeIfHas("knowledgeId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/tag", http.MethodPost, v1.CreateKnowledgeTag, "创建知识库标签")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/tag", http.MethodPut, v1.UpdateKnowledgeTag, "修改知识库标签")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/tag", http.MethodDelete, v1.DeleteKnowledgeTag, "删除知识库标签")
	// Bind knowledge base tags
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/tag/bind/count", http.MethodGet, v1.SelectKnowledgeTagBindCount, "查询标签绑定的知识库数量")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/tag/bind", http.MethodPost, v1.BindKnowledgeTag, "绑定知识库标签", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// Knowledge base keyword management--the bottom layer has been filtered by knowledgeId, so there is no need to process it here.
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodGet, v1.GetKnowledgeKeywordsList, "查询知识库关键词列表")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodPost, v1.CreateKnowledgeKeywords, "新增知识库关键词")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/keywords/detail", http.MethodGet, v1.GetKnowledgeKeywordsDetail, "查询知识库关键词详情")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodPut, v1.UpdateKnowledgeKeywords, "编辑知识库关键词")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/keywords", http.MethodDelete, v1.DeleteDocCategoryKeywords, "删除知识库关键词")

	// Knowledge base separator addition, deletion, modification, and query are related to the user and have nothing to do with the knowledge base. No verification of permissions is required.
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodGet, v1.GetKnowledgeSplitterSelect, "查询知识库分隔符列表")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodPost, v1.CreateKnowledgeSplitter, "创建知识库分隔符")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodPut, v1.UpdateKnowledgeSplitter, "修改知识库分隔符")
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/splitter", http.MethodDelete, v1.DeleteKnowledgeSplitter, "删除知识库分隔符")

	// Knowledge base permissions
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/org", http.MethodGet, v1.SelectKnowledgeOrg, "知识库组织权限列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/user", http.MethodGet, v1.SelectKnowledgeUserPermit, "知识库用户权限列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/user/no/permit", http.MethodGet, v1.SelectKnowledgeUserNoPermit, "没有知识库用户列表", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/user/add", http.MethodPost, v1.AddKnowledgeUser, "增加知识库用户", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/user/edit", http.MethodPost, v1.EditKnowledgeUser, "修改知识库用户", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/user/delete", http.MethodDelete, v1.DeleteKnowledgeUser, "删除知识库用户", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeGrant))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/user/admin/transfer", http.MethodPost, v1.TransferKnowledgeUserAdmin, "转让管理员权限", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))

	// Knowledge Base Community Report
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/report/list", http.MethodGet, v1.GetKnowledgeReport, "获取社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/report/generate", http.MethodPost, v1.GenerateKnowledgeReport, "生成社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/report/delete", http.MethodDelete, v1.DeleteKnowledgeReport, "删除社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/report/update", http.MethodPost, v1.UpdateKnowledgeReport, "更新社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/report/add", http.MethodPost, v1.AddKnowledgeReport, "单条新增社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/report/batch/add", http.MethodPost, v1.BatchAddKnowledgeReport, "批量新增社区报告", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// Knowledge graph details
	mid.Sub("knowledge").Reg(apiV1, "/knowledge/graph", http.MethodGet, v1.GetKnowledgeGraph, "获取知识图谱详情", middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))

}
