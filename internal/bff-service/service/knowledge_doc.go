package service

import (
	"encoding/json"
	"path/filepath"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgebase_doc_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-doc-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

const (
	AutoSegment        = "0" //自动分段
	ParentChildSegment = "1" //父子分段
)

var docAnalyzerMap = map[string]string{
	"text":       "文字提取",
	"ocr":        "OCR解析",
	"model":      "模型解析",
	"asr":        "ASR",
	"multimodal": "图文问答模型",
}

// GetDocList 查询知识库所属文档列表
func GetDocList(ctx *gin.Context, userId, orgId string, r *request.DocListReq) (*response.DocPageResult, error) {
	resp, err := knowledgeBaseDoc.GetDocList(ctx.Request.Context(), &knowledgebase_doc_service.GetDocListReq{
		KnowledgeId:   r.KnowledgeId,
		DocName:       strings.TrimSpace(r.DocName),
		Status:        r.Status,
		PageSize:      int32(r.PageSize),
		PageNum:       int32(r.PageNo),
		UserId:        userId,
		OrgId:         orgId,
		MetaValue:     strings.TrimSpace(r.MetaValue),
		MetaType:      strings.TrimSpace(r.MetaType),
		MetaStartTime: strings.TrimSpace(r.MetaStartTime),
		MetaEndTime:   strings.TrimSpace(r.MetaEndTime),
		GraphStatus:   r.GraphStatus,
		DocIdList:     r.DocIdList,
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
	return &response.DocPageResult{
		List:     buildDocRespList(ctx, resp.Docs, r.KnowledgeId),
		Total:    resp.Total,
		PageNo:   int(resp.PageNum),
		PageSize: int(resp.PageSize),
		DocKnowledgeInfo: &response.DocKnowledgeInfo{
			KnowledgeId:     knowledgeInfo.KnowledgeId,
			KnowledgeName:   knowledgeInfo.KnowledgeName,
			GraphSwitch:     knowledgeInfo.GraphSwitch,
			ShowGraphReport: knowledgeInfo.ShowGraphReport,
			Description:     knowledgeInfo.Description,
			Keywords:        buildKeywordsInfo(knowledgeInfo.Keywords),
			EmbeddingModel:  embModelInfo,
			LlmModelId:      knowledgeInfo.LlmModelId,
			Category:        knowledgeInfo.Category,
			Avatar:          cacheKnowledgeAvatar(ctx, knowledgeInfo.AvatarPath, knowledgeInfo.Category),
			PermissionType:  knowledgeInfo.PermissionType,
		},
	}, nil
}

// GetDocConfig 查询知识库文档配置
func GetDocConfig(ctx *gin.Context, userId, orgId string, r *request.DocConfigReq) (*response.DocConfigResult, error) {
	data, err := knowledgeBaseDoc.GetDocDetail(ctx.Request.Context(), &knowledgebase_doc_service.GetDocDetailReq{
		DocId:      r.DocId,
		NeedConfig: true,
	})
	if err != nil {
		return nil, err
	}
	configInfo := data.DocConfigInfo
	docSegment := configInfo.DocSegment

	return &response.DocConfigResult{
		DocSegment:        buildDocSegment(docSegment),
		DocPreprocess:     configInfo.DocPreprocess,
		DocAnalyzer:       configInfo.DocAnalyzer,
		DocImportType:     configInfo.DocImportType,
		ParserModelId:     configInfo.OcrModelId,
		AsrModelId:        configInfo.AsrModelId,
		MultimodalModelId: configInfo.MultimodalModelId,
	}, nil
}

// GetDocDetail 查询知识库所属文档详情
func GetDocDetail(ctx *gin.Context, userId, orgId, docId string) (*response.ListDocResp, error) {
	data, err := knowledgeBaseDoc.GetDocDetail(ctx.Request.Context(), &knowledgebase_doc_service.GetDocDetailReq{
		DocId:  docId,
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListDocResp{
		DocId:         data.DocId,
		DocName:       data.DocName,
		DocType:       data.DocType,
		UploadTime:    data.UploadTime,
		Status:        int(data.Status),
		ErrorMsg:      gin_util.I18nKey(ctx, data.ErrorMsg),
		FileSize:      data.DocSize,
		KnowledgeId:   data.KnowledgeId,
		SegmentMethod: data.SegmentMethod,
		GraphStatus:   data.GraphStatus,
		GraphErrMsg:   data.GraphErrMsg,
		IsMultimodal:  data.IsMultimodal,
		DocProgress:   data.DocProgress,
		GraphProgress: data.GraphProgress,
	}, nil
}

// ImportDoc 导入文档
func ImportDoc(ctx *gin.Context, userId, orgId string, req *request.DocImportReq) error {
	segment := req.DocSegment
	docInfoList, err := buildDocInfoList(ctx, req.DocInfo)
	if err != nil {
		log.Errorf("上传失败(构建文档信息列表失败(%v) ", err)
		return err
	}
	_, err = knowledgeBaseDoc.ImportDoc(ctx.Request.Context(), &knowledgebase_doc_service.ImportDocReq{
		UserId:        userId,
		OrgId:         orgId,
		KnowledgeId:   req.KnowledgeId,
		DocImportType: int32(req.DocImportType),
		DocSegment: &knowledgebase_doc_service.DocSegment{
			SegmentType:    segment.SegmentType,
			Splitter:       segment.Splitter,
			MaxSplitter:    int32(segment.MaxSplitter),
			Overlap:        segment.Overlap,
			SegmentMethod:  segment.SegmentMethod,
			SubMaxSplitter: int32(segment.SubMaxSplitter),
			SubSplitter:    segment.SubSplitter,
		},
		DocAnalyzer:       req.DocAnalyzer,
		DocInfoList:       docInfoList,
		OcrModelId:        req.ParserModelId,
		AsrModelId:        req.AsrModelId,
		MultimodalModelId: req.MultimodalModelId,
		DocPreprocess:     req.DocPreprocess,
		DocMetaDataList:   buildMetaInfoList(req),
	})
	if err != nil {
		log.Errorf("上传失败(保存上传任务 失败(%v) ", err)
		return err
	}
	return nil
}

// ImportDocOpenapi 导入文档
func ImportDocOpenapi(ctx *gin.Context, userId, orgId string, req *request.DocImportReq) error {
	var err error
	if req.ParserModelId != "" {
		req.ParserModelId, err = GetModelIdByUuid(ctx, req.ParserModelId)
		if err != nil {
			return err
		}
	}
	return ImportDoc(ctx, userId, orgId, req)
}

// UpdateDocConfig 更新文档配置
func UpdateDocConfig(ctx *gin.Context, userId, orgId string, req *request.DocConfigUpdateReq) error {
	segment := req.DocSegment
	_, err := knowledgeBaseDoc.UpdateDocImportConfig(ctx.Request.Context(), &knowledgebase_doc_service.UpdateDocImportConfigReq{
		KnowledgeId: req.KnowledgeId,
		DocIdList:   req.DocIdList,
		ImportDocReq: &knowledgebase_doc_service.ImportDocReq{
			UserId:        userId,
			OrgId:         orgId,
			KnowledgeId:   req.KnowledgeId,
			DocImportType: int32(req.DocImportType),
			DocSegment: &knowledgebase_doc_service.DocSegment{
				SegmentType:    segment.SegmentType,
				Splitter:       segment.Splitter,
				MaxSplitter:    int32(segment.MaxSplitter),
				Overlap:        segment.Overlap,
				SegmentMethod:  segment.SegmentMethod,
				SubMaxSplitter: int32(segment.SubMaxSplitter),
				SubSplitter:    segment.SubSplitter,
			},
			DocAnalyzer:       req.DocAnalyzer,
			OcrModelId:        req.ParserModelId,
			DocPreprocess:     req.DocPreprocess,
			AsrModelId:        req.AsrModelId,
			MultimodalModelId: req.MultimodalModelId,
		},
	})
	if err != nil {
		log.Errorf("文档配置更新失败(%v) ", err)
		return err
	}
	return nil
}

// UpdateDocConfigOpenapi 更新文档配置
func UpdateDocConfigOpenapi(ctx *gin.Context, userId, orgId string, req *request.DocConfigUpdateReq) error {
	var err error
	if req.ParserModelId != "" {
		req.ParserModelId, err = GetModelIdByUuid(ctx, req.ParserModelId)
		if err != nil {
			return err
		}
	}
	return UpdateDocConfig(ctx, userId, orgId, req)
}

// ReImportDoc 重新解析文档
func ReImportDoc(ctx *gin.Context, userId, orgId string, req *request.DocReImportReq) error {
	_, err := knowledgeBaseDoc.ReImportDoc(ctx.Request.Context(), &knowledgebase_doc_service.ReImportDocReq{
		KnowledgeId: req.KnowledgeId,
		DocIdList:   req.DocIdList,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		log.Errorf("文档重新解析失败(%v) ", err)
		return err
	}
	return nil
}

// UpdateDocMetaData 更新文档元数据
func UpdateDocMetaData(ctx *gin.Context, userId, orgId string, r *request.DocMetaDataReq) error {
	_, err := knowledgeBaseDoc.UpdateDocMetaData(ctx.Request.Context(), &knowledgebase_doc_service.UpdateDocMetaDataReq{
		UserId:       userId,
		OrgId:        orgId,
		DocId:        r.DocId,
		MetaDataList: buildMetaDataList(r.MetaDataList),
		KnowledgeId:  r.KnowledgeId,
	})
	return err
}

// BatchUpdateDocMetaData 批量文档元数据
func BatchUpdateDocMetaData(ctx *gin.Context, userId, orgId string, r *request.BatchDocMetaDataReq) error {
	_, err := knowledgeBaseDoc.BatchUpdateDocMetaData(ctx.Request.Context(), &knowledgebase_doc_service.BatchUpdateDocMetaDataReq{
		UserId:       userId,
		OrgId:        orgId,
		MetaDataList: buildMetaDataList(r.MetaDataList),
		KnowledgeId:  r.KnowledgeId,
	})
	return err
}

func UpdateDocStatus(ctx *gin.Context, r *request.CallbackUpdateDocStatusReq) error {
	_, err := knowledgeBaseDoc.UpdateDocStatus(ctx.Request.Context(), &knowledgebase_doc_service.UpdateDocStatusReq{
		DocId:        r.DocId,
		Status:       r.Status,
		MetaDataList: buildCallbackMetaDataList(r.MetaDataList),
	})
	return err
}

func DocStatusInit(ctx *gin.Context, userId, orgId string) (interface{}, error) {
	_, err := knowledgeBaseDoc.InitDocStatus(ctx.Request.Context(), &knowledgebase_doc_service.InitDocStatusReq{
		UserId: userId,
		OrgId:  orgId,
	})
	return nil, err
}

func GetDocImportTip(ctx *gin.Context, userId, orgId string, r *request.QueryKnowledgeReq) (*response.DocImportTipResp, error) {
	resp, err := knowledgeBaseDoc.GetDocCategoryUploadTip(ctx.Request.Context(), &knowledgebase_doc_service.DocImportTipReq{
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
	return &response.DocImportTipResp{
		Message:       message,
		UploadStatus:  resp.UploadStatus,
		KnowledgeId:   resp.KnowledgeId,
		KnowledgeName: resp.KnowledgeName,
	}, nil
}

func DeleteDoc(ctx *gin.Context, userId, orgId string, r *request.DeleteDocReq) error {
	_, err := knowledgeBaseDoc.DeleteDoc(ctx.Request.Context(), &knowledgebase_doc_service.DeleteDocReq{
		Ids:         r.DocIdList,
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: r.KnowledgeId,
	})
	return err
}

func GetDocSegmentList(ctx *gin.Context, userId, orgId string, req *request.DocSegmentListReq) (*response.DocSegmentResp, error) {
	resp, err := knowledgeBaseDoc.GetDocSegmentList(ctx.Request.Context(), &knowledgebase_doc_service.DocSegmentListReq{
		UserId:   userId,
		OrgId:    orgId,
		DocId:    req.DocId,
		PageSize: int32(req.PageSize),
		PageNo:   int32(req.PageNo),
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	// 查询ASR和图文问答模型信息
	analyzerModelMap := buildDocAnalyzerModelMap(ctx, resp.DocAnalyzer, resp.AsrModelId, resp.MultimodalModelId, resp.ParserModelId)
	return buildDocSegmentResp(resp, analyzerModelMap), nil
}

func UpdateDocSegmentStatus(ctx *gin.Context, userId, orgId string, r *request.UpdateDocSegmentStatusReq) error {
	_, err := knowledgeBaseDoc.UpdateDocSegmentStatus(ctx.Request.Context(), &knowledgebase_doc_service.UpdateDocSegmentStatusReq{
		UserId:        userId,
		OrgId:         orgId,
		DocId:         r.DocId,
		ContentId:     r.ContentId,
		ContentStatus: r.ContentStatus,
		All:           r.ALL,
	})
	return err
}

func AnalysisDocUrl(ctx *gin.Context, userId, orgId string, r *request.AnalysisUrlDocReq) (*response.AnalysisDocUrlResp, error) {
	resp, err := knowledgeBaseDoc.AnalysisDocUrl(ctx.Request.Context(), &knowledgebase_doc_service.AnalysisUrlDocReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: r.KnowledgeId,
		UrlList:     r.UrlList,
	})
	if err != nil {
		return nil, err
	}
	var urlList []*response.DocUrl
	if len(resp.UrlList) > 0 {
		for _, url := range resp.UrlList {
			urlList = append(urlList, &response.DocUrl{
				Url:      url.Url,
				FileName: url.FileName,
				FileSize: int(url.FileSize),
			})
		}
	}
	return &response.AnalysisDocUrlResp{UrlList: urlList}, nil
}

// buildDocRespList 构造文档返回列表
func buildDocRespList(ctx *gin.Context, dataList []*knowledgebase_doc_service.DocInfo, knowledgeId string) []*response.ListDocResp {
	retList := make([]*response.ListDocResp, 0)
	authorMap := buildAuthorMap(ctx, dataList)
	for _, data := range dataList {
		retList = append(retList, &response.ListDocResp{
			DocId:         data.DocId,
			DocName:       data.DocName,
			DocType:       data.DocType,
			UploadTime:    data.UploadTime,
			Status:        int(data.Status),
			ErrorMsg:      gin_util.I18nKey(ctx, data.ErrorMsg),
			FileSize:      data.DocSize,
			KnowledgeId:   knowledgeId,
			SegmentMethod: data.SegmentMethod,
			Author:        authorMap[data.UserId],
			GraphStatus:   data.GraphStatus,
			GraphErrMsg:   data.GraphErrMsg,
			IsMultimodal:  data.IsMultimodal,
			DocProgress:   data.DocProgress,
			GraphProgress: data.GraphProgress,
		})
	}
	return retList
}

func buildAuthorMap(ctx *gin.Context, dataList []*knowledgebase_doc_service.DocInfo) map[string]string {
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

// buildDocSegmentResp 构造doc分片返回信息
func buildDocSegmentResp(docSegmentListResp *knowledgebase_doc_service.DocSegmentListResp, analyzerModelMap map[string]*response.ModelInfo) *response.DocSegmentResp {
	var segmentContentList = make([]*response.SegmentContent, 0)
	if len(docSegmentListResp.ContentList) > 0 {
		for _, contentInfo := range docSegmentListResp.ContentList {
			var contentLabels = make([]string, 0)
			if len(contentInfo.Labels) > 0 {
				contentLabels = contentInfo.Labels
			}
			segmentContentList = append(segmentContentList, &response.SegmentContent{
				ContentId:  contentInfo.ContentId,
				Content:    contentInfo.Content,
				Available:  contentInfo.Available,
				ContentNum: int(contentInfo.ContentNum),
				Labels:     contentLabels,
				IsParent:   contentInfo.IsParent,
				ChildNum:   int(contentInfo.ChildNum),
			})
		}
	}
	return &response.DocSegmentResp{
		FileName:            docSegmentListResp.FileName,
		PageTotal:           int(docSegmentListResp.PageTotal),
		SegmentTotalNum:     int(docSegmentListResp.SegmentTotalNum),
		MaxSegmentSize:      int(docSegmentListResp.MaxSegmentSize),
		SegmentType:         docSegmentListResp.SegType,
		UploadTime:          docSegmentListResp.CreatedAt,
		Splitter:            docSegmentListResp.Splitter,
		SegmentContentList:  segmentContentList,
		MetaDataList:        buildMetaDataResultList(docSegmentListResp.MetaDataList),
		SegmentImportStatus: docSegmentListResp.SegmentImportStatus,
		SegmentMethod:       docSegmentListResp.SegmentMethod,
		DocAnalyzerText:     buildDocAnalyzerText(docSegmentListResp.DocAnalyzer, analyzerModelMap),
		DownloadUrl:         docSegmentListResp.DownloadUrl,
		CanPreview:          docSegmentListResp.CanPreview,
		PreviewFailReason:   docSegmentListResp.PreviewFailReason,
	}
}

func buildDocAnalyzerText(docAnalyzer []string, analyzerModelMap map[string]*response.ModelInfo) []*response.DocAnalyzerTextInfo {
	if len(docAnalyzer) == 0 {
		return []*response.DocAnalyzerTextInfo{{Text: "无"}}
	}
	return lo.Map(docAnalyzer, func(item string, index int) *response.DocAnalyzerTextInfo {
		info := &response.DocAnalyzerTextInfo{Text: docAnalyzerMap[item]}
		if modelInfo, ok := analyzerModelMap[item]; ok && modelInfo != nil {
			info.DisplayName = modelInfo.DisplayName
			info.Tags = modelInfo.Tags
		}
		return info
	})
}

// buildDocAnalyzerModelMap 查询docAnalyzer中涉及的模型信息，返回 analyzerKey -> ModelInfo 的映射
func buildDocAnalyzerModelMap(ctx *gin.Context, docAnalyzer []string, asrModelId, multimodalModelId, parserModelId string) map[string]*response.ModelInfo {
	modelMap := make(map[string]*response.ModelInfo)
	analyzerSet := lo.SliceToMap(docAnalyzer, func(item string) (string, bool) { return item, true })

	// OCR和模型解析共用parserModelId
	if (analyzerSet["ocr"] || analyzerSet["model"]) && parserModelId != "" {
		modelInfo, err := GetModelById(ctx, &request.GetModelRequest{
			BaseModelRequest: request.BaseModelRequest{ModelId: parserModelId},
		})
		if err != nil {
			log.Warnf("获取解析模型信息失败 modelId=%s err=%v", parserModelId, err)
		} else {
			if analyzerSet["ocr"] {
				modelMap["ocr"] = modelInfo
			}
			if analyzerSet["model"] {
				modelMap["model"] = modelInfo
			}
		}
	}
	if analyzerSet["asr"] && asrModelId != "" {
		modelInfo, err := GetModelById(ctx, &request.GetModelRequest{
			BaseModelRequest: request.BaseModelRequest{ModelId: asrModelId},
		})
		if err != nil {
			log.Warnf("获取ASR模型信息失败 modelId=%s err=%v", asrModelId, err)
		} else {
			modelMap["asr"] = modelInfo
		}
	}
	if analyzerSet["multimodal"] && multimodalModelId != "" {
		modelInfo, err := GetModelById(ctx, &request.GetModelRequest{
			BaseModelRequest: request.BaseModelRequest{ModelId: multimodalModelId},
		})
		if err != nil {
			log.Warnf("获取图文问答模型信息失败 modelId=%s err=%v", multimodalModelId, err)
		} else {
			modelMap["multimodal"] = modelInfo
		}
	}
	return modelMap
}

func buildDocChildSegmentResp(docSegmentListResp *knowledgebase_doc_service.GetDocChildSegmentListResp) *response.DocChildSegmentResp {
	var segmentContentList = make([]*response.ChildSegmentInfo, 0)
	if len(docSegmentListResp.ContentList) > 0 {
		for _, contentInfo := range docSegmentListResp.ContentList {
			segmentContentList = append(segmentContentList, &response.ChildSegmentInfo{
				ChildId:  contentInfo.ChildId,
				Content:  contentInfo.Content,
				ChildNum: int(contentInfo.ChildNum),
				ParentId: contentInfo.ParentId,
			})
		}
	}
	return &response.DocChildSegmentResp{SegmentContentList: segmentContentList}
}

func buildMetaDataList(metaDataList []*request.DocMetaData) []*knowledgebase_doc_service.MetaData {
	if len(metaDataList) == 0 {
		return make([]*knowledgebase_doc_service.MetaData, 0)
	}
	return lo.Map(metaDataList, func(item *request.DocMetaData, index int) *knowledgebase_doc_service.MetaData {
		return &knowledgebase_doc_service.MetaData{
			MetaId:    item.MetaId,
			Key:       item.MetaKey,
			Value:     item.MetaValue,
			Option:    item.Option,
			ValueType: item.MetaValueType,
		}
	})
}

func buildCallbackMetaDataList(metaDataList []*request.CallbackMetaData) []*knowledgebase_doc_service.MetaData {
	if len(metaDataList) == 0 {
		return make([]*knowledgebase_doc_service.MetaData, 0)
	}
	return lo.Map(metaDataList, func(item *request.CallbackMetaData, index int) *knowledgebase_doc_service.MetaData {
		return &knowledgebase_doc_service.MetaData{
			MetaId: item.MetaId,
			Key:    item.Key,
			Value:  item.Value,
		}
	})
}

func buildMetaDataResultList(metaDataList []*knowledgebase_doc_service.MetaData) []*response.DocMetaData {
	if len(metaDataList) == 0 {
		return make([]*response.DocMetaData, 0)
	}
	return lo.Map(metaDataList, func(item *knowledgebase_doc_service.MetaData, index int) *response.DocMetaData {
		return &response.DocMetaData{
			MetaId:        item.MetaId,
			MetaKey:       item.Key,
			MetaValue:     item.Value,
			MetaValueType: item.ValueType,
			MetaRule:      item.Rule,
		}
	})
}

func UpdateDocSegmentLabels(ctx *gin.Context, userId, orgId string, r *request.DocSegmentLabelsReq) error {
	_, err := knowledgeBaseDoc.UpdateDocSegmentLabels(ctx.Request.Context(), &knowledgebase_doc_service.DocSegmentLabelsReq{
		UserId:    userId,
		OrgId:     orgId,
		ContentId: r.ContentId,
		DocId:     r.DocId,
		Labels:    r.Labels,
	})
	return err
}

func CreateDocSegment(ctx *gin.Context, userId, orgId string, r *request.CreateDocSegmentReq) error {
	_, err := knowledgeBaseDoc.CreateDocSegment(ctx.Request.Context(), &knowledgebase_doc_service.CreateDocSegmentReq{
		UserId:  userId,
		OrgId:   orgId,
		DocId:   r.DocId,
		Content: r.Content,
		Labels:  r.Labels,
	})
	return err
}

func BatchCreateDocSegment(ctx *gin.Context, userId, orgId string, r *request.BatchCreateDocSegmentReq) error {
	docUrl, err := minio.GetUploadFileWithExpire(ctx, r.FileUploadId)
	if err != nil {
		log.Errorf("GetUploadFileWithNotExpire error %v", err)
		return grpc_util.ErrorStatus(errs.Code_KnowledgeDocImportUrlFailed)
	}
	ext := filepath.Ext(docUrl)
	if ext != ".csv" {
		return grpc_util.ErrorStatus(errs.Code_KnowledgeDocSegmentFileCSVTypeFail)
	}
	_, err = knowledgeBaseDoc.BatchCreateDocSegment(ctx.Request.Context(), &knowledgebase_doc_service.BatchCreateDocSegmentReq{
		UserId:  userId,
		OrgId:   orgId,
		DocId:   r.DocId,
		FileUrl: docUrl,
	})
	return err
}

func DeleteDocSegment(ctx *gin.Context, userId, orgId string, r *request.DeleteDocSegmentReq) error {
	_, err := knowledgeBaseDoc.DeleteDocSegment(ctx.Request.Context(), &knowledgebase_doc_service.DeleteDocSegmentReq{
		UserId:    userId,
		OrgId:     orgId,
		DocId:     r.DocId,
		ContentId: r.ContentId,
	})
	return err
}

func UpdateDocSegment(ctx *gin.Context, userId, orgId string, r *request.UpdateDocSegmentReq) error {
	_, err := knowledgeBaseDoc.UpdateDocSegment(ctx.Request.Context(), &knowledgebase_doc_service.UpdateDocSegmentReq{
		UserId:    userId,
		OrgId:     orgId,
		DocId:     r.DocId,
		ContentId: r.ContentId,
		Content:   r.Content,
	})
	return err
}

func CreateDocChildSegment(ctx *gin.Context, userId, orgId string, r *request.CreateDocChildSegmentReq) error {
	_, err := knowledgeBaseDoc.CreateDocChildSegment(ctx.Request.Context(), &knowledgebase_doc_service.CreateDocChildSegmentReq{
		UserId:        userId,
		OrgId:         orgId,
		DocId:         r.DocId,
		ParentChunkId: r.ParentId,
		Content:       r.Content,
	})
	return err
}

func UpdateDocChildSegment(ctx *gin.Context, userId, orgId string, r *request.UpdateDocChildSegmentReq) error {
	_, err := knowledgeBaseDoc.UpdateDocChildSegment(ctx.Request.Context(), &knowledgebase_doc_service.UpdateDocChildSegmentReq{
		UserId:        userId,
		OrgId:         orgId,
		DocId:         r.DocId,
		ParentChunkId: r.ParentId,
		ParentChunkNo: r.ParentChunkNo,
		ChildChunk: &knowledgebase_doc_service.ChildChunk{
			ChunkNo: r.ChildChunk.ChildNo,
			Content: r.ChildChunk.Content,
		},
	})
	return err
}

func DeleteDocChildSegment(ctx *gin.Context, userId, orgId string, r *request.DeleteDocChildSegmentReq) error {
	_, err := knowledgeBaseDoc.DeleteDocChildSegment(ctx.Request.Context(), &knowledgebase_doc_service.DeleteDocChildSegmentReq{
		UserId:        userId,
		OrgId:         orgId,
		DocId:         r.DocId,
		ParentChunkId: r.ParentId,
		ParentChunkNo: r.ParentChunkNo,
		ChildChunkNo:  r.ChildChunkNoList,
	})
	return err
}

func GetDocChildSegmentList(ctx *gin.Context, userId, orgId string, req *request.DocChildListReq) (*response.DocChildSegmentResp, error) {
	docSegmentListResp, err := knowledgeBaseDoc.GetDocChildSegmentList(ctx.Request.Context(), &knowledgebase_doc_service.GetDocChildSegmentListReq{
		UserId:    userId,
		OrgId:     orgId,
		DocId:     req.DocId,
		ContentId: req.ContentId,
	})
	if err != nil {
		return nil, err
	}
	return buildDocChildSegmentResp(docSegmentListResp), err
}

// ExportKnowledgeDoc 导出文档
func ExportKnowledgeDoc(ctx *gin.Context, userId, orgId string, req *request.KnowledgeDocExportReq) error {
	_, err := knowledgeBaseDoc.ExportDoc(ctx.Request.Context(), &knowledgebase_doc_service.ExportDocReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: req.KnowledgeId,
		DocIdList:   req.DocIdList,
	})
	if err != nil {
		log.Errorf("导出失败(保存导出任务 失败(%v) ", err)
		return err
	}
	return nil
}

func GetDocUploadLimit(ctx *gin.Context, userId, orgId string, req *request.QueryKnowledgeReq) (*response.DocUploadLimitResp, error) {
	// 1.查询知识库获取emb模型id
	knowledge, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgebase_service.KnowledgeDetailSelectReq{
		KnowledgeId: req.KnowledgeId,
	})
	if err != nil {
		log.Errorf("查询知识库失败(%v) ", err)
		return nil, err
	}
	if knowledge == nil || knowledge.EmbeddingModelInfo == nil {
		log.Errorf("查询知识库失败")
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "knowledge is nil")
	}
	embModelId := knowledge.EmbeddingModelInfo.ModelId
	// 2.获取图片限制大小
	imageSize, err := getEmbImageSize(ctx, userId, orgId, embModelId)
	if err != nil {
		return nil, err
	}
	// 3.获取文件上传后缀
	docUploadLimitResp, err := knowledgeBaseDoc.GetDocUploadLimit(ctx.Request.Context(), nil)
	if err != nil {
		return nil, err
	}
	return buildDocUploadLimitResp(imageSize, docUploadLimitResp), nil
}

func getEmbImageSize(ctx *gin.Context, userId, orgId, embModelId string) (int, error) {
	modelInfo, err := GetModel(ctx, userId, orgId, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{ModelId: embModelId},
	})
	if err != nil {
		log.Errorf("查询模型失败(%v) ", err)
		return 0, err
	}
	// 校验模型类型
	if modelInfo.ModelType != mp.ModelTypeMultiEmbedding {
		return 0, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "modelType mismatch")
	}
	// 模型配置断言，通过 JSON 序列化转换为 map
	configBytes, err := json.Marshal(modelInfo.Config)
	if err != nil {
		return 0, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "embedding模型配置错误")
	}
	var modelConfig map[string]interface{}
	if err = json.Unmarshal(configBytes, &modelConfig); err != nil {
		return 0, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "embedding模型配置错误")
	}
	maxImageSizeVal, ok := modelConfig["maxImageSize"]
	if !ok || maxImageSizeVal == nil {
		return 0, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "embedding模型配置错误")
	}
	size, ok := maxImageSizeVal.(float64)
	if !ok {
		return 0, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "embedding模型配置错误")
	}
	return int(int64(size)), nil
}

func buildDocUploadLimitResp(imageSize int, docUploadLimitResp *knowledgebase_doc_service.DocUploadLimitResp) *response.DocUploadLimitResp {
	retList := make([]*response.DocUploadLimit, 0)
	var maxSize int
	for _, file := range docUploadLimitResp.List {
		switch file.FileType {
		case "image":
			maxSize = imageSize
		case "video":
			maxSize = 100
		default:
			maxSize = 0
		}
		retList = append(retList, &response.DocUploadLimit{
			ExtList:  file.ExtList,
			FileType: file.FileType,
			MaxSize:  maxSize,
		})
	}
	return &response.DocUploadLimitResp{
		UploadLimitList: retList,
	}
}

func buildMetaInfoList(req *request.DocImportReq) []*knowledgebase_doc_service.DocMetaData {
	var metaList []*knowledgebase_doc_service.DocMetaData
	for _, meta := range req.DocMetaData {
		metaList = append(metaList, &knowledgebase_doc_service.DocMetaData{
			Key:       meta.MetaKey,
			Value:     meta.MetaValue,
			ValueType: meta.MetaValueType,
			Rule:      meta.MetaRule,
		})
	}
	return metaList
}

func buildDocInfoList(ctx *gin.Context, docList []*request.DocInfo) ([]*knowledgebase_doc_service.DocFileInfo, error) {
	if len(docList) == 0 {
		return make([]*knowledgebase_doc_service.DocFileInfo, 0), nil
	}
	var docInfoList []*knowledgebase_doc_service.DocFileInfo
	for _, info := range docList {
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
		docInfoList = append(docInfoList, &knowledgebase_doc_service.DocFileInfo{
			DocName: info.DocName,
			DocId:   info.DocId,
			DocUrl:  docUrl,
			DocType: docType,
			DocSize: info.DocSize,
		})
	}
	return docInfoList, nil
}

func buildDocSegment(docSegment *knowledgebase_doc_service.DocSegment) *response.DocSegment {
	if docSegment.SegmentType == AutoSegment {
		return &response.DocSegment{
			SegmentType:   docSegment.SegmentType,
			SegmentMethod: docSegment.SegmentMethod,
		}
	}
	maxSubMaxSplitter := int(docSegment.MaxSplitter)
	var subMaxSplitter *int
	var subSplitter []string
	if docSegment.SegmentMethod == ParentChildSegment {
		subMaxSplitterValue := int(docSegment.SubMaxSplitter)
		subMaxSplitter = &subMaxSplitterValue
		subSplitter = docSegment.SubSplitter
	}
	return &response.DocSegment{
		SegmentType:    docSegment.SegmentType,
		Splitter:       docSegment.Splitter,
		MaxSplitter:    &maxSubMaxSplitter,
		Overlap:        &docSegment.Overlap,
		SegmentMethod:  docSegment.SegmentMethod,
		SubMaxSplitter: subMaxSplitter,
		SubSplitter:    subSplitter,
	}
}

func buildKeywordsInfo(keywords []*knowledgebase_doc_service.KeywordsInfo) []*response.KeywordsInfo {
	retList := make([]*response.KeywordsInfo, 0)
	if len(keywords) > 0 {
		for _, v := range keywords {
			keyword := &response.KeywordsInfo{
				Id:                 v.Id,
				Name:               v.Name,
				Alias:              v.Alias,
				KnowledgeBaseIds:   v.KnowledgeBaseIds,
				KnowledgeBaseNames: v.KnowledgeBaseNames,
				UpdatedAt:          v.UpdatedAt,
			}
			retList = append(retList, keyword)
		}
	}
	return retList
}
