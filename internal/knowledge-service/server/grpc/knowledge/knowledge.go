package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	knowledge_service "github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	MetaValueTypeNumber   = "number"
	MetaValueTypeTime     = "time"
	MetaConditionEmpty    = "empty"
	MetaConditionNotEmpty = "not empty"
	MetaOperationAdd      = "add"
	MetaOperationUpdate   = "update"
	MetaOperationDelete   = "delete"
)

type ExternalKnowledgeInfo struct {
	Provider              string `json:"provider"`
	ExternalAPIId         string `json:"externalApiId"`
	ExternalAPIName       string `json:"externalApiName"`
	ExternalKnowledgeId   string `json:"externalKnowledgeId"`
	ExternalKnowledgeName string `json:"externalKnowledgeName"`
}

func (s *Service) SelectKnowledgeList(ctx context.Context, req *knowledgebase_service.KnowledgeSelectReq) (*knowledgebase_service.KnowledgeSelectListResp, error) {
	list, permissionMap, err := orm.SelectKnowledgeList(ctx, req.UserId, req.OrgId, req.Name, buildCategoryList(req.Category), int(req.External), req.TagIdList)
	if err != nil {
		log.Errorf(fmt.Sprintf("获取知识库列表失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeBaseSelectFailed)
	}
	var tagMap = make(map[string][]*orm.TagRelationDetail)
	var knowledgeIdList []string
	if len(list) > 0 {
		for _, k := range list {
			knowledgeIdList = append(knowledgeIdList, k.KnowledgeId)
		}
		relation := orm.SelectKnowledgeTagListWithRelation(ctx, req.UserId, req.OrgId, "", knowledgeIdList)
		tagMap = buildKnowledgeTagMap(relation)
	}
	return buildKnowledgeListResp(list, tagMap, permissionMap), nil
}

func (s *Service) SelectKnowledgeListByUserList(ctx context.Context, req *knowledgebase_service.KnowledgeSelectUserListReq) (*knowledgebase_service.KnowledgeSelectListResp, error) {
	list, permissionMap, err := orm.SelectKnowledgeListByOrgList(ctx, req.UserId, req.OrgId, req.Name, buildCategoryList(req.Category), int(req.External))
	if err != nil {
		log.Errorf(fmt.Sprintf("获取知识库列表失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeBaseSelectFailed)
	}
	return buildKnowledgeListResp(list, nil, permissionMap), nil
}

func (s *Service) SelectKnowledgeListByIdList(ctx context.Context, req *knowledgebase_service.BatchKnowledgeSelectReq) (*knowledgebase_service.KnowledgeSelectListResp, error) {
	list, permissionMap, err := orm.SelectKnowledgeByIdListPermission(ctx, req.KnowledgeIdList, req.UserId, req.OrgId, req.NoPermission)
	if err != nil {
		log.Errorf(fmt.Sprintf("获取知识库列表失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeBaseSelectFailed)
	}
	return buildKnowledgeListResp(list, nil, permissionMap), nil
}

func (s *Service) SelectKnowledgeDetailById(ctx context.Context, req *knowledgebase_service.KnowledgeDetailSelectReq) (*knowledgebase_service.KnowledgeInfo, error) {
	knowledgeInfo, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, req.UserId, req.OrgId)
	if err != nil {
		log.Errorf(fmt.Sprintf("获取知识库详情(%v)  参数(%v)", err, req))
		return nil, err
	}
	return buildKnowledgeInfo(knowledgeInfo), nil
}

func (s *Service) SelectKnowledgeDetailByName(ctx context.Context, req *knowledgebase_service.KnowledgeDetailSelectReq) (*knowledgebase_service.KnowledgeInfo, error) {
	knowledgeInfo, err := orm.SelectKnowledgeByName(ctx, req.KnowledgeName, req.UserId, req.OrgId)
	if err != nil {
		log.Errorf(fmt.Sprintf("根据名称获取知识库详情失败(%v)  参数(%v)", err, req))
		return nil, err
	}
	return buildKnowledgeInfo(knowledgeInfo), nil
}

// SelectKnowledgeIdByRagName 根据ragName获取知识库ID
func (s *Service) SelectKnowledgeIdByRagName(ctx context.Context, req *knowledgebase_service.SelectKnowledgeIdByRagNameReq) (*knowledgebase_service.SelectKnowledgeIdByRagNameResp, error) {
	knowledgeInfo, err := orm.SelectKnowledgeByRagName(ctx, req.RagName, req.UserId, req.OrgId)
	if err != nil {
		log.Errorf(fmt.Sprintf("根据ragName获取知识库ID失败(%v)  参数(%v)", err, req))
		return nil, err
	}
	return &knowledgebase_service.SelectKnowledgeIdByRagNameResp{
		KnowledgeId: knowledgeInfo.KnowledgeId,
	}, nil
}

func (s *Service) SelectKnowledgeDetailByIdList(ctx context.Context, req *knowledgebase_service.KnowledgeDetailSelectListReq) (*knowledgebase_service.KnowledgeDetailSelectListResp, error) {
	if len(req.KnowledgeIds) == 0 {
		return buildKnowledgeInfoList([]*model.KnowledgeBase{}), nil
	}
	knowledgeInfoList, _, err := orm.SelectKnowledgeByIdList(ctx, req.KnowledgeIds, req.UserId, req.OrgId)
	if err != nil {
		log.Errorf(fmt.Sprintf("根据id列表获取知识库详情列表失败(%v)  参数(%v)", err, req))
		return nil, err
	}
	return buildKnowledgeInfoList(knowledgeInfoList), nil
}

func (s *Service) CreateKnowledge(ctx context.Context, req *knowledgebase_service.CreateKnowledgeReq) (*knowledgebase_service.CreateKnowledgeResp, error) {
	//1.重名校验
	err := orm.CheckSameKnowledgeName(ctx, req.UserId, req.OrgId, req.Name, "", int(req.Category))
	if err != nil {
		return nil, err
	}
	//2.创建知识库
	knowledgeModel, err := buildKnowledgeBaseModel(req)
	if err != nil {
		log.Errorf("buildKnowledgeBaseModel error %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseCreateFailed)
	}
	err = orm.CreateKnowledge(ctx, knowledgeModel, req.EmbeddingModelInfo.ModelId, int(req.Category))
	if err != nil {
		log.Errorf("CreateKnowledge error %v params %v", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseCreateFailed)
	}
	//3.异步存储知识图谱schema
	storeKnowledgeStoreSchema(knowledgeModel.KnowledgeId, req.KnowledgeGraph)
	//4.返回结果
	return &knowledgebase_service.CreateKnowledgeResp{
		KnowledgeId: knowledgeModel.KnowledgeId,
	}, nil
}

func (s *Service) UpdateKnowledge(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeReq) (*emptypb.Empty, error) {
	//1.查询知识库详情,这里前置做了前置权限校验，所以这里不需要再次校验
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	//2.重名校验
	err = orm.CheckSameKnowledgeName(ctx, req.UserId, req.OrgId, req.Name, knowledge.KnowledgeId, knowledge.Category)
	if err != nil {
		return nil, err
	}
	//3.更新知识库
	err = orm.UpdateKnowledge(ctx, req.Name, req.Description, req.AvatarPath, knowledge)
	if err != nil {
		log.Errorf("知识库更新失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseUpdateFailed)
	}
	return &emptypb.Empty{}, nil
}

// DeleteKnowledge 删除知识库
func (s *Service) DeleteKnowledge(ctx context.Context, req *knowledgebase_service.DeleteKnowledgeReq) (*emptypb.Empty, error) {
	//1.查询知识库详情
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	//2.校验导入状态
	err = orm.SelectKnowledgeRunningImportTask(ctx, knowledge.KnowledgeId)
	if err != nil {
		return nil, err
	}
	//3.先删除知识库，异步删除资源数据
	err = orm.DeleteKnowledge(ctx, knowledge)
	if err != nil {
		log.Errorf("删除知识库失败 error %v params %v", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseDeleteFailed)
	}
	return &emptypb.Empty{}, nil
}

// KnowledgeHit 知识库命中测试
func (s *Service) KnowledgeHit(ctx context.Context, req *knowledgebase_service.KnowledgeHitReq) (*knowledgebase_service.KnowledgeHitResp, error) {
	// 1.获取知识库信息列表
	if len(req.KnowledgeList) == 0 || req.KnowledgeMatchParams == nil {
		return nil, util.ErrCode(errs.Code_KnowledgeInvalidArguments)
	}
	if len(req.DocInfoList) == 0 && req.Question == "" {
		return nil, util.ErrCode(errs.Code_KnowledgeInvalidArguments)
	}
	var knowledgeIdList []string
	for _, k := range req.KnowledgeList {
		knowledgeIdList = append(knowledgeIdList, k.KnowledgeId)
	}
	//命中测试这里需要校验权限
	list, _, err := orm.SelectKnowledgeByIdList(ctx, knowledgeIdList, req.UserId, req.OrgId)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		marshal, _ := json.Marshal(knowledgeIdList)
		log.Errorf("knowledge hit error no permission userId %s orgId %s knowledgeID %s", req.UserId, req.OrgId, string(marshal))
		return nil, util.ErrCode(errs.Code_KnowledgeBaseHitFailed)
	}
	knowledgeIDToName := make(map[string]string)
	for _, k := range list {
		if _, exists := knowledgeIDToName[k.KnowledgeId]; !exists {
			knowledgeIDToName[k.KnowledgeId] = k.RagName
		}
	}
	// 2.RAG请求
	ragHitParams, err := buildRagHitParams(req, list, knowledgeIDToName)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_KnowledgeBaseHitFailed, err.Error())
	}
	hitResp, err := knowledge_service.RagKnowledgeHit(ctx, ragHitParams)
	if err != nil {
		log.Errorf("RagKnowledgeHit error %s", err)
		return nil, grpc_util.ErrorStatus(errs.Code_KnowledgeBaseHitFailed, err.Error())
	}
	return buildKnowledgeBaseHitResp(hitResp), nil
}

func (s *Service) GetKnowledgeMetaSelect(ctx context.Context, req *knowledgebase_service.SelectKnowledgeMetaReq) (*knowledgebase_service.SelectKnowledgeMetaResp, error) {
	metaList, err := orm.SelectMetaByKnowledgeId(ctx, "", "", req.KnowledgeId)
	if err != nil {
		log.Errorf("获取知识库元数据列表失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeMetaFetchFailed)
	}
	return buildKnowledgeMetaSelectResp(metaList), nil
}

func (s *Service) GetKnowledgeMetaValueList(ctx context.Context, req *knowledgebase_service.KnowledgeMetaValueListReq) (*knowledgebase_service.KnowledgeMetaValueListResp, error) {
	metaList, err := orm.SelectMetaByDocIds(ctx, "", "", req.DocIdList)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeMetaFetchFailed)
	}
	return buildKnowledgeMetaValueListResp(metaList), nil
}

func (s *Service) UpdateKnowledgeMetaValue(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeMetaValueReq) (*emptypb.Empty, error) {
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("没有操作该知识库的权限 参数(%v)", req)
		return nil, err
	}
	switch knowledge.Category {
	case model.CategoryQA:
		return updateQAMetaValue(ctx, req)
	default:
		return updateKnowledgeMetaValue(ctx, req)
	}
}

// GetDocByKnowledgeNameAndDocName 根据知识库名称和文档名称获取文档信息
func (s *Service) GetDocByKnowledgeNameAndDocName(ctx context.Context, req *knowledgebase_service.GetDocByKnowledgeNameAndDocNameReq) (*knowledgebase_service.GetDocByKnowledgeNameAndDocNameResp, error) {
	// 1. 根据知识库名称查询知识库信息
	knowledgeInfo, err := orm.SelectKnowledgeByRagName(ctx, req.KnowledgeName, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("根据名称获取知识库详情失败(%v)  参数(%v)", err, req))
		return nil, err
	}
	// 2. 根据知识库ID和文档名称查询文档信息
	doc, err := orm.GetDocByKnowledgeIdAndDocName(ctx, "", "", knowledgeInfo.KnowledgeId, req.DocName)
	if err != nil {
		log.Errorf(fmt.Sprintf("根据知识库ID和文档名称获取文档信息失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeDocSearchFail)
	}
	// 3. 返回结果
	return &knowledgebase_service.GetDocByKnowledgeNameAndDocNameResp{
		KnowledgeId:   knowledgeInfo.KnowledgeId,
		KnowledgeName: knowledgeInfo.Name,
		DocId:         doc.DocId,
		DocType:       doc.FileType,
		DocName:       doc.Name,
	}, nil
}

func updateQAMetaValue(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeMetaValueReq) (*emptypb.Empty, error) {
	//1.查询问答对详情
	qaPairList, err := orm.SelectQAPairByQAPairIdList(ctx, req.DocIdList, "", "")
	if err != nil {
		log.Errorf("没有操作该问答库文档的权限 参数(%v)", req)
		return nil, err
	}
	qaPair := qaPairList[0]
	//2.状态校验
	if qaPair.Status != model.QAPairSuccess {
		log.Errorf("非处理完成文档无法修改元数据 状态(%d) 错误(%v) 参数(%v)", qaPair.Status, err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeDocUpdateMetaStatusFailed)
	}
	//3.查询知识库信息
	knowledge, err := orm.SelectKnowledgeById(ctx, qaPair.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("没有操作该知识库的权限 参数(%v)", req)
		return nil, err
	}
	//4.查询元数据
	docMetaList, err := orm.SelectMetaByDocIds(ctx, "", "", req.DocIdList)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeMetaFetchFailed)
	}
	//5.构造文档元数据map
	docMetaMap := buildDocMetaMap(docMetaList)
	//6.构造元数据列表
	addList, updateList, deleteList := buildMetaList(req, docMetaMap, qaPair.KnowledgeId)
	//7.更新数据库并发送rag请求
	err = orm.BatchUpdateQAMetaValue(ctx, addList, updateList, deleteList, knowledge, knowledge.UserId, req.DocIdList)
	if err != nil {
		log.Errorf("更新文档元数据失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeMetaUpdateFailed)
	}
	return nil, nil
}

func updateKnowledgeMetaValue(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeMetaValueReq) (*emptypb.Empty, error) {
	//1.查询文档详情
	docList, err := orm.SelectDocByDocIdList(ctx, req.DocIdList, "", "")
	if err != nil {
		log.Errorf("没有操作该知识库文档的权限 参数(%v)", req)
		return nil, err
	}
	doc := docList[0]
	//2.状态校验
	if util.BuildDocRespStatus(doc.Status) != model.DocSuccess {
		log.Errorf("非处理完成文档无法修改元数据 状态(%d) 错误(%v) 参数(%v)", doc.Status, err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeDocUpdateMetaStatusFailed)
	}
	//3.查询知识库信息
	knowledge, err := orm.SelectKnowledgeById(ctx, doc.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("没有操作该知识库的权限 参数(%v)", req)
		return nil, err
	}
	//4.查询元数据
	docMetaList, err := orm.SelectMetaByDocIds(ctx, "", "", req.DocIdList)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeMetaFetchFailed)
	}
	//5.构造文档元数据map
	docMetaMap := buildDocMetaMap(docMetaList)
	//6.构造元数据列表
	addList, updateList, deleteList := buildMetaList(req, docMetaMap, doc.KnowledgeId)
	//7.更新数据库并发送rag请求
	err = orm.BatchUpdateDocMetaValue(ctx, addList, updateList, deleteList, knowledge, docList, knowledge.UserId, req.DocIdList)
	if err != nil {
		log.Errorf("更新文档元数据失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeMetaUpdateFailed)
	}
	return nil, nil
}

func (s *Service) UpdateKnowledgeStatus(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeStatusReq) (*emptypb.Empty, error) {
	err := orm.UpdateKnowledgeReportStatus(ctx, req.KnowledgeId, int(req.ReportStatus))
	if err != nil {
		log.Errorf("更新知识库状态失败(%v)  参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseUpdateFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetKnowledgeGraph(ctx context.Context, req *knowledgebase_service.KnowledgeGraphReq) (*knowledgebase_service.KnowledgeGraphResp, error) {
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	docInfo, err := orm.SelectGraphStatus(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	var processCount, successCount, failCount int32
	for _, info := range docInfo {
		switch info.GraphStatus {
		case model.GraphProcessing, model.GraphSchemaSuccess, model.GraphChunkSuccess, model.GraphExtractSuccess, model.GraphStoreSuccess:
			processCount++
		case model.GraphSuccess:
			successCount++
		case model.GraphChunkFail, model.GraphExtractFail, model.GraphStoreFail, model.GraphSchemaFail:
			failCount++
		}
	}
	resp, err := knowledge_service.RagKnowledgeGraph(ctx, &knowledge_service.RagKnowledgeGraphParams{
		KnowledgeId:   knowledge.KnowledgeId,
		KnowledgeBase: knowledge.RagName,
		UserId:        knowledge.UserId,
	})
	if err != nil {
		log.Errorf("RagKnowledgeGraph error %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseGraphFailed)
	}
	schema, err := json.Marshal(resp.Data)
	if err != nil {
		log.Errorf("RagKnowledgeGraph marshal error %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseGraphFailed)
	}
	return &knowledgebase_service.KnowledgeGraphResp{
		ProcessingCount: processCount,
		SuccessCount:    successCount,
		FailedCount:     failCount,
		Total:           processCount + successCount + failCount,
		Schema:          string(schema),
	}, nil
}

func (s *Service) GetExportRecordList(ctx context.Context, req *knowledgebase_service.GetExportRecordListReq) (*knowledgebase_service.GetExportRecordListResp, error) {
	userId, orgId := req.UserId, req.OrgId
	permission, err := orm.SelectUserKnowledgePermission(ctx, req.UserId, req.OrgId, req.KnowledgeId)
	if err != nil {
		log.Errorf(fmt.Sprintf("CheckKnowledgeUserPermission 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	if permission.PermissionType == model.PermissionTypeGrant || permission.PermissionType == model.PermissionTypeSystem {
		userId, orgId = "", ""
	}
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	list, err := orm.SelectKnowledgeExportTaskByKnowledgeId(ctx, req.KnowledgeId, userId, orgId, req.PageSize, req.PageNum)
	if err != nil {
		log.Errorf("select QA export task failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeExportRecordsSelectFailed)
	}
	return buildExportRecordListResp(knowledge, list, int64(len(list)), req.PageSize, req.PageNum), nil
}

func (s *Service) DeleteExportRecord(ctx context.Context, req *knowledgebase_service.DeleteExportRecordReq) (*emptypb.Empty, error) {
	err := orm.DeleteExportTaskById(ctx, req.ExportRecordId)
	if err != nil {
		log.Errorf("delete knowledge qa export record fail: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeDeleteExportRecordFailed)
	}
	return nil, nil
}

// buildCategoryList 构造分类列表
func buildCategoryList(category int32) []int {
	switch category {
	case model.CategoryALL:
		return make([]int, 0)
	case model.CategoryQA:
		return []int{model.CategoryQA}
	default:
		return []int{model.CategoryKnowledge, model.CategoryMultimodal}
	}
}

func (s *Service) SelectKnowledgeExternalAPIList(ctx context.Context, req *knowledgebase_service.KnowledgeExternalAPIListSelectReq) (*knowledgebase_service.KnowledgeExternalAPISelectListResp, error) {
	externalAPIList, err := orm.GetKnowledgeExternalAPIList(ctx, req.UserId, req.OrgId, req.ExternalAPIIds)
	if err != nil {
		log.Errorf("get knowledge external api list err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPIListSelectFailed)
	}
	return buildKnowledgeExternalAPISelectListResp(externalAPIList), nil
}

func (s *Service) SelectKnowledgeExternalAPIInfo(ctx context.Context, req *knowledgebase_service.KnowledgeExternalAPIInfoSelectReq) (*knowledgebase_service.KnowledgeExternalAPIInfo, error) {
	externalAPIInfo, err := orm.GetKnowledgeExternalAPIInfo(ctx, "", "", req.ExternalAPIId)
	if err != nil {
		log.Errorf("get knowledge external api info err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPIInfoSelectFailed)
	}
	return buildKnowledgeExternalAPIInfoResp(externalAPIInfo), nil
}

func (s *Service) CreateKnowledgeExternalAPI(ctx context.Context, req *knowledgebase_service.CreateKnowledgeExternalAPIReq) (*knowledgebase_service.CreateKnowledgeExternalAPIResp, error) {
	externalAPIId := wanwu_util.NewID()
	externalAPI := &model.KnowledgeExternalAPI{
		ExternalAPIId: externalAPIId,
		Name:          req.Name,
		Description:   req.Description,
		BaseUrl:       req.BaseUrl,
		APIKey:        req.ApiKey,
		Provider:      model.KnowledgeExternalAPIProviderDify,
		UserId:        req.UserId,
		OrgId:         req.OrgId,
	}
	_, err := knowledge_service.DifyGetDatasets(ctx, externalAPI, &knowledge_service.DifyGetDatasetsParams{IncludeAll: true})
	if err != nil {
		log.Errorf("check dify getDatasets err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPICheckFailed)
	}
	err = orm.CreateKnowledgeExternalAPI(ctx, externalAPI)
	if err != nil {
		log.Errorf("create knowledge external api err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPICreateFailed)
	}
	return &knowledgebase_service.CreateKnowledgeExternalAPIResp{
		ExternalAPIId: externalAPIId,
	}, nil
}

func (s *Service) UpdateKnowledgeExternalAPI(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeExternalAPIReq) (*emptypb.Empty, error) {
	err := orm.UpdateKnowledgeExternalAPI(ctx, &model.KnowledgeExternalAPI{
		ExternalAPIId: req.ExternalAPIId,
		Name:          req.Name,
		Description:   req.Description,
		BaseUrl:       req.BaseUrl,
		APIKey:        req.ApiKey,
	})
	if err != nil {
		log.Errorf("update knowledge external api err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalUpdateFailed)
	}
	return nil, nil
}

func (s *Service) DeleteKnowledgeExternalAPI(ctx context.Context, req *knowledgebase_service.DeleteKnowledgeExternalAPIReq) (*emptypb.Empty, error) {
	err := orm.DeleteKnowledgeExternalAPI(ctx, req.ExternalAPIId)
	if err != nil {
		log.Errorf("delete knowledge external api err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalDeleteFailed)
	}
	return nil, nil
}

func (s *Service) SelectKnowledgeExternalList(ctx context.Context, req *knowledgebase_service.KnowledgeExternalListSelectReq) (*knowledgebase_service.KnowledgeExternalSelectListResp, error) {
	externalAPIInfo, err := orm.GetKnowledgeExternalAPIInfo(ctx, "", "", req.ExternalAPIId)
	if err != nil {
		log.Errorf("get knowledge external api info err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPIInfoSelectFailed)
	}
	resp, err := knowledge_service.DifyGetDatasets(ctx, externalAPIInfo, &knowledge_service.DifyGetDatasetsParams{IncludeAll: true})
	if err != nil {
		log.Errorf("get dify getDatasets err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalListSelectFailed)
	}
	return buildKnowledgeExternalSelectListResp(externalAPIInfo, resp), nil
}

func (s *Service) SelectKnowledgeExternalInfo(ctx context.Context, req *knowledgebase_service.KnowledgeExternalInfoSelectReq) (*knowledgebase_service.KnowledgeExternalInfo, error) {
	//1.查询知识库详情,这里前置做了前置权限校验，所以这里不需要再次校验
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	externalKnowledgeInfo := &ExternalKnowledgeInfo{}
	err = json.Unmarshal([]byte(knowledge.ExternalKnowledge), externalKnowledgeInfo)
	if err != nil {
		return nil, err
	}
	// 2.查询外部知识库API详情
	externalAPIInfo, err := orm.GetKnowledgeExternalAPIInfo(ctx, "", "", externalKnowledgeInfo.ExternalAPIId)
	if err != nil {
		log.Errorf("get knowledge external api info err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPIInfoSelectFailed)
	}
	// 3.调用dify api
	resp, err := knowledge_service.DifyGetDataset(ctx, externalAPIInfo, externalKnowledgeInfo.ExternalKnowledgeId)
	if err != nil {
		log.Errorf("get dify getDataset err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalInfoSelectFailed)
	}
	return buildKnowledgeExternalInfoResp(externalAPIInfo, resp), nil
}

func (s *Service) CreateKnowledgeExternal(ctx context.Context, req *knowledgebase_service.CreateKnowledgeExternalReq) (*knowledgebase_service.CreateKnowledgeExternalResp, error) {
	//1.重名校验
	err := orm.CheckSameKnowledgeName(ctx, req.UserId, req.OrgId, req.Name, "", model.CategoryKnowledge)
	if err != nil {
		return nil, err
	}
	// 2.查询外部知识库API详情
	externalAPIInfo, err := orm.GetKnowledgeExternalAPIInfo(ctx, "", "", req.ExternalApiId)
	if err != nil {
		log.Errorf("get knowledge external api info err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPIInfoSelectFailed)
	}
	// 3.查询外部知识库详情
	difyKnowledgeInfo, err := knowledge_service.DifyGetDataset(ctx, externalAPIInfo, req.ExternalKnowledgeId)
	if err != nil {
		log.Errorf("get dify getDataset err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalInfoSelectFailed)
	}
	// 4.创建知识库
	knowledgeModel, err := buildExternalKnowledgeBaseModel(req, externalAPIInfo, difyKnowledgeInfo)
	if err != nil {
		log.Errorf("buildExternalKnowledgeBaseModel error %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalCreateFailed)
	}
	err = orm.CreateKnowledgeExternal(ctx, knowledgeModel)
	if err != nil {
		log.Errorf("CreateExternalKnowledge error %v params %v", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalCreateFailed)
	}
	return &knowledgebase_service.CreateKnowledgeExternalResp{
		KnowledgeId: knowledgeModel.KnowledgeId,
	}, nil
}

func (s *Service) UpdateKnowledgeExternal(ctx context.Context, req *knowledgebase_service.UpdateKnowledgeExternalReq) (*emptypb.Empty, error) {
	//1.查询知识库详情,这里前置做了前置权限校验，所以这里不需要再次校验
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	//2.重名校验
	err = orm.CheckSameKnowledgeName(ctx, req.UserId, req.OrgId, req.Name, knowledge.KnowledgeId, knowledge.Category)
	if err != nil {
		return nil, err
	}
	// 3.查询外部知识库API详情
	externalAPIInfo, err := orm.GetKnowledgeExternalAPIInfo(ctx, "", "", req.ExternalApiId)
	if err != nil {
		log.Errorf("get knowledge external api info err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalAPIInfoSelectFailed)
	}
	// 4.查询外部知识库详情
	difyKnowledgeInfo, err := knowledge_service.DifyGetDataset(ctx, externalAPIInfo, req.ExternalKnowledgeId)
	if err != nil {
		log.Errorf("get  dify getDataset err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalInfoSelectFailed)
	}
	//3.更新知识库
	externalKnowledgeInfo, err := json.Marshal(ExternalKnowledgeInfo{
		Provider:              req.Provider,
		ExternalAPIId:         req.ExternalApiId,
		ExternalAPIName:       externalAPIInfo.Name,
		ExternalKnowledgeId:   req.ExternalKnowledgeId,
		ExternalKnowledgeName: difyKnowledgeInfo.Name,
	})
	if err != nil {
		return nil, err
	}
	err = orm.UpdateKnowledgeExternal(ctx, req.KnowledgeId, req.Name, req.Description, string(externalKnowledgeInfo))
	if err != nil {
		log.Errorf("update knowledge external err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalUpdateFailed)
	}
	return nil, nil
}

func (s *Service) DeleteKnowledgeExternal(ctx context.Context, req *knowledgebase_service.DeleteKnowledgeExternalReq) (*emptypb.Empty, error) {
	//1.查询知识库详情
	_, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf(fmt.Sprintf("没有操作该知识库的权限 参数(%v)", req))
		return nil, err
	}
	//3.删除知识库
	err = orm.DeleteKnowledgeExternal(ctx, req.KnowledgeId)
	if err != nil {
		log.Errorf("delete knowledge external err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeExternalDeleteFailed)
	}
	return nil, nil
}

// buildExportRecordListResp 构造问答库导出记录列表
func buildExportRecordListResp(knowledge *model.KnowledgeBase, list []*model.KnowledgeExportTask, total int64, pageSize int32, pageNum int32) *knowledgebase_service.GetExportRecordListResp {
	var retList = make([]*knowledgebase_service.ExportRecordInfo, 0)
	if len(list) > 0 {
		for _, item := range list {
			retList = append(retList, &knowledgebase_service.ExportRecordInfo{
				ExportRecordId: item.ExportId,
				Status:         int32(item.Status),
				ErrorMsg:       item.ErrorMsg,
				FilePath:       item.ExportFilePath,
				UserId:         item.UserId,
				ExportTime:     wanwu_util.Time2Str(item.CreatedAt),
				KnowledgeName:  knowledge.Name,
			})
		}
	}
	return &knowledgebase_service.GetExportRecordListResp{
		ExportRecordInfos: retList,
		Total:             total,
		PageSize:          pageSize,
		PageNum:           pageNum,
	}
}

func buildDocMetaMap(docMetaList []*model.KnowledgeDocMeta) map[string]map[string][]*model.KnowledgeDocMeta {
	docMetaMap := make(map[string]map[string][]*model.KnowledgeDocMeta)
	for _, v := range docMetaList {
		if _, exists := docMetaMap[v.DocId]; !exists {
			docMetaMap[v.DocId] = make(map[string][]*model.KnowledgeDocMeta)
		}
		if v.ValueMain != "" {
			docMetaMap[v.DocId][v.Key] = append(docMetaMap[v.DocId][v.Key], v)
		}
	}
	return docMetaMap
}

func buildMetaList(req *knowledgebase_service.UpdateKnowledgeMetaValueReq, docMetaMap map[string]map[string][]*model.KnowledgeDocMeta, knowledgeId string) (addList, updateList []*model.KnowledgeDocMeta, deleteList []string) {
	// 处理请求数据
	reqMetaList := handleReqMetaList(req.MetaList)
	for _, meta := range reqMetaList {
		switch meta.Option {
		case MetaOperationAdd:
			handleAddMeta(req, meta, docMetaMap, knowledgeId, &addList, &updateList, &deleteList)
		case MetaOperationUpdate:
			handleUpdateMeta(req, meta, docMetaMap, knowledgeId, &addList, &updateList, &deleteList)
		case MetaOperationDelete:
			handleDeleteMeta(req, meta, docMetaMap, &deleteList)
		}
	}
	return
}

func handleReqMetaList(metaList []*knowledgebase_service.MetaValueOperation) (reqMetaList []*knowledgebase_service.MetaValueOperation) {
	if len(metaList) > 100 {
		log.Infof("metaList size exceeds 100")
		metaList = metaList[:100]
	}
	keyMap := make(map[string]*knowledgebase_service.MetaValueOperation)
	for _, meta := range metaList {
		if _, exists := keyMap[meta.MetaInfo.Key]; !exists {
			keyMap[meta.MetaInfo.Key] = meta
		} else {
			switch meta.Option {
			case MetaOperationDelete:
				keyMap[meta.MetaInfo.Key] = meta
			case MetaOperationUpdate:
				if keyMap[meta.MetaInfo.Key].Option == MetaOperationAdd {
					keyMap[meta.MetaInfo.Key] = meta
				}
			}
		}
	}
	for _, meta := range keyMap {
		reqMetaList = append(reqMetaList, meta)
	}
	return
}

func handleAddMeta(req *knowledgebase_service.UpdateKnowledgeMetaValueReq, meta *knowledgebase_service.MetaValueOperation, docMetaMap map[string]map[string][]*model.KnowledgeDocMeta, knowledgeId string, addList, updateList *[]*model.KnowledgeDocMeta, deleteList *[]string) {
	for _, docId := range req.DocIdList {
		existMetaList := docMetaMap[docId][meta.MetaInfo.Key]
		if len(existMetaList) > 0 {
			existMetaList[0].ValueMain = meta.MetaInfo.Value
			*updateList = append(*updateList, existMetaList[0])
			// 删除多余的元数据,只保留更新后的
			for i := 1; i < len(existMetaList); i++ {
				*deleteList = append(*deleteList, existMetaList[i].MetaId)
			}
		} else {
			*addList = append(*addList, &model.KnowledgeDocMeta{
				MetaId:      wanwu_util.NewID(),
				DocId:       docId,
				KnowledgeId: knowledgeId,
				UserId:      req.UserId,
				OrgId:       req.OrgId,
				Key:         meta.MetaInfo.Key,
				ValueMain:   meta.MetaInfo.Value,
				ValueType:   meta.MetaInfo.Type,
			})
		}
	}
}

func handleUpdateMeta(req *knowledgebase_service.UpdateKnowledgeMetaValueReq, meta *knowledgebase_service.MetaValueOperation, docMetaMap map[string]map[string][]*model.KnowledgeDocMeta, knowledgeId string, addList, updateList *[]*model.KnowledgeDocMeta, deleteList *[]string) {
	for _, docId := range req.DocIdList {
		existMetaList := docMetaMap[docId][meta.MetaInfo.Key]
		if len(existMetaList) > 0 {
			existMetaList[0].ValueMain = meta.MetaInfo.Value
			*updateList = append(*updateList, existMetaList[0])
			// 删除多余的元数据,只保留更新后的
			for i := 1; i < len(existMetaList); i++ {
				*deleteList = append(*deleteList, existMetaList[i].MetaId)
			}
		} else if req.ApplyToSelected {
			*addList = append(*addList, &model.KnowledgeDocMeta{
				MetaId:      wanwu_util.NewID(),
				DocId:       docId,
				KnowledgeId: knowledgeId,
				UserId:      req.UserId,
				OrgId:       req.OrgId,
				Key:         meta.MetaInfo.Key,
				ValueMain:   meta.MetaInfo.Value,
				ValueType:   meta.MetaInfo.Type,
			})
		}
	}
}

func handleDeleteMeta(req *knowledgebase_service.UpdateKnowledgeMetaValueReq, meta *knowledgebase_service.MetaValueOperation, docMetaMap map[string]map[string][]*model.KnowledgeDocMeta, deleteList *[]string) {
	for _, docId := range req.DocIdList {
		existMetaList := docMetaMap[docId][meta.MetaInfo.Key]
		for _, v := range existMetaList {
			*deleteList = append(*deleteList, v.MetaId)
		}
	}
}

func buildRagHitParams(req *knowledgebase_service.KnowledgeHitReq, list []*model.KnowledgeBase, knowledgeIDToName map[string]string) (*knowledge_service.KnowledgeHitParams, error) {
	matchParams := req.KnowledgeMatchParams
	priorityMatch := matchParams.PriorityMatch
	filterEnable, metaParams, err := buildRagHitMetaParams(req, knowledgeIDToName)
	if err != nil {
		return nil, err
	}
	idList, nameList := buildKnowledgeList(list)
	// bff做了代理，直接传请求里的userId
	ret := &knowledge_service.KnowledgeHitParams{
		UserId:               req.UserId,
		Question:             req.Question,
		KnowledgeIdList:      idList,
		KnowledgeBase:        nameList,
		TopK:                 matchParams.TopK,
		Threshold:            float64(matchParams.Score),
		RerankModelId:        buildRerankId(priorityMatch, matchParams.RerankModelId),
		RetrieveMethod:       buildRetrieveMethod(matchParams.MatchType),
		RerankMod:            buildRerankMod(priorityMatch),
		Weight:               buildWeight(priorityMatch, matchParams.SemanticsPriority, matchParams.KeywordPriority),
		TermWeight:           buildTermWeight(matchParams.TermWeight, matchParams.TermWeightEnable),
		MetaFilter:           filterEnable,
		MetaFilterConditions: metaParams,
		UseGraph:             matchParams.UseGraph,
		AttachmentList:       buildAttachmentList(req.DocInfoList),
	}
	return ret, nil
}

func buildAttachmentList(attachmentFiles []*knowledgebase_service.DocFileInfo) []*knowledge_service.AttachmentInfo {
	retList := make([]*knowledge_service.AttachmentInfo, 0)
	if len(attachmentFiles) > 0 {
		for _, attachment := range attachmentFiles {
			retList = append(retList, &knowledge_service.AttachmentInfo{
				FileType: "image",
				FileUrl:  attachment.DocUrl,
			})
		}
	}
	return retList
}

func buildRagHitMetaParams(req *knowledgebase_service.KnowledgeHitReq, knowledgeIDToName map[string]string) (bool, []*knowledge_service.MetadataFilterItem, error) {
	filterEnable := false // 标记是否有启用的元数据过滤
	var metaFilterConditions []*knowledge_service.MetadataFilterItem
	for _, k := range req.KnowledgeList {
		// 检查元数据过滤参数是否有效
		filterParams := k.MetaDataFilterParams
		if !isValidFilterParams(k.MetaDataFilterParams) {
			continue
		}
		// 校验合法值
		if k.MetaDataFilterParams.FilterLogicType == "" {
			return false, nil, errors.New("FilterLogicType is empty")
		}
		// 标记元数据过滤生效
		filterEnable = true
		// 构建元数据过滤条件
		metaItems, err := buildRagHitMetaItems(k.KnowledgeId, filterParams.MetaFilterParams)
		if err != nil {
			return false, nil, err
		}
		// 添加过滤项到结果
		metaFilterConditions = append(metaFilterConditions, &knowledge_service.MetadataFilterItem{
			FilterKnowledgeName: knowledgeIDToName[k.KnowledgeId],
			LogicalOperator:     filterParams.FilterLogicType,
			Conditions:          metaItems,
		})
	}
	return filterEnable, metaFilterConditions, nil
}

// 构建元数据项列表
func buildRagHitMetaItems(knowledgeID string, params []*knowledgebase_service.MetaFilterParams) ([]*knowledge_service.MetaItem, error) {
	var metaItems []*knowledge_service.MetaItem
	for _, param := range params {
		// 基础参数校验
		if err := validateMetaFilterParam(knowledgeID, param); err != nil {
			return nil, err
		}
		// 转换参数值
		ragValue, err := BuildValueData(param.Type, param.Value)
		if err != nil {
			log.Errorf("kbId: %s, convert value failed: %v", knowledgeID, err)
			return nil, fmt.Errorf("convert value for key %s: %s", param.Key, err.Error())
		}
		metaItems = append(metaItems, &knowledge_service.MetaItem{
			MetaName:           param.Key,
			MetaType:           param.Type,
			ComparisonOperator: param.Condition,
			Value:              ragValue,
		})
	}
	return metaItems, nil
}

// 校验元数据过滤参数
func validateMetaFilterParam(knowledgeID string, param *knowledgebase_service.MetaFilterParams) error {
	// 检查关键参数是否为空
	if param.Key == "" || param.Type == "" || param.Condition == "" {
		errMsg := "key/type/condition cannot be empty"
		log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
		return errors.New(errMsg)
	}

	// 检查空条件与值的匹配性
	if param.Condition == MetaConditionEmpty || param.Condition == MetaConditionNotEmpty {
		if param.Value != "" {
			errMsg := "condition is empty/non-empty, value should be empty"
			log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
			return errors.New(errMsg)
		}
	} else {
		if param.Value == "" {
			errMsg := "value is empty"
			log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
			return errors.New(errMsg)
		}
	}

	return nil
}

func isValidFilterParams(params *knowledgebase_service.MetaDataFilterParams) bool {
	return params != nil &&
		params.FilterEnable &&
		params.MetaFilterParams != nil &&
		len(params.MetaFilterParams) > 0
}

func BuildValueData(valueType string, value string) (interface{}, error) {
	if value == "" {
		return value, nil
	}
	switch valueType {
	case model.MetaTypeNumber, model.MetaTypeTime:
		return strconv.ParseInt(value, 10, 64)
	}
	return value, nil
}

func buildKnowledgeMetaSelectResp(metaList []*model.KnowledgeDocMeta) *knowledgebase_service.SelectKnowledgeMetaResp {
	if len(metaList) == 0 {
		return &knowledgebase_service.SelectKnowledgeMetaResp{}
	}
	var retMetaList []*knowledgebase_service.KnowledgeMetaData
	newMetaList := checkRepeatedMetaKey(metaList)
	for _, meta := range newMetaList {
		if meta.Key != "" {
			retMetaList = append(retMetaList, &knowledgebase_service.KnowledgeMetaData{
				MetaId: meta.MetaId,
				Key:    meta.Key,
				Type:   meta.ValueType,
			})
		}
	}
	return &knowledgebase_service.SelectKnowledgeMetaResp{
		MetaList: retMetaList,
	}
}

// buildKnowledgeListResp 构造知识库列表返回结果
func buildKnowledgeListResp(knowledgeList []*model.KnowledgeBase, knowledgeTagMap map[string][]*orm.TagRelationDetail, permissionMap map[string]int) *knowledgebase_service.KnowledgeSelectListResp {
	if len(knowledgeList) == 0 {
		return &knowledgebase_service.KnowledgeSelectListResp{}
	}
	var retList []*knowledgebase_service.KnowledgeInfo
	for _, knowledge := range knowledgeList {
		knowledgeInfo := buildKnowledgeInfo(knowledge)
		knowledgeInfo.KnowledgeTagInfoList = buildKnowledgeTagList(knowledge.KnowledgeId, knowledgeTagMap)
		knowledgeInfo.PermissionType = buildKnowledgePermission(knowledge.KnowledgeId, permissionMap)
		retList = append(retList, knowledgeInfo)
	}
	return &knowledgebase_service.KnowledgeSelectListResp{
		KnowledgeList: retList,
	}
}

func buildKnowledgeTagMap(tagRelation *orm.TagRelation) map[string][]*orm.TagRelationDetail {
	if tagRelation.RelationErr != nil || tagRelation.TagErr != nil {
		return make(map[string][]*orm.TagRelationDetail)
	}
	var knowledgeTagMap = make(map[string][]*orm.TagRelationDetail)
	for _, relation := range tagRelation.RelationList {
		details := knowledgeTagMap[relation.KnowledgeId]
		if details == nil {
			details = make([]*orm.TagRelationDetail, 0)
		}
		for _, tag := range tagRelation.TagList {
			if tag.TagId == relation.TagId {
				details = append(details, &orm.TagRelationDetail{
					TagId:   tag.TagId,
					TagName: tag.Name,
				})
			}
		}
		if len(details) == 0 {
			continue
		}
		knowledgeTagMap[relation.KnowledgeId] = details
	}
	return knowledgeTagMap
}

func buildKnowledgeTagList(knowledgeId string, knowledgeTagMap map[string][]*orm.TagRelationDetail) []*knowledgebase_service.KnowledgeTagInfo {
	if len(knowledgeTagMap) == 0 {
		return []*knowledgebase_service.KnowledgeTagInfo{}
	}
	tagList := knowledgeTagMap[knowledgeId]
	if len(tagList) == 0 {
		return []*knowledgebase_service.KnowledgeTagInfo{}
	}
	var retList []*knowledgebase_service.KnowledgeTagInfo
	for _, tag := range tagList {
		retList = append(retList, &knowledgebase_service.KnowledgeTagInfo{
			TagId:   tag.TagId,
			TagName: tag.TagName,
		})
	}
	return retList
}

func buildKnowledgePermission(knowledgeId string, permissionMap map[string]int) int32 {
	if len(permissionMap) == 0 {
		return 0
	}
	return int32(permissionMap[knowledgeId])
}

func checkRepeatedMetaKey(metaList []*model.KnowledgeDocMeta) []*model.KnowledgeDocMeta {
	if len(metaList) == 0 {
		return []*model.KnowledgeDocMeta{}
	}
	return lo.UniqBy(metaList, func(item *model.KnowledgeDocMeta) string {
		return item.Key
	})
}

// buildKnowledgeInfo 构造知识库信息
func buildKnowledgeInfo(knowledge *model.KnowledgeBase) *knowledgebase_service.KnowledgeInfo {
	embeddingModelInfo := &knowledgebase_service.EmbeddingModelInfo{}
	_ = json.Unmarshal([]byte(knowledge.EmbeddingModel), embeddingModelInfo)
	externalKnowledgeInfo := &knowledgebase_service.KnowledgeExternalInfo{}
	_ = json.Unmarshal([]byte(knowledge.ExternalKnowledge), externalKnowledgeInfo)
	graph := orm.BuildKnowledgeGraph(knowledge.KnowledgeGraph)
	docCount := knowledge.DocCount
	if docCount < 0 {
		docCount = 0
	}
	return &knowledgebase_service.KnowledgeInfo{
		KnowledgeId:           knowledge.KnowledgeId,
		Name:                  knowledge.Name,
		Description:           knowledge.Description,
		DocCount:              int32(docCount),
		ShareCount:            int32(knowledge.ShareCount),
		EmbeddingModelInfo:    embeddingModelInfo,
		CreatedAt:             wanwu_util.Time2Str(knowledge.CreatedAt),
		CreateOrgId:           knowledge.OrgId,
		CreateUserId:          knowledge.UserId,
		RagName:               knowledge.RagName,
		GraphSwitch:           int32(knowledge.KnowledgeGraphSwitch),
		Category:              int32(knowledge.Category),
		LlmModelId:            graph.GraphModelId,
		UpdatedAt:             wanwu_util.Time2Str(knowledge.UpdatedAt),
		External:              int32(knowledge.External),
		KnowledgeExternalInfo: externalKnowledgeInfo,
		AvatarPath:            knowledge.AvatarPath,
	}
}

// buildKnowledgeInfoList 构造知识库信息列表
func buildKnowledgeInfoList(knowledgeList []*model.KnowledgeBase) *knowledgebase_service.KnowledgeDetailSelectListResp {
	var retList []*knowledgebase_service.KnowledgeInfo
	for _, v := range knowledgeList {
		info := buildKnowledgeInfo(v)
		retList = append(retList, info)
	}
	return &knowledgebase_service.KnowledgeDetailSelectListResp{
		List:  retList,
		Total: int32(len(retList)),
	}
}

// buildKnowledgeBaseModel 构造知识库模型
func buildKnowledgeBaseModel(req *knowledgebase_service.CreateKnowledgeReq) (*model.KnowledgeBase, error) {
	embeddingModelInfo, err := json.Marshal(req.EmbeddingModelInfo)
	if err != nil {
		return nil, err
	}
	knowledgeGraph, err := json.Marshal(req.KnowledgeGraph)
	if err != nil {
		return nil, err
	}
	return &model.KnowledgeBase{
		KnowledgeId:          wanwu_util.NewID(),
		Name:                 req.Name,
		RagName:              wanwu_util.NewID(), //重新生成的 不是knowledgeID
		Description:          req.Description,
		OrgId:                req.OrgId,
		UserId:               req.UserId,
		EmbeddingModel:       string(embeddingModelInfo),
		KnowledgeGraph:       string(knowledgeGraph),
		KnowledgeGraphSwitch: buildKnowledgeGraphSwitch(req.KnowledgeGraph.Switch),
		CreatedAt:            time.Now().UnixMilli(),
		UpdatedAt:            time.Now().UnixMilli(),
		Category:             int(req.Category),
		AvatarPath:           req.AvatarPath,
	}, nil
}

// buildExternalKnowledgeBaseModel 构造外部知识库
func buildExternalKnowledgeBaseModel(req *knowledgebase_service.CreateKnowledgeExternalReq, externalAPIInfo *model.KnowledgeExternalAPI, difyKnowledgeInfo *knowledge_service.DifyDatasetData) (*model.KnowledgeBase, error) {
	externalKnowledgeInfo, err := json.Marshal(ExternalKnowledgeInfo{
		ExternalAPIId:         req.ExternalApiId,
		ExternalAPIName:       externalAPIInfo.Name,
		ExternalKnowledgeId:   req.ExternalKnowledgeId,
		ExternalKnowledgeName: difyKnowledgeInfo.Name,
		Provider:              model.KnowledgeExternalAPIProviderDify,
	})
	if err != nil {
		return nil, err
	}
	return &model.KnowledgeBase{
		KnowledgeId:       wanwu_util.NewID(),
		Name:              req.Name,
		Description:       req.Description,
		DocCount:          int(difyKnowledgeInfo.TotalDocuments),
		OrgId:             req.OrgId,
		UserId:            req.UserId,
		CreatedAt:         time.Now().UnixMilli(),
		UpdatedAt:         time.Now().UnixMilli(),
		Category:          model.CategoryKnowledge,
		External:          model.ExternalKnowledge,
		ExternalKnowledge: string(externalKnowledgeInfo),
	}, nil
}

// buildKnowledgeGraphSwitch 构造知识图谱开关
func buildKnowledgeGraphSwitch(graphSwitch bool) int {
	if graphSwitch {
		return 1
	}
	return 0
}

// buildKnowledgeList 构造知识库名称
func buildKnowledgeList(knowledgeList []*model.KnowledgeBase) (knowledgeIdList, knowledgeNameList []string) {
	if len(knowledgeList) == 0 {
		return make([]string, 0), make([]string, 0)
	}
	for _, knowledge := range knowledgeList {
		knowledgeNameList = append(knowledgeNameList, knowledge.RagName)
		knowledgeIdList = append(knowledgeIdList, knowledge.KnowledgeId)
	}
	return
}

// buildKnowledgeBaseHitResp 构造知识库命中返回
func buildKnowledgeBaseHitResp(ragKnowledgeHitResp *knowledge_service.RagKnowledgeHitResp) *knowledgebase_service.KnowledgeHitResp {
	knowledgeHitData := ragKnowledgeHitResp.Data
	var searchList = make([]*knowledgebase_service.KnowledgeSearchInfo, 0)
	list := knowledgeHitData.SearchList
	if len(list) > 0 {
		for _, search := range list {
			childContentList := make([]*knowledgebase_service.ChildContent, 0)
			for _, child := range search.ChildContentList {
				childContentList = append(childContentList, &knowledgebase_service.ChildContent{
					ChildSnippet: child.ChildSnippet,
					Score:        float32(child.Score),
				})
			}
			childScore := make([]float32, 0)
			for _, score := range search.ChildScore {
				childScore = append(childScore, float32(score))
			}
			//todo knowledgeName 替换
			searchList = append(searchList, &knowledgebase_service.KnowledgeSearchInfo{
				Title:            search.Title,
				Snippet:          search.Snippet,
				KnowledgeName:    search.KbName,
				ChildContentList: childContentList,
				ChildScore:       childScore,
				ContentType:      search.ContentType,
				Score:            float32(search.Score),
				RerankInfo:       buildRerankInfo(search.RerankInfo),
			})
		}
	}
	return &knowledgebase_service.KnowledgeHitResp{
		Prompt:     knowledgeHitData.Prompt,
		Score:      knowledgeHitData.Score,
		SearchList: searchList,
		UseGraph:   knowledgeHitData.UseGraph,
	}
}

func buildRerankInfo(rerankInfo []*knowledge_service.RerankInfo) []*knowledgebase_service.RerankInfo {
	rerankInfoList := make([]*knowledgebase_service.RerankInfo, 0)
	if len(rerankInfo) > 0 {
		for _, v := range rerankInfo {
			rerankInfoList = append(rerankInfoList, &knowledgebase_service.RerankInfo{
				FileUrl: v.FileUrl,
				Score:   float32(v.Score),
				Type:    v.Type,
			})
		}
	}
	return rerankInfoList
}

// buildRerankId 构造重排序模型id
func buildRerankId(priorityType int32, rerankId string) string {
	if priorityType == 1 {
		return ""
	}
	return rerankId
}

// buildRetrieveMethod 构造检索方式
func buildRetrieveMethod(matchType string) string {
	switch matchType {
	case "vector":
		return "semantic_search"
	case "text":
		return "full_text_search"
	case "mix":
		return "hybrid_search"
	}
	return ""
}

// buildRerankMod 构造重排序模式
func buildRerankMod(priorityType int32) string {
	if priorityType == 1 {
		return "weighted_score"
	}
	return "rerank_model"
}

// buildWeight 构造权重信息
func buildWeight(priorityType int32, semanticsPriority float32, keywordPriority float32) *knowledge_service.WeightParams {
	if priorityType != 1 {
		return nil
	}
	return &knowledge_service.WeightParams{
		VectorWeight: semanticsPriority,
		TextWeight:   keywordPriority,
	}
}

// buildTermWeight 构造关键词系数信息
func buildTermWeight(termWeight float32, termWeightEnable bool) float32 {
	if termWeightEnable {
		return termWeight
	}
	return 0.0
}
func buildKnowledgeMetaValueListResp(metaList []*model.KnowledgeDocMeta) *knowledgebase_service.KnowledgeMetaValueListResp {
	retMap := make(map[string]*knowledgebase_service.KnowledgeMetaValues)
	var retList []*knowledgebase_service.KnowledgeMetaValues
	for _, meta := range metaList {
		if meta.ValueMain == "" || meta.Key == "" || meta.ValueType == "" {
			continue
		}
		if _, exists := retMap[meta.Key]; !exists {
			retMap[meta.Key] = &knowledgebase_service.KnowledgeMetaValues{
				MetaId:    meta.MetaId,
				Key:       meta.Key,
				Type:      meta.ValueType,
				ValueList: []string{meta.ValueMain},
			}
		} else {
			retMap[meta.Key].ValueList = append(retMap[meta.Key].ValueList, meta.ValueMain)
		}
	}
	for _, retMeta := range retMap {
		retMeta.ValueList = lo.Uniq(retMeta.ValueList)
		retList = append(retList, retMeta)
	}
	return &knowledgebase_service.KnowledgeMetaValueListResp{
		MetaList: retList,
	}
}

// storeKnowledgeStoreSchema 存储知识库图谱Url
func storeKnowledgeStoreSchema(knowledgeId string, knowledgeGraph *knowledgebase_service.KnowledgeGraph) {
	if knowledgeGraph.Switch && knowledgeGraph.SchemaUrl != "" {
		go func() {
			defer wanwu_util.PrintPanicStack()
			copyFile, _, _, err := knowledge_service.CopyFile(context.Background(), knowledgeGraph.SchemaUrl, "", false)
			if err != nil {
				log.Errorf("store knowledge copy file (%v) err: %v", knowledgeGraph.SchemaUrl, err)
				return
			}
			knowledgeGraph.SchemaUrl = copyFile
			marshal, err := json.Marshal(knowledgeGraph)
			if err != nil {
				log.Errorf("store knowledge marshal err: %v", err)
				return
			}
			err = orm.UpdateKnowledgeGraph(db.GetClient().DB, knowledgeId, string(marshal))
			if err != nil {
				log.Errorf("store knowledge update err: %v", err)
				return
			}
		}()
	}
}

func buildKnowledgeExternalAPISelectListResp(externalAPIs []*model.KnowledgeExternalAPI) *knowledgebase_service.KnowledgeExternalAPISelectListResp {
	var externalAPIList []*knowledgebase_service.KnowledgeExternalAPIInfo
	for _, externalAPI := range externalAPIs {
		externalAPIList = append(externalAPIList, buildKnowledgeExternalAPIInfoResp(externalAPI))
	}
	return &knowledgebase_service.KnowledgeExternalAPISelectListResp{
		ExternalAPIList: externalAPIList,
	}
}

func buildKnowledgeExternalAPIInfoResp(externalAPI *model.KnowledgeExternalAPI) *knowledgebase_service.KnowledgeExternalAPIInfo {
	return &knowledgebase_service.KnowledgeExternalAPIInfo{
		ExternalAPIId: externalAPI.ExternalAPIId,
		Name:          externalAPI.Name,
		Description:   externalAPI.Description,
		BaseUrl:       externalAPI.BaseUrl,
		ApiKey:        externalAPI.APIKey,
	}
}

func buildKnowledgeExternalSelectListResp(externalAPI *model.KnowledgeExternalAPI, difyGetDatasetsResp *knowledge_service.DifyGetDatasetsResp) *knowledgebase_service.KnowledgeExternalSelectListResp {
	var externalKnowledgeList []*knowledgebase_service.KnowledgeExternalInfo
	for _, dataset := range difyGetDatasetsResp.Data {
		if dataset.ExternalKnowledgeInfo != nil && dataset.ExternalKnowledgeInfo.ExternalKnowledgeId != "" {
			continue
		}
		externalKnowledgeList = append(externalKnowledgeList, buildKnowledgeExternalInfoResp(externalAPI, dataset))
	}
	return &knowledgebase_service.KnowledgeExternalSelectListResp{
		ExternalKnowledgeList: externalKnowledgeList,
	}
}

func buildKnowledgeExternalInfoResp(externalAPI *model.KnowledgeExternalAPI, difyDatasetData *knowledge_service.DifyDatasetData) *knowledgebase_service.KnowledgeExternalInfo {
	return &knowledgebase_service.KnowledgeExternalInfo{
		ExternalKnowledgeId:   difyDatasetData.Id,
		ExternalKnowledgeName: difyDatasetData.Name,
		ExternalAPIId:         externalAPI.ExternalAPIId,
		ExternalAPIName:       externalAPI.Name,
		ExternalAPIUrl:        externalAPI.BaseUrl,
		ExternalAPIKey:        externalAPI.APIKey,
		DocCount:              difyDatasetData.TotalDocuments,
		RetrievalModelInfo: &knowledgebase_service.RetrievalModelInfo{
			SearchMethod:    difyDatasetData.RetrievalModelDict.SearchMethod,
			RerankingEnable: difyDatasetData.RetrievalModelDict.RerankingEnable,
			RerankingMode:   difyDatasetData.RetrievalModelDict.RerankingMode,
			RerankingModel: &knowledgebase_service.RerankingModel{
				RerankingModelName:    difyDatasetData.RetrievalModelDict.RerankingModel.RerankingModelName,
				RerankingProviderName: difyDatasetData.RetrievalModelDict.RerankingModel.RerankingProviderName,
			},
			Weights:               buildDifyWeights(difyDatasetData.RetrievalModelDict.Weights),
			TopK:                  difyDatasetData.RetrievalModelDict.TopK,
			ScoreThresholdEnabled: difyDatasetData.RetrievalModelDict.ScoreThresholdEnabled,
			ScoreThreshold:        difyDatasetData.RetrievalModelDict.ScoreThreshold,
		},
	}
}

func buildDifyWeights(weights *knowledge_service.DifyWeights) *knowledgebase_service.Weights {
	if weights == nil {
		return nil
	}
	return &knowledgebase_service.Weights{
		WeightType: weights.WeightType,
		KeywordSetting: &knowledgebase_service.KeywordSetting{
			KeywordWeight: weights.KeywordSetting.KeywordWeight,
		},
		VectorSetting: &knowledgebase_service.VectorSetting{
			VectorWeight:          weights.VectorSetting.VectorWeight,
			EmbeddingModelName:    weights.VectorSetting.EmbeddingModelName,
			EmbeddingProviderName: weights.VectorSetting.EmbeddingProviderName,
		},
	}
}
