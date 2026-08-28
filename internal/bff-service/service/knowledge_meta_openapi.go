package service

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_doc_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-doc-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// metaDocIdMaxNum 与 knowledge-service 按 docIdList 查文档的单页上限一致
const metaDocIdMaxNum = 10000

// GetKnowledgeDocMetaValueListOpenapi openapi 查询文档元数据值列表
func GetKnowledgeDocMetaValueListOpenapi(ctx *gin.Context, userId, orgId string, r *request.KnowledgeMetaValueListReq) (*response.KnowledgeMetaValueListResp, error) {
	if err := checkDocsBelongToKnowledge(ctx, userId, orgId, r.KnowledgeId, r.DocIdList); err != nil {
		return nil, err
	}
	return GetKnowledgeMetaValueList(ctx, userId, orgId, r)
}

// UpdateKnowledgeMetaKeyOpenapi openapi 定义知识库元数据key
//
// v1 的 /knowledge/doc/meta 靠 docId 是否为空分流「改文档元数据值」和「定义知识库元数据key」，
// openapi 侧拆成两个端点，这里显式走后者（不传 docId）。
func UpdateKnowledgeMetaKeyOpenapi(ctx *gin.Context, userId, orgId string, r *request.KnowledgeMetaKeyOpenapiReq) error {
	return UpdateDocMetaData(ctx, userId, orgId, &request.DocMetaDataReq{
		KnowledgeId: r.KnowledgeId,
		MetaDataList: lo.Map(r.MetaKeyList, func(item *request.KnowledgeMetaKeyItem, _ int) *request.DocMetaData {
			return &request.DocMetaData{
				MetaId:        item.MetaId,
				MetaKey:       item.MetaKey,
				MetaValueType: item.MetaValueType,
				Option:        item.Option,
			}
		}),
	})
}

// UpdateKnowledgeDocMetaValueOpenapi openapi 更新文档元数据值
func UpdateKnowledgeDocMetaValueOpenapi(ctx *gin.Context, userId, orgId string, r *request.KnowledgeDocMetaValueOpenapiReq) error {
	if err := checkDocsBelongToKnowledge(ctx, userId, orgId, r.KnowledgeId, r.DocIdList); err != nil {
		return err
	}
	return UpdateKnowledgeMetaValue(ctx, userId, orgId, &request.UpdateMetaValueReq{
		KnowledgeId: r.KnowledgeId,
		DocIdList:   r.DocIdList,
		MetaValueList: lo.Map(r.MetaValueList, func(item *request.KnowledgeMetaValueItem, _ int) *request.DocMetaData {
			return &request.DocMetaData{
				MetaKey:       item.MetaKey,
				MetaValue:     item.MetaValue,
				MetaValueType: item.MetaValueType,
				Option:        item.Option,
			}
		}),
	})
}

// checkDocsBelongToKnowledge 校验 docIdList 是否都属于该知识库
//
// 元数据值的读写在 knowledge-service 侧只按 docId 定位，请求里的 knowledgeId 不参与过滤，
// 而鉴权中间件只认 knowledgeId，openapi 侧补这道校验避免拿自己的知识库读写别人文档的元数据。
func checkDocsBelongToKnowledge(ctx *gin.Context, userId, orgId, knowledgeId string, docIdList []string) error {
	// knowledge-service 按 docIdList 查文档时单页上限 10000，超过则无法完整校验归属
	if len(docIdList) > metaDocIdMaxNum {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "docIdList size can not exceed 10000")
	}
	resp, err := knowledgeBaseDoc.GetDocList(ctx.Request.Context(), &knowledgebase_doc_service.GetDocListReq{
		KnowledgeId: knowledgeId,
		UserId:      userId,
		OrgId:       orgId,
		DocIdList:   docIdList,
	})
	if err != nil {
		return err
	}
	docIdSet := make(map[string]struct{}, len(resp.Docs))
	for _, doc := range resp.Docs {
		docIdSet[doc.DocId] = struct{}{}
	}
	for _, docId := range docIdList {
		if _, ok := docIdSet[docId]; !ok {
			return grpc_util.ErrorStatus(err_code.Code_KnowledgeDocNotExist)
		}
	}
	return nil
}
