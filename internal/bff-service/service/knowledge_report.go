package service

import (
	"path/filepath"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_report_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-report-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/gin-gonic/gin"
)

func GetKnowledgeReport(ctx *gin.Context, userId, orgId string, req *request.KnowledgeReportSelectReq) (*response.KnowledgeReportPageResult, error) {
	resp, err := knowledgeBaseReport.GetKnowledgeReport(ctx.Request.Context(), &knowledgebase_report_service.GetReportReq{
		KnowledgeInfo: &knowledgebase_report_service.ReportIdentity{
			KnowledgeId: req.KnowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		},
		PageSize: int32(req.PageSize),
		PageNum:  int32(req.PageNo),
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeReportList(req, resp), nil
}

func GenerateKnowledgeReport(ctx *gin.Context, userId, orgId string, req *request.KnowledgeReportGenerateReq) error {
	_, err := knowledgeBaseReport.GenerateKnowledgeReport(ctx.Request.Context(), &knowledgebase_report_service.ReportIdentity{
		KnowledgeId: req.KnowledgeId,
		UserId:      userId,
		OrgId:       orgId,
	})
	return err
}

func DeleteKnowledgeReport(ctx *gin.Context, userId, orgId string, req *request.KnowledgeReportDeleteReq) error {
	_, err := knowledgeBaseReport.DeleteKnowledgeReport(ctx.Request.Context(), &knowledgebase_report_service.DeleteReportReq{
		KnowledgeInfo: &knowledgebase_report_service.ReportIdentity{
			KnowledgeId: req.KnowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		},
		ContentId: req.ContentId,
	})
	return err
}

func UpdateKnowledgeReport(ctx *gin.Context, userId, orgId string, req *request.KnowledgeReportUpdateReq) error {
	_, err := knowledgeBaseReport.UpdateKnowledgeReport(ctx.Request.Context(), &knowledgebase_report_service.UpdateReportReq{
		KnowledgeInfo: &knowledgebase_report_service.ReportIdentity{
			KnowledgeId: req.KnowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		},
		Data: &knowledgebase_report_service.ReportInfo{
			ContentId: req.ContentId,
			Title:     req.Title,
			Content:   req.Content,
		},
	})
	return err
}

func AddKnowledgeReport(ctx *gin.Context, userId, orgId string, req *request.KnowledgeReportAddReq) error {
	_, err := knowledgeBaseReport.AddKnowledgeReport(ctx.Request.Context(), &knowledgebase_report_service.AddReportReq{
		KnowledgeInfo: &knowledgebase_report_service.ReportIdentity{
			KnowledgeId: req.KnowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		},
		Title:   req.Title,
		Content: req.Content,
	})
	return err
}

func BatchAddKnowledgeReport(ctx *gin.Context, userId, orgId string, req *request.KnowledgeReportBatchAddReq) error {
	docUrl, err := minio.GetUploadFileWithExpire(ctx, req.FileUploadId)
	if err != nil {
		log.Errorf("GetUploadFileWithNotExpire error %v", err)
		return grpc_util.ErrorStatus(errs.Code_KnowledgeDocImportUrlFailed)
	}
	ext := filepath.Ext(docUrl)
	if ext != ".csv" {
		return grpc_util.ErrorStatus(errs.Code_KnowledgeDocSegmentFileCSVTypeFail)
	}
	_, err = knowledgeBaseReport.BatchAddKnowledgeReport(ctx.Request.Context(), &knowledgebase_report_service.BatchAddKnowledgeReportReq{
		KnowledgeInfo: &knowledgebase_report_service.ReportIdentity{
			KnowledgeId: req.KnowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		},
		FileUrl: docUrl,
	})
	return err
}

func buildKnowledgeReportList(req *request.KnowledgeReportSelectReq, resp *knowledgebase_report_service.GetReportResp) *response.KnowledgeReportPageResult {
	retList := make([]*response.KnowledgeReportInfo, 0)
	for _, v := range resp.List {
		retList = append(retList, &response.KnowledgeReportInfo{
			Content:   v.Content,
			ContentId: v.ContentId,
			Title:     v.Title,
		})
	}

	return &response.KnowledgeReportPageResult{
		List:             retList,
		Total:            resp.Total,
		PageNo:           req.PageNo,
		PageSize:         req.PageSize,
		CreatedAt:        resp.CreatedAt,
		Status:           resp.Status,
		CanGenerate:      resp.CanGenerate,
		CanAddReport:     resp.CanAddReport,
		GenerateLabel:    resp.GenerateLabel,
		LastImportStatus: resp.LastImportStatus,
	}
}
