package service

import (
	"path/filepath"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgebase_doc_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-doc-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// GetDocList 查询知识库所属文档列表 [EN] GetDocList Query the list of documents belonging to the knowledge base
func GetDocList(ctx *gin.Context, userId, orgId string, r *request.DocListReq) (*response.DocPageResult, error) {
	resp, err := knowledgeBaseDoc.GetDocList(ctx.Request.Context(), &knowledgebase_doc_service.GetDocListReq{
		KnowledgeId: r.KnowledgeId,
		DocName:     r.DocName,
		Status:      int32(r.Status),
		PageSize:    int32(r.PageSize),
		PageNum:     int32(r.PageNo),
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		return nil, err
	}
	knowledgeInfo := resp.KnowledgeInfo
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
		},
	}, nil
}

// GetDocDetail 查询知识库所属文档详情 [EN] GetDocDetail queries the document details of the knowledge base
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
		FileSize:      util.ToFileSizeStr(data.DocSize),
		KnowledgeId:   data.KnowledgeId,
		SegmentMethod: data.SegmentMethod,
	}, nil
}

// ImportDoc 导入文档 [EN] ImportDoc import document
func ImportDoc(ctx *gin.Context, userId, orgId string, req *request.DocImportReq) error {
	segment := req.DocSegment
	docInfoList, err := buildDocInfoList(ctx, req)
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
		DocAnalyzer:     req.DocAnalyzer,
		DocInfoList:     docInfoList,
		OcrModelId:      req.ParserModelId,
		DocPreprocess:   req.DocPreprocess,
		DocMetaDataList: buildMetaInfoList(req),
	})
	if err != nil {
		log.Errorf("上传失败(保存上传任务 失败(%v) ", err)
		return err
	}
	return nil
}

// UpdateDocMetaData 更新文档元数据 [EN] UpdateDocMetaData updates document metadata
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

// BatchUpdateDocMetaData 批量文档元数据 [EN] BatchUpdateDocMetaData batch document metadata
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
	_, err := knowledgeBaseDoc.InitDocStatus(ctx, &knowledgebase_doc_service.InitDocStatusReq{
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
		Ids:    r.DocIdList,
		UserId: userId,
		OrgId:  orgId,
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
	})
	if err != nil {
		return nil, err
	}
	return buildDocSegmentResp(resp), nil
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

// buildDocRespList 构造文档返回列表 [EN] buildDocRespList constructs a document return list
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
			FileSize:      util.ToFileSizeStr(data.DocSize),
			KnowledgeId:   knowledgeId,
			SegmentMethod: data.SegmentMethod,
			Author:        authorMap[data.UserId],
			GraphStatus:   data.GraphStatus,
			GraphErrMsg:   data.GraphErrMsg,
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
	userInfoList, err := iam.GetUserSelectByUserIDs(ctx, &iam_service.GetUserSelectByUserIDsReq{
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

// buildDocSegmentResp 构造doc分片返回信息 [EN] buildDocSegmentResp constructs doc segmentation and returns information
func buildDocSegmentResp(docSegmentListResp *knowledgebase_doc_service.DocSegmentListResp) *response.DocSegmentResp {
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
	}
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

func buildDocInfoList(ctx *gin.Context, req *request.DocImportReq) ([]*knowledgebase_doc_service.DocFileInfo, error) {
	var docInfoList []*knowledgebase_doc_service.DocFileInfo
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
			//特殊处理类型 [EN] special processing type
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
