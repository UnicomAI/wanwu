package service

import (
	"net/url"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgebase_qa_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-qa-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// ImportKnowledgeQAPair 导入问答对
func ImportKnowledgeQAPair(ctx *gin.Context, userId, orgId string, req *request.KnowledgeQAPairImportReq) error {
	if len(req.DocInfo) != 1 {
		return grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "文档数量超过限制")
	}
	docInfoList, err := buildQAPairDocInfoList(ctx, req)
	if err != nil {
		log.Errorf("上传失败(构建文档信息列表失败(%v) ", err)
		return err
	}
	_, err = knowledgeBaseQA.ImportQAPair(ctx.Request.Context(), &knowledgebase_qa_service.ImportQAPairReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: req.KnowledgeId,
		DocInfoList: docInfoList,
	})
	if err != nil {
		log.Errorf("上传失败(保存上传任务 失败(%v) ", err)
		return err
	}
	return nil
}

// GetKnowledgeQAPairImportTip 获取导入进度提示条
func GetKnowledgeQAPairImportTip(ctx *gin.Context, userId, orgId string, r *request.KnowledgeQAPairImportTipReq) (*response.KnowledgeQAPairImportTipResp, error) {
	resp, err := knowledgeBaseQA.GetQAImportTip(ctx.Request.Context(), &knowledgebase_qa_service.QAImportTipReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: r.KnowledgeId,
	})
	if err != nil {
		return nil, err
	}
	var message = ""
	if len(resp.Message) > 0 {
		message = gin_util.I18nKey(ctx, "know_doc_last_failure_info", resp.Message)
	}
	return &response.KnowledgeQAPairImportTipResp{
		Message:       message,
		UploadStatus:  resp.UploadStatus,
		KnowledgeId:   resp.KnowledgeId,
		KnowledgeName: resp.KnowledgeName,
	}, nil
}

// ExportKnowledgeQAPair 导出问答对
func ExportKnowledgeQAPair(ctx *gin.Context, userId, orgId string, req *request.KnowledgeQAPairExportReq) error {
	_, err := knowledgeBaseQA.ExportQAPair(ctx.Request.Context(), &knowledgebase_qa_service.ExportQAPairReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: req.KnowledgeId,
	})
	if err != nil {
		log.Errorf("导出失败(保存导出任务 失败(%v) ", err)
		return err
	}
	return nil
}

// GetQAPairDetail 查询问答库问答对详情
func GetQAPairDetail(ctx *gin.Context, userId, orgId, qaPairId string) (*response.ListKnowledgeQAPairResp, error) {
	data, err := knowledgeBaseQA.GetQAPairInfo(ctx.Request.Context(), &knowledgebase_qa_service.GetQAPairInfoReq{
		QaPairId: qaPairId,
		UserId:   userId,
		OrgId:    orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListKnowledgeQAPairResp{
		QAPairId:    data.QaPairId,
		Question:    data.Question,
		Answer:      data.Answer,
		UploadTime:  data.UploadTime,
		Status:      int(data.Status),
		ErrorMsg:    gin_util.I18nKey(ctx, data.ErrorMsg),
		KnowledgeId: data.KnowledgeId,
	}, nil
}

// GetKnowledgeQAPairList 查询问答库问答对列表
func GetKnowledgeQAPairList(ctx *gin.Context, userId, orgId string, r *request.KnowledgeQAPairListReq) (*response.KnowledgeQAPairPageResult, error) {
	resp, err := knowledgeBaseQA.GetQAPairList(ctx.Request.Context(), &knowledgebase_qa_service.GetQAPairListReq{
		KnowledgeId: r.KnowledgeId,
		Name:        strings.TrimSpace(r.Name),
		Status:      r.Status,
		PageSize:    int32(r.PageSize),
		PageNum:     int32(r.PageNo),
		UserId:      userId,
		OrgId:       orgId,
		MetaValue:   strings.TrimSpace(r.MetaValue),
	})
	if err != nil {
		return nil, err
	}
	knowledgeInfo := resp.KnowledgeInfo
	embModelInfo, _ := GetModel(ctx, userId, orgId, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{
			ModelId: knowledgeInfo.EmbeddingModelId,
		},
	})
	return &response.KnowledgeQAPairPageResult{
		List:     buildQAPairRespList(ctx, resp.QaPairInfos),
		Total:    resp.Total,
		PageNo:   int(resp.PageNum),
		PageSize: int(resp.PageSize),
		QAKnowledgeInfo: &response.QAKnowledgeInfo{
			KnowledgeId:    knowledgeInfo.KnowledgeId,
			KnowledgeName:  knowledgeInfo.KnowledgeName,
			Description:    knowledgeInfo.Description,
			EmbeddingModel: embModelInfo,
			Avatar:         cacheKnowledgeAvatar(ctx, knowledgeInfo.AvatarPath, knowledgeInfo.Category),
			PermissionType: knowledgeInfo.PermissionType,
		},
	}, nil
}

// GetKnowledgeExportRecordList 查询知识库导出记录列表
func GetKnowledgeExportRecordList(ctx *gin.Context, userId, orgId string, r *request.KnowledgeExportRecordListReq) (*response.KnowledgeExportRecordPageResult, error) {
	resp, err := knowledgeBase.GetExportRecordList(ctx.Request.Context(), &knowledgebase_service.GetExportRecordListReq{
		KnowledgeId: r.KnowledgeId,
		UserId:      userId,
		OrgId:       orgId,
		PageSize:    int32(r.PageSize),
		PageNum:     int32(r.PageNo),
	})
	if err != nil {
		return nil, err
	}
	return &response.KnowledgeExportRecordPageResult{
		List:     buildKnowledgeExportRecordRespList(ctx, resp.ExportRecordInfos),
		Total:    resp.Total,
		PageNo:   int(resp.PageNum),
		PageSize: int(resp.PageSize),
	}, nil
}

// CreateKnowledgeQAPair 新建问答对
func CreateKnowledgeQAPair(ctx *gin.Context, userId, orgId string, req *request.CreateKnowledgeQAPairReq) (*response.CreateKnowledgeQAPairResp, error) {
	resp, err := knowledgeBaseQA.CreateQAPair(ctx.Request.Context(), &knowledgebase_qa_service.CreateQAPairReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: req.KnowledgeId,
		Question:    req.Question,
		Answer:      req.Answer,
	})
	if err != nil {
		log.Errorf("新建问答对 失败(%v) ", err)
		return nil, err
	}
	return &response.CreateKnowledgeQAPairResp{
		QAPairId: resp.QaPairId,
	}, nil
}

