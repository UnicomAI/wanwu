package import_service

import (
	"context"
	"os"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type UrlFileDocImportService struct{}

var urlFileDocImportService = &UrlFileDocImportService{}

func init() {
	AddDocImportService(urlFileDocImportService)
}

func (f UrlFileDocImportService) ImportType() int {
	return model.UrlFileImportType
}

func (f UrlFileDocImportService) AnalyzeDoc(ctx context.Context, importTask *model.KnowledgeImportTask, importDocInfo *model.DocImportInfo) ([]*model.DocInfo, error) {
	docInfo := importDocInfo.DocInfoList[0] // an excel file
	//1. Download the compressed file to local
	var localFilePath = util.BuildFilePath(config.GetConfig().KnowledgeDocConfig.DocLocalFilePath, docInfo.DocType)
	err := service.DownloadFileToLocal(ctx, docInfo.DocUrl, localFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		log.Infof("remove local file %s", localFilePath)
		err1 := os.Remove(localFilePath)
		if err1 != nil {
			log.Errorf("DoFileExtract local file delete %v", err)
		}
	}()
	//2.Read excel
	columnList, err := util.ReadExcelColumn(localFilePath, 1)
	if err != nil {
		return nil, err
	}
	//3. Perform document parsing
	docUrlRespList, err := service.BatchRagDocUrlAnalysis(ctx, columnList)
	if err != nil {
		return nil, err
	}
	//4. Conversion results
	var retList []*model.DocInfo
	for _, docUrlREsp := range docUrlRespList {
		retList = append(retList, &model.DocInfo{
			DocName: docUrlREsp.FileName,
			DocSize: int64(docUrlREsp.FileSize),
			DocUrl:  docUrlREsp.Url,
		})
	}
	return retList, nil
}

func (f UrlFileDocImportService) CheckDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*model.DocInfo) ([]*CheckFileResult, error) {
	return urlDocImportService.CheckDoc(ctx, importTask, docList)
}

func (f UrlFileDocImportService) ImportDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*CheckFileResult) ([]*model.DocInfo, error) {
	return urlDocImportService.ImportDoc(ctx, importTask, docList)
}
