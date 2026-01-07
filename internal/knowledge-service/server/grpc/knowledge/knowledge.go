package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	db2 "github.com/UnicomAI/wanwu/pkg/db"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	rag_service "github.com/UnicomAI/wanwu/internal/knowledge-service/service"
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

func (s *Service) SelectKnowledgeList(ctx context.Context, req *knowledgebase_service.KnowledgeSelectReq) (*knowledgebase_service.KnowledgeSelectListResp, error) {
	list, permissionMap, err := orm.SelectKnowledgeList(ctx, req.UserId, req.OrgId, req.Name, int(req.Category), req.TagIdList)
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

func (s *Service) SelectKnowledgeListByIdList(ctx context.Context, req *knowledgebase_service.BatchKnowledgeSelectReq) (*knowledgebase_service.KnowledgeSelectListResp, error) {
	list, permissionMap, err := orm.SelectKnowledgeByIdList(ctx, req.KnowledgeIdList, req.UserId, req.OrgId)
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
	err = orm.UpdateKnowledge(ctx, req.Name, req.Description, knowledge)
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
	if len(req.KnowledgeList) == 0 || req.Question == "" || req.KnowledgeMatchParams == nil {
		return nil, util.ErrCode(errs.Code_KnowledgeInvalidArguments)
	}
	var knowledgeIdList []string
	for _, k := range req.KnowledgeList {
		knowledgeIdList = append(knowledgeIdList, k.KnowledgeId)
	}
	list, _, err := orm.SelectKnowledgeByIdList(ctx, knowledgeIdList, "", "")
	if err != nil {
		return nil, err
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
		return nil, util.ErrCode(errs.Code_KnowledgeBaseHitFailed)
	}
	hitResp, err := rag_service.RagKnowledgeHit(ctx, ragHitParams)
	if err != nil {
		log.Errorf("RagKnowledgeHit error %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseHitFailed)
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
		if info.GraphStatus == model.GraphProcessing {
			processCount++
		} else if info.GraphStatus == model.GraphSuccess {
			successCount++
		} else if info.GraphStatus == model.GraphChunkFail || info.GraphStatus == model.GraphExtractFail || info.GraphStatus == model.GraphStoreFail {
			failCount++
		}
	}
	resp, err := rag_service.RagKnowledgeGraph(ctx, &rag_service.RagKnowledgeGraphParams{
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

// buildExportRecordListResp 构造问答库导出记录列表
func buildExportRecordListResp(knowledge *model.KnowledgeBase, list []*model.KnowledgeExportTask, total int64, pageSize int32, pageNum int32) *knowledgebase_service.GetExportRecordListResp {
	var retList = make([]*knowledgebase_service.ExportRecordInfo, 0)
	if len(list) > 0 {
		for _, item := range list {
			retList = append(retList, &knowledgebase_service.ExportRecordInfo{
				ExportRecordId: item.ExportId,
				Status:         int32(item.Status),
				ErrorMsg:       string(item.ErrorMsg),
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
			// 同一key优先级：删除 > 更新 > 新增
			if meta.Option == MetaOperationDelete {
				keyMap[meta.MetaInfo.Key] = meta
			} else if meta.Option == MetaOperationUpdate {
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

func buildRagHitParams(req *knowledgebase_service.KnowledgeHitReq, list []*model.KnowledgeBase, knowledgeIDToName map[string]string) (*rag_service.KnowledgeHitParams, error) {
	matchParams := req.KnowledgeMatchParams
	priorityMatch := matchParams.PriorityMatch
	filterEnable, metaParams, err := buildRagHitMetaParams(req, knowledgeIDToName)
	if err != nil {
		return nil, err
	}
	idList, nameList := buildKnowledgeList(list)
	// bff做了代理，直接传请求里的userId
	ret := &rag_service.KnowledgeHitParams{
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
	}
	return ret, nil
}

func buildRagHitMetaParams(req *knowledgebase_service.KnowledgeHitReq, knowledgeIDToName map[string]string) (bool, []*rag_service.MetadataFilterItem, error) {
	filterEnable := false // 标记是否有启用的元数据过滤
	var metaFilterConditions []*rag_service.MetadataFilterItem
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
		metaFilterConditions = append(metaFilterConditions, &rag_service.MetadataFilterItem{
			FilterKnowledgeName: knowledgeIDToName[k.KnowledgeId],
			LogicalOperator:     filterParams.FilterLogicType,
			Conditions:          metaItems,
		})
	}
	return filterEnable, metaFilterConditions, nil
}

// 构建元数据项列表
func buildRagHitMetaItems(knowledgeID string, params []*knowledgebase_service.MetaFilterParams) ([]*rag_service.MetaItem, error) {
	var metaItems []*rag_service.MetaItem
	for _, param := range params {
		// 基础参数校验
		if err := validateMetaFilterParam(knowledgeID, param); err != nil {
			return nil, err
		}
		// 转换参数值
		ragValue, err := buildValueData(param.Type, param.Value)
		if err != nil {
			log.Errorf("kbId: %s, convert value failed: %v", knowledgeID, err)
			return nil, fmt.Errorf("convert value for key %s: %s", param.Key, err.Error())
		}
		metaItems = append(metaItems, &rag_service.MetaItem{
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

func buildValueData(valueType string, value string) (interface{}, error) {
	switch valueType {
	case model.MetaTypeNumber:
	case model.MetaTypeTime:
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
	graph := orm.BuildKnowledgeGraph(string(knowledge.KnowledgeGraph))
	docCount := knowledge.DocCount
	if docCount < 0 {
		docCount = 0
	}
	return &knowledgebase_service.KnowledgeInfo{
		KnowledgeId:        knowledge.KnowledgeId,
		Name:               knowledge.Name,
		Description:        knowledge.Description,
		DocCount:           int32(docCount),
		ShareCount:         int32(knowledge.ShareCount),
		EmbeddingModelInfo: embeddingModelInfo,
		CreatedAt:          wanwu_util.Time2Str(knowledge.CreatedAt),
		CreateOrgId:        knowledge.OrgId,
		CreateUserId:       knowledge.UserId,
		RagName:            knowledge.RagName,
		GraphSwitch:        int32(knowledge.KnowledgeGraphSwitch),
		Category:           int32(knowledge.Category),
		LlmModelId:         graph.GraphModelId,
		UpdatedAt:          wanwu_util.Time2Str(knowledge.UpdatedAt),
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
		EmbeddingModel:       db2.LongText(embeddingModelInfo),
		KnowledgeGraph:       db2.LongText(knowledgeGraph),
		KnowledgeGraphSwitch: buildKnowledgeGraphSwitch(req.KnowledgeGraph.Switch),
		CreatedAt:            time.Now().UnixMilli(),
		UpdatedAt:            time.Now().UnixMilli(),
		Category:             int(req.Category),
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
func buildKnowledgeList(knowledgeList []*model.KnowledgeBase) (knowledgeIdList []string, knowledgeNameList []string) {
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
func buildKnowledgeBaseHitResp(ragKnowledgeHitResp *rag_service.RagKnowledgeHitResp) *knowledgebase_service.KnowledgeHitResp {
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
func buildWeight(priorityType int32, semanticsPriority float32, keywordPriority float32) *rag_service.WeightParams {
	if priorityType != 1 {
		return nil
	}
	return &rag_service.WeightParams{
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
			copyFile, _, _, err := rag_service.CopyFile(context.Background(), knowledgeGraph.SchemaUrl, "", false)
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