// UpdateKnowledgeQAPair 更新问答对
func UpdateKnowledgeQAPair(ctx *gin.Context, userId, orgId string, req *request.UpdateKnowledgeQAPairReq) error {
	_, err := knowledgeBaseQA.UpdateQAPair(ctx.Request.Context(), &knowledgebase_qa_service.UpdateQAPairReq{
		UserId:   userId,
		OrgId:    orgId,
		QaPairId: req.QAPairId,
		Question: req.Question,
		Answer:   req.Answer,
	})
	if err != nil {
		log.Errorf("更新问答对 失败(%v) ", err)
		return err
	}
	return nil
}

// UpdateKnowledgeQAPairSwitch 启停问答对
func UpdateKnowledgeQAPairSwitch(ctx *gin.Context, userId, orgId string, req *request.UpdateKnowledgeQAPairSwitchReq) error {
	_, err := knowledgeBaseQA.UpdateQAPairSwitch(ctx.Request.Context(), &knowledgebase_qa_service.UpdateQAPairSwitchReq{
		UserId:   userId,
		OrgId:    orgId,
		QaPairId: req.QAPairId,
		Switch:   req.Switch,
	})
	if err != nil {
		log.Errorf("启停问答对 失败(%v) ", err)
		return err
	}
	return nil
}

