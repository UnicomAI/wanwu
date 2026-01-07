package knowledge_qa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_qa_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-qa-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	util2 "github.com/UnicomAI/wanwu/pkg/util"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) ImportQAPair(ctx context.Context, req *knowledgebase_qa_service.ImportQAPairReq) (*emptypb.Empty, error) {
	task, err := buildQAPairImportTask(req)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairImportFailed)
	}
	//创建导入任务
	err = orm.CreateKnowledgeQAPairImportTask(ctx, task)
	if err != nil {
		log.Errorf("import qa pairs fail %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairImportFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetQAImportTip(ctx context.Context, req *knowledgebase_qa_service.QAImportTipReq) (*knowledgebase_qa_service.QAImportTipResp, error) {
	//1.查询知识库详情,前置参数校验
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("select QA knowledge failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQABaseSelectFailed)
	}
	//2.查询第一个异步任务信息
	taskList, err := orm.SelectKnowledgeQALatestImportTask(ctx, req.KnowledgeId)
	if err != nil {
		log.Errorf("select QA import task info failed err: (%v)", err)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairImportTaskSelectFailed)
	}
	if len(taskList) == 0 {
		return &knowledgebase_qa_service.QAImportTipResp{
			KnowledgeId:   req.KnowledgeId,
			KnowledgeName: knowledge.Name,
			UploadStatus:  model.KnowledgeQAPairImportSuccess,
		}, nil
	}
	if len(taskList) > 0 {
		task := taskList[0]
		if task.Status == model.KnowledgeQAPairImportFail {
			return &knowledgebase_qa_service.QAImportTipResp{
				KnowledgeId:   req.KnowledgeId,
				KnowledgeName: knowledge.Name,
				Message:       string("\n" + task.ErrorMsg),
				UploadStatus:  model.KnowledgeQAPairImportFail,
			}, nil
		} else if task.Status == model.KnowledgeQAPairImportSuccess {
			return &knowledgebase_qa_service.QAImportTipResp{
				KnowledgeId:   req.KnowledgeId,
				KnowledgeName: knowledge.Name,
				UploadStatus:  model.KnowledgeQAPairImportSuccess,
			}, nil
		}
	}
	return &knowledgebase_qa_service.QAImportTipResp{
		KnowledgeId:   req.KnowledgeId,
		KnowledgeName: knowledge.Name,
		Message:       "",
		UploadStatus:  model.KnowledgeQAPairImportImporting,
	}, nil
}

func (s *Service) ExportQAPair(ctx context.Context, req *knowledgebase_qa_service.ExportQAPairReq) (*emptypb.Empty, error) {
	task, err := buildQAPairExportTask(req)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairExportFailed)
	}
	//创建导出任务
	err = orm.CreateKnowledgeQAPairExportTask(ctx, task)
	if err != nil {
		log.Errorf("export qa pairs fail %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairExportFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetQAPairList(ctx context.Context, req *knowledgebase_qa_service.GetQAPairListReq) (*knowledgebase_qa_service.GetQAPairListResp, error) {
	//查询知识库信息
	knowledge, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("select QA knowledge failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQABaseSelectFailed)
	}
	qaPairIdList := make([]string, 0)
	//查找元数据值所对应的文档列表
	if req.MetaValue != "" {
		qaPairIdList, err = orm.SelectDocIdListByMetaValue(ctx, "", "", req.KnowledgeId, req.MetaValue)
		if err != nil {
			log.Errorf("获取知识库元数据失败(%v)  参数(%v)", err, req)
			return nil, util.ErrCode(errs.Code_KnowledgeMetaFetchFailed)
		}
		if len(qaPairIdList) == 0 {
			return buildQAPairListResp(nil, knowledge, nil, 0, req.PageSize, req.PageNum), nil
		}
	}
	list, total, err := orm.GetQAPairList(ctx, "", "", req.KnowledgeId,
		req.Name, int(req.Status), qaPairIdList, req.PageSize, req.PageNum)
	if err != nil {
		log.Errorf("select QA pairs failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairsSelectFailed)
	}
	var qaIds []string
	for _, item := range list {
		qaIds = append(qaIds, item.QAPairId)
	}
	// 查询元数据
	docMetaList, err := orm.SelectMetaByDocIds(ctx, "", "", qaIds)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeMetaFetchFailed)
	}
	return buildQAPairListResp(list, knowledge, docMetaList, total, req.PageSize, req.PageNum), nil
}