// DeleteKnowledgeQAPair 删除问答对
func DeleteKnowledgeQAPair(ctx *gin.Context, userId, orgId string, req *request.DeleteKnowledgeQAPairReq) error {
	_, err := knowledgeBaseQA.DeleteQAPair(ctx.Request.Context(), &knowledgebase_qa_service.DeleteQAPairReq{
		KnowledgeId: req.KnowledgeId,
		QaPairIds:   req.QAPairIdList,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		log.Errorf("删除问答对 失败(%v) ", err)
		return err
	}
	return nil
}

// DeleteKnowledgeExportRecord 删除导出记录
func DeleteKnowledgeExportRecord(ctx *gin.Context, userId, orgId string, req *request.DeleteKnowledgeExportRecordReq) error {
	_, err := knowledgeBase.DeleteExportRecord(ctx.Request.Context(), &knowledgebase_service.DeleteExportRecordReq{
		UserId:         userId,
		OrgId:          orgId,
		ExportRecordId: req.ExportRecordId,
	})
	if err != nil {
		log.Errorf("删除导出记录 失败(%v) ", err)
		return err
	}
	return nil
}

// internal
func buildQAPairDocInfoList(ctx *gin.Context, req *request.KnowledgeQAPairImportReq) ([]*knowledgebase_qa_service.DocFileInfo, error) {
	var docInfoList []*knowledgebase_qa_service.DocFileInfo
	for _, info := range req.DocInfo {
		var docUrl = info.DocUrl
		var docType = info.DocType
		if len(docUrl) == 0 {
			var err error
			docUrl, err = minio.GetUploadFileWithExpire(ctx, info.DocId)
			if err != nil {
				log.Errorf("GetUploadFileWithNotExpire error %v", err)
				return nil, grpc_util.ErrorStatus(errs.Code_KnowledgeDocImportUrlFailed)
			}
			//特殊处理类型
			if strings.HasSuffix(docUrl, ".tar.gz") {
				docType = ".tar.gz"
			}
		}
		docInfoList = append(docInfoList, &knowledgebase_qa_service.DocFileInfo{
			DocName: info.DocName,
			DocId:   info.DocId,
			DocUrl:  docUrl,
			DocType: docType,
			DocSize: info.DocSize,
		})
	}
	return docInfoList, nil
}

// buildQAPairRespList 构造问答对返回列表
func buildQAPairRespList(ctx *gin.Context, dataList []*knowledgebase_qa_service.QAPairInfo) []*response.ListKnowledgeQAPairResp {
	retList := make([]*response.ListKnowledgeQAPairResp, 0)
	authorMap := buildQAPairAuthorMap(ctx, dataList)
	for _, data := range dataList {
		retList = append(retList, &response.ListKnowledgeQAPairResp{
			QAPairId:     data.QaPairId,
			KnowledgeId:  data.KnowledgeId,
			Question:     data.Question,
			Answer:       data.Answer,
			UploadTime:   data.UploadTime,
			Status:       int(data.Status),
			ErrorMsg:     gin_util.I18nKey(ctx, data.ErrorMsg),
			Author:       authorMap[data.UserId],
			Switch:       data.Switch,
			MetaDataList: buildQAPairMetaDataResultList(data.MetaDataList),
		})
	}
	return retList
}

// buildKnowledgeExportRecordRespList 构造知识库导出记录返回列表
func buildKnowledgeExportRecordRespList(ctx *gin.Context, dataList []*knowledgebase_service.ExportRecordInfo) []*response.ListKnowledgeExportRecordResp {
	retList := make([]*response.ListKnowledgeExportRecordResp, 0)
	authorMap := buildKnowledgeExportAuthorMap(ctx, dataList)
	for _, data := range dataList {
		filePath, _ := url.JoinPath(config.Cfg().Minio.DownloadURL, data.FilePath)
		filePath = urlDecode(filePath)
		retList = append(retList, &response.ListKnowledgeExportRecordResp{
			ExportRecordId: data.ExportRecordId,
			ExportTime:     data.ExportTime,
			FilePath:       filePath,
			Status:         int(data.Status),
			ErrorMsg:       gin_util.I18nKey(ctx, data.ErrorMsg),
			Author:         authorMap[data.UserId],
			KnowledgeName:  data.KnowledgeName,
		})
	}
	return retList
}

// URLDecode 对输入的字符串进行 URL 解码。
// 参数 encodeStr 为需要解码的字符串（例如 "%E7%9F%A5%E8%AF%86%E5%BA%931.zip"）。
// 返回解码后的字符串和可能的错误。
func urlDecode(encodeStr string) string {
	// 使用 QueryUnescape 可以处理 %xx 形式的编码，并且会将 '+' 还原为空格
	decoded, err := url.QueryUnescape(encodeStr)
	if err != nil {
		log.Errorf("URL decode error: %v", err)
		return encodeStr
	}
	return decoded
}

func buildQAPairAuthorMap(ctx *gin.Context, dataList []*knowledgebase_qa_service.QAPairInfo) map[string]string {
	authorMap := make(map[string]string)
	userIdSet := make(map[string]bool)
	for _, data := range dataList {
		if data.UserId != "" {
			userIdSet[data.UserId] = true
			authorMap[data.UserId] = ""
		}
	}
	if len(userIdSet) == 0 {
		return authorMap
	}
	userIdList := make([]string, len(userIdSet))
	for userId := range userIdSet {
		userIdList = append(userIdList, userId)
	}
	userInfoList, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: userIdList,
	})
	if err != nil {
		log.Errorf("knowledge gets user info error: %v", err)
		return authorMap
	}
	for _, userInfo := range userInfoList.Selects {
		if userInfo.Id != "" {
			authorMap[userInfo.Id] = userInfo.Name
		}
	}
	return authorMap
}

func buildKnowledgeExportAuthorMap(ctx *gin.Context, dataList []*knowledgebase_service.ExportRecordInfo) map[string]string {
	authorMap := make(map[string]string)
	userIdSet := make(map[string]bool)
	for _, data := range dataList {
		if data.UserId != "" {
			userIdSet[data.UserId] = true
			authorMap[data.UserId] = ""
		}
	}
	if len(userIdSet) == 0 {
		return authorMap
	}
	userIdList := make([]string, len(userIdSet))
	for userId := range userIdSet {
		userIdList = append(userIdList, userId)
	}
	userInfoList, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: userIdList,
	})
	if err != nil {
		log.Errorf("knowledge gets user info error: %v", err)
		return authorMap
	}
	for _, userInfo := range userInfoList.Selects {
		if userInfo.Id != "" {
			authorMap[userInfo.Id] = userInfo.Name
		}
	}
	return authorMap
}

// KnowledgeQAHit 知识库命中
func KnowledgeQAHit(ctx *gin.Context, userId, orgId string, r *request.KnowledgeQAHitReq) (*response.KnowledgeQAHitResp, error) {
	matchParams := r.KnowledgeMatchParams
	resp, err := knowledgeBaseQA.KnowledgeQAHit(ctx.Request.Context(), &knowledgebase_qa_service.KnowledgeQAHitReq{
		Question:      r.Question,
		UserId:        userId,
		OrgId:         orgId,
		KnowledgeList: buildKnowledgeQAListReq(r),
		KnowledgeMatchParams: &knowledgebase_qa_service.KnowledgeMatchParams{
			MatchType:         matchParams.MatchType,
			RerankModelId:     matchParams.RerankModelId,
			PriorityMatch:     matchParams.PriorityMatch,
			SemanticsPriority: matchParams.SemanticsPriority,
			KeywordPriority:   matchParams.KeywordPriority,
			TopK:              matchParams.TopK,
			Score:             matchParams.Threshold,
			TermWeight:        matchParams.TermWeight,
			TermWeightEnable:  matchParams.TermWeightEnable,
		},
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeQAHitResp(resp), nil
}

// buildKnowledgeQAListReq 构造命中测试 - 问答库列表参数
func buildKnowledgeQAListReq(r *request.KnowledgeQAHitReq) []*knowledgebase_qa_service.KnowledgeParams {
	var knowledgeList []*knowledgebase_qa_service.KnowledgeParams
	for _, k := range r.KnowledgeList {
		kp := &knowledgebase_qa_service.KnowledgeParams{KnowledgeId: k.ID}
		// metaDataFilterParams 可选，不传时为 nil，需判空避免解引用 panic
		if k.MetaDataFilterParams != nil {
			kp.MetaDataFilterParams = &knowledgebase_qa_service.MetaDataFilterParams{
				FilterEnable:     k.MetaDataFilterParams.FilterEnable,
				FilterLogicType:  k.MetaDataFilterParams.FilterLogicType,
				MetaFilterParams: buildQAMetaFilterParams(k.MetaDataFilterParams.MetaFilterParams),
			}
		}
		knowledgeList = append(knowledgeList, kp)
	}
	return knowledgeList
}

func buildQAMetaFilterParams(meta []*request.MetaFilterParams) []*knowledgebase_qa_service.MetaFilterParams {
	var metaList []*knowledgebase_qa_service.MetaFilterParams
	for _, m := range meta {
		metaList = append(metaList, &knowledgebase_qa_service.MetaFilterParams{
			Key:       m.Key,
			Value:     m.Value,
			Type:      m.Type,
			Condition: m.Condition,
		})
	}
	return metaList
}

func buildQAPairMetaDataResultList(metaDataList []*knowledgebase_qa_service.MetaData) []*response.DocMetaData {
	if len(metaDataList) == 0 {
		return make([]*response.DocMetaData, 0)
	}
	return lo.Map(metaDataList, func(item *knowledgebase_qa_service.MetaData, index int) *response.DocMetaData {
		return &response.DocMetaData{
			MetaId:        item.MetaId,
			MetaKey:       item.Key,
			MetaValue:     item.Value,
			MetaValueType: item.ValueType,
			MetaRule:      item.Rule,
		}
	})
}

// buildKnowledgeQAHitResp 构造问答库命中返回
func buildKnowledgeQAHitResp(resp *knowledgebase_qa_service.KnowledgeQAHitResp) *response.KnowledgeQAHitResp {
	var searchList = make([]*response.QAHitSearchList, 0)
	if len(resp.SearchList) > 0 {
		for _, search := range resp.SearchList {
			searchList = append(searchList, &response.QAHitSearchList{
				Title:       search.Title,
				Question:    search.Question,
				Answer:      search.Answer,
				QAPairId:    search.QaPairId,
				QABase:      search.QaBase,
				QAId:        search.QaId,
				ContentType: search.ContentType,
			})
		}
	}
	return &response.KnowledgeQAHitResp{
		Score:      resp.Score,
		SearchList: searchList,
	}
}