func (s *Service) GetQAPairInfo(ctx context.Context, req *knowledgebase_qa_service.GetQAPairInfoReq) (*knowledgebase_qa_service.QAPairInfo, error) {
	qaPairInfo, err := orm.GetQAPairInfoById(ctx, req.QaPairId, req.UserId, req.OrgId)
	if err != nil {
		log.Errorf("select QA pair failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairSelectFailed)
	}
	return buildQAPairInfo(qaPairInfo), nil
}

func (s *Service) CreateQAPair(ctx context.Context, req *knowledgebase_qa_service.CreateQAPairReq) (*knowledgebase_qa_service.CreateQAPairResp, error) {
	//1.查询问答库详情
	knowledgeBase, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("select QA knowledge failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQABaseSelectFailed)
	}
	//2.检查问题MD5
	question := strings.Trim(req.Question, " ")
	answer := strings.Trim(req.Answer, " ")
	questionMD5 := util.MD5(question)
	err = orm.CheckKnowledgeQAPairQuestion(ctx, "", req.KnowledgeId, questionMD5)
	if err != nil {
		log.Errorf("check qa pair question md5 fail %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeDuplicateQAPirQuestion)
	}
	// 3.新建问答对
	qaPairId := wanwu_util.NewID()
	qaPairs, ragParams := buildCreateQAPairParams(knowledgeBase, question, answer, questionMD5, qaPairId, req.UserId, req.OrgId)
	err = orm.CreateKnowledgeQAPairAndCount(ctx, req.KnowledgeId, qaPairs, ragParams)
	if err != nil {
		log.Errorf("create qa pair fail: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeCreateQAPairFailed)
	}
	return &knowledgebase_qa_service.CreateQAPairResp{
		QaPairId: qaPairId,
	}, nil
}

func (s *Service) UpdateQAPair(ctx context.Context, req *knowledgebase_qa_service.UpdateQAPairReq) (*emptypb.Empty, error) {
	//1.查询问答对详情
	qaPair, err := orm.GetQAPairInfoById(ctx, req.QaPairId, "", "")
	if err != nil {
		log.Errorf("get qa pair info fail err: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairSelectFailed)
	}
	//2。查询问答库详情
	knowledgeBase, err := orm.SelectKnowledgeById(ctx, qaPair.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("select QA knowledge failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQABaseSelectFailed)
	}
	//3.校验问答对
	question := strings.Trim(req.Question, " ")
	answer := strings.Trim(req.Answer, " ")
	questionMD5 := util.MD5(question)
	if string(qaPair.Question) == question && string(qaPair.Answer) == answer {
		return nil, nil
	}
	questionOmitempty, answerOmitempty := string(qaPair.Question) == question, string(qaPair.Answer) == answer
	// 4.更新问答对
	qaPair, ragParams := buildUpdateQAPairParams(knowledgeBase, question, answer, questionMD5, req.QaPairId, questionOmitempty, answerOmitempty)
	err = orm.UpdateKnowledgeQAPair(ctx, qaPair, ragParams)
	if err != nil {
		log.Errorf("update knowledge qa pair fail: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeUpdateQAPairFailed)
	}
	return nil, nil
}

func (s *Service) UpdateQAPairSwitch(ctx context.Context, req *knowledgebase_qa_service.UpdateQAPairSwitchReq) (*emptypb.Empty, error) {
	//1.查询问答对详情
	qaPair, err := orm.GetQAPairInfoById(ctx, req.QaPairId, "", "")
	if err != nil {
		log.Errorf("get qa pair info fail err: %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeQAPairSelectFailed)
	}
	//2。查询问答库详情
	knowledgeBase, err := orm.SelectKnowledgeById(ctx, qaPair.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("select QA knowledge failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQABaseSelectFailed)
	}
	// 3.启停问答对
	qaPair, ragParams := buildUpdateQAPairSwitchParams(knowledgeBase, req.Switch, req.QaPairId)
	err = orm.UpdateKnowledgeQAPairSwitch(ctx, qaPair, ragParams)
	if err != nil {
		log.Errorf("update knowledge qa pair switch fail: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeUpdateQAPairSwitchFailed)
	}
	return nil, nil
}

func (s *Service) DeleteQAPair(ctx context.Context, req *knowledgebase_qa_service.DeleteQAPairReq) (*emptypb.Empty, error) {
	//1。查询问答库详情
	knowledgeBase, err := orm.SelectKnowledgeById(ctx, req.KnowledgeId, "", "")
	if err != nil {
		log.Errorf("select QA knowledge failed err: (%v) req:(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeQABaseSelectFailed)
	}
	// 2。删除问答对
	ragParams := buildDeleteQAPairParams(knowledgeBase, req.QaPairIds)
	err = orm.DeleteKnowledgeQAPair(ctx, req.KnowledgeId, req.QaPairIds, ragParams)
	if err != nil {
		log.Errorf("delete knowledge qa pair fail: %v", err)
		return nil, util.ErrCode(errs.Code_KnowledgeDeleteQAPairFailed)
	}
	return nil, nil
}

func (s *Service) KnowledgeQAHit(ctx context.Context, req *knowledgebase_qa_service.KnowledgeQAHitReq) (*knowledgebase_qa_service.KnowledgeQAHitResp, error) {
	// 1.获取问答库信息列表
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
	ragHitParams, err := buildRagQAHitParams(req, list, knowledgeIDToName)
	if err != nil {
		return nil, util.ErrCode(errs.Code_KnowledgeBaseHitFailed)
	}
	hitResp, err := service.RagKnowledgeQAHit(ctx, ragHitParams)
	if err != nil {
		log.Errorf("RagKnowledgeQAHit error %s", err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseHitFailed)
	}
	return buildKnowledgeBaseHitResp(hitResp), nil
}

// buildDeleteQAPairParams 构造问答库删除问答对参数
func buildDeleteQAPairParams(knowledgeBase *model.KnowledgeBase, qaPairIds []string) *service.RagDeleteQAPairParams {
	ragUpdateQAPairParams := &service.RagDeleteQAPairParams{
		UserId:     knowledgeBase.UserId,
		QAId:       knowledgeBase.KnowledgeId,
		QABaseName: knowledgeBase.RagName,
		QAPairIds:  qaPairIds,
	}
	return ragUpdateQAPairParams
}

func buildKnowledgeBaseHitResp(hitResp *service.RagKnowledgeQAHitResp) *knowledgebase_qa_service.KnowledgeQAHitResp {
	knowledgeHitData := hitResp.Data
	var searchList = make([]*knowledgebase_qa_service.KnowledgeQASearchInfo, 0)
	list := knowledgeHitData.SearchList
	if len(list) > 0 {
		for _, search := range list {
			searchList = append(searchList, &knowledgebase_qa_service.KnowledgeQASearchInfo{
				Title:       search.Title,
				Question:    search.Question,
				Answer:      search.Answer,
				QaPairId:    search.QAPairId,
				QaBase:      search.QABase,
				QaId:        search.QAId,
				ContentType: search.ContentType,
			})
		}
	}
	return &knowledgebase_qa_service.KnowledgeQAHitResp{
		Score:      knowledgeHitData.Score,
		SearchList: searchList,
	}
}

func buildRagQAHitParams(req *knowledgebase_qa_service.KnowledgeQAHitReq, list []*model.KnowledgeBase, knowledgeIDToName map[string]string) (*service.KnowledgeQAHitParams, error) {
	matchParams := req.KnowledgeMatchParams
	priorityMatch := matchParams.PriorityMatch
	filterEnable, metaParams, err := buildRagQAHitMetaParams(req, knowledgeIDToName)
	if err != nil {
		return nil, err
	}
	idList := buildKnowledgeList(list)
	ret := &service.KnowledgeQAHitParams{
		UserId:                      req.UserId,
		Question:                    req.Question,
		KnowledgeIdList:             idList,
		TopK:                        int64(matchParams.TopK),
		Threshold:                   float64(matchParams.Score),
		RerankModelId:               buildRerankId(priorityMatch, matchParams.RerankModelId),
		RetrieveMethod:              buildRetrieveMethod(matchParams.MatchType),
		RerankMod:                   buildRerankMod(priorityMatch),
		Weight:                      buildWeight(priorityMatch, matchParams.SemanticsPriority, matchParams.KeywordPriority),
		MetadataFiltering:           filterEnable,
		MetadataFilteringConditions: metaParams,
	}
	return ret, nil
}

// buildKnowledgeList 构造问答库id列表
func buildKnowledgeList(knowledgeList []*model.KnowledgeBase) []string {
	if len(knowledgeList) == 0 {
		return make([]string, 0)
	}
	knowledgeIdList := make([]string, 0, len(knowledgeList))
	for _, k := range knowledgeList {
		knowledgeIdList = append(knowledgeIdList, k.KnowledgeId)
	}
	return knowledgeIdList
}
func isValidFilterParams(params *knowledgebase_qa_service.MetaDataFilterParams) bool {
	return params != nil &&
		params.FilterEnable &&
		params.MetaFilterParams != nil &&
		len(params.MetaFilterParams) > 0
}

func buildRagQAHitMetaParams(req *knowledgebase_qa_service.KnowledgeQAHitReq, knowledgeIDToName map[string]string) (bool, []*service.QAMetadataFilterItem, error) {
	filterEnable := false // 标记是否有启用的元数据过滤
	var metaFilterConditions []*service.QAMetadataFilterItem
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
		metaItems, err := buildRagQAHitMetaItems(k.KnowledgeId, filterParams.MetaFilterParams)
		if err != nil {
			return false, nil, err
		}
		// 添加过滤项到结果
		metaFilterConditions = append(metaFilterConditions, &service.QAMetadataFilterItem{
			FilteringQaBaseName: knowledgeIDToName[k.KnowledgeId],
			LogicalOperator:     filterParams.FilterLogicType,
			Conditions:          metaItems,
		})
	}
	return filterEnable, metaFilterConditions, nil
}

func buildRagQAHitMetaItems(knowledgeID string, params []*knowledgebase_qa_service.MetaFilterParams) ([]*service.QAMetaItem, error) {
	var metaItems []*service.QAMetaItem
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
		metaItems = append(metaItems, &service.QAMetaItem{
			MetaName:           param.Key,
			MetaType:           param.Type,
			ComparisonOperator: param.Condition,
			Value:              ragValue,
		})
	}
	return metaItems, nil
}

func buildValueData(valueType string, value string) (interface{}, error) {
	switch valueType {
	case model.MetaTypeNumber:
	case model.MetaTypeTime:
		return strconv.ParseInt(value, 10, 64)
	}
	return value, nil
}

func validateMetaFilterParam(knowledgeID string, param *knowledgebase_qa_service.MetaFilterParams) error {
	// 检查关键参数是否为空
	if param.Key == "" || param.Type == "" || param.Condition == "" {
		errMsg := "key/type/condition cannot be empty"
		log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
		return errors.New(errMsg)
	}

	// 检查空条件与值的匹配性
	if param.Condition == knowledge.MetaConditionEmpty || param.Condition == knowledge.MetaConditionNotEmpty {
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
func buildWeight(priorityType int32, semanticsPriority float32, keywordPriority float32) *service.WeightParams {
	if priorityType != 1 {
		return nil
	}
	return &service.WeightParams{
		VectorWeight: semanticsPriority,
		TextWeight:   keywordPriority,
	}
}

// buildImportTask 构造导入任务
func buildQAPairImportTask(req *knowledgebase_qa_service.ImportQAPairReq) (*model.KnowledgeQAPairImportTask, error) {
	docList := make([]*model.DocInfo, 0)
	for _, docInfo := range req.DocInfoList {
		docList = append(docList, &model.DocInfo{
			DocId:   docInfo.DocId,
			DocName: docInfo.DocName,
			DocUrl:  docInfo.DocUrl,
			DocType: docInfo.DocType,
			DocSize: docInfo.DocSize,
		})
	}
	docImportInfo, err := json.Marshal(&model.DocImportInfo{
		DocInfoList: docList,
	})
	if err != nil {
		return nil, err
	}
	return &model.KnowledgeQAPairImportTask{
		ImportId:    wanwu_util.NewID(),
		KnowledgeId: req.KnowledgeId,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		DocInfo:     db.LongText(docImportInfo),
		Status:      model.KnowledgeQAPairImportInit,
		UserId:      req.UserId,
		OrgId:       req.OrgId,
	}, nil
}

// buildExportTask 构造导出任务
func buildQAPairExportTask(req *knowledgebase_qa_service.ExportQAPairReq) (*model.KnowledgeExportTask, error) {
	return &model.KnowledgeExportTask{
		ExportId:    wanwu_util.NewID(),
		KnowledgeId: req.KnowledgeId,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		Status:      model.KnowledgeExportInit,
		UserId:      req.UserId,
		OrgId:       req.OrgId,
	}, nil
}

// buildQAPairListResp 构造问答库问答对列表
func buildQAPairListResp(list []*model.KnowledgeQAPair, knowledge *model.KnowledgeBase, docMetaList []*model.KnowledgeDocMeta, total int64, pageSize int32, pageNum int32) *knowledgebase_qa_service.GetQAPairListResp {
	var retList = make([]*knowledgebase_qa_service.QAPairInfo, 0)
	metaMap := buildQAPairMetaMap(docMetaList)
	if len(list) > 0 {
		for _, item := range list {
			retList = append(retList, &knowledgebase_qa_service.QAPairInfo{
				QaPairId:     item.QAPairId,
				KnowledgeId:  item.KnowledgeId,
				Question:     string(item.Question),
				Answer:       string(item.Answer),
				Status:       int32(item.Status),
				Switch:       item.Switch,
				ErrorMsg:     string(item.ErrorMsg),
				UploadTime:   util2.Time2Str(item.CreatedAt),
				UserId:       item.UserId,
				MetaDataList: buildMetaList(metaMap, item.QAPairId),
			})
		}
	}
	return &knowledgebase_qa_service.GetQAPairListResp{
		Total:       total,
		QaPairInfos: retList,
		PageSize:    pageSize,
		PageNum:     pageNum,
		KnowledgeInfo: &knowledgebase_qa_service.KnowledgeInfo{
			KnowledgeId:   knowledge.KnowledgeId,
			KnowledgeName: knowledge.Name,
		},
	}
}

// buildCreateQAPairParams 构造问答库新建问答对参数
func buildCreateQAPairParams(knowledgeBase *model.KnowledgeBase, question, answer, questionMD5, qaPairId, userId, orgId string) ([]*model.KnowledgeQAPair, *service.RagAddQAPairParams) {
	qaPairs := []*model.KnowledgeQAPair{&model.KnowledgeQAPair{
		QAPairId:    qaPairId,
		KnowledgeId: knowledgeBase.KnowledgeId,
		Question:    db.LongText(question),
		Answer:      db.LongText(answer),
		Status:      model.KnowledgeQAPairImportSuccess,
		Switch:      true,
		QuestionMd5: questionMD5,
		UserId:      userId,
		OrgId:       orgId,
	}}
	ragAddQAPairParams := &service.RagAddQAPairParams{
		UserId:     knowledgeBase.UserId,
		QAId:       knowledgeBase.KnowledgeId,
		QABaseName: knowledgeBase.RagName,
		QAPairs: []*service.RagQAPairItem{&service.RagQAPairItem{
			QAPairId: qaPairId,
			Question: question,
			Answer:   answer,
		}},
	}
	return qaPairs, ragAddQAPairParams
}

// buildUpdateQAPairParams 构造问答库更新问答对参数
func buildUpdateQAPairParams(knowledgeBase *model.KnowledgeBase, question, answer, questionMD5, qaPairId string, qesOmi, ansOmi bool) (*model.KnowledgeQAPair, *service.RagUpdateQAPairParams) {
	qaPair := &model.KnowledgeQAPair{
		QAPairId:    qaPairId,
		Question:    db.LongText(question),
		Answer:      db.LongText(answer),
		QuestionMd5: questionMD5,
		KnowledgeId: knowledgeBase.KnowledgeId,
	}
	qaPairItem := &service.RagQAPairItem{
		QAPairId: qaPairId,
	}
	if !qesOmi {
		qaPairItem.Question = question
	}
	if !ansOmi {
		qaPairItem.Answer = answer
	}
	ragUpdateQAPairParams := &service.RagUpdateQAPairParams{
		UserId:     knowledgeBase.UserId,
		QAId:       knowledgeBase.KnowledgeId,
		QABaseName: knowledgeBase.RagName,
		QAPair:     qaPairItem,
	}
	return qaPair, ragUpdateQAPairParams
}

// buildUpdateQAPairSwitchParams 构造启停问答对参数
func buildUpdateQAPairSwitchParams(knowledgeBase *model.KnowledgeBase, qaPairSwitch bool, qaPairId string) (*model.KnowledgeQAPair, *service.RagUpdateQAPairStatusParams) {
	qaPair := &model.KnowledgeQAPair{
		QAPairId:    qaPairId,
		Switch:      qaPairSwitch,
		KnowledgeId: knowledgeBase.KnowledgeId,
	}
	ragUpdateQAPairStatusParams := &service.RagUpdateQAPairStatusParams{
		UserId:     knowledgeBase.UserId,
		QAId:       knowledgeBase.KnowledgeId,
		QABaseName: knowledgeBase.RagName,
		QAPairId:   qaPairId,
		Status:     qaPairSwitch,
	}
	return qaPair, ragUpdateQAPairStatusParams
}

// buildQAPairInfo 构造问答库问答对
func buildQAPairInfo(item *model.KnowledgeQAPair) *knowledgebase_qa_service.QAPairInfo {
	return &knowledgebase_qa_service.QAPairInfo{
		QaPairId:    item.QAPairId,
		KnowledgeId: item.KnowledgeId,
		Question:    string(item.Question),
		Answer:      string(item.Answer),
		UploadTime:  util2.Time2Str(item.CreatedAt),
		Status:      int32(item.Status),
		ErrorMsg:    string(item.ErrorMsg),
		Switch:      item.Switch,
		UserId:      item.UserId,
	}
}

// buildQAPairMetaMap
func buildQAPairMetaMap(qaPairMetaList []*model.KnowledgeDocMeta) map[string][]*model.KnowledgeDocMeta {
	qaPairMetaMap := make(map[string][]*model.KnowledgeDocMeta)
	if len(qaPairMetaList) > 0 {
		for _, v := range qaPairMetaList {
			if _, exists := qaPairMetaMap[v.DocId]; !exists {
				qaPairMetaMap[v.DocId] = make([]*model.KnowledgeDocMeta, 0)
			}
			if v.ValueMain != "" {
				qaPairMetaMap[v.DocId] = append(qaPairMetaMap[v.DocId], v)
			}
		}
	}
	return qaPairMetaMap
}

func buildMetaList(metaMaps map[string][]*model.KnowledgeDocMeta, qaId string) []*knowledgebase_qa_service.MetaData {
	if _, exists := metaMaps[qaId]; !exists {
		return make([]*knowledgebase_qa_service.MetaData, 0)
	}
	return lo.Map(metaMaps[qaId], func(item *model.KnowledgeDocMeta, index int) *knowledgebase_qa_service.MetaData {
		var valueType = item.ValueType
		if valueType == "" {
			valueType = model.MetaTypeString
		}
		return &knowledgebase_qa_service.MetaData{
			MetaId:    item.MetaId,
			Key:       item.Key,
			Value:     item.ValueMain,
			ValueType: valueType,
			Rule:      item.Rule,
		}
	})
}
