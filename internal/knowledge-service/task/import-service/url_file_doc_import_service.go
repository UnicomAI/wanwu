package import_service

import (
	"context"
	"errors"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// UrlFileDocImportService url文件上传已下架，保留注册仅为让残留的异步任务走正常失败流程，不再执行解析
type UrlFileDocImportService struct{}

var urlFileDocImportService = &UrlFileDocImportService{}

func init() {
	AddDocImportService(urlFileDocImportService)
}

func (f UrlFileDocImportService) ImportType() int {
	return model.UrlFileImportType
}

func (f UrlFileDocImportService) AnalyzeDoc(ctx context.Context, importTask *model.KnowledgeImportTask, importDocInfo *model.DocImportInfo) ([]*model.DocInfo, error) {
	// 任务级 error_msg 直接展示给用户，不走 i18n key（见 bff GetDocImportTip）
	log.Errorf("url file import offline, skip analyze task %s", importTask.ImportId)
	return nil, errors.New("url导入方式已下架，该文档不再解析")
}

// 原 url 文件解析实现（下载 excel、逐行调 rag 解析 url）已下架，保留备查
// func (f UrlFileDocImportService) analyzeDoc(ctx context.Context, importTask *model.KnowledgeImportTask, importDocInfo *model.DocImportInfo) ([]*model.DocInfo, error) {
// 	docInfo := importDocInfo.DocInfoList[0] // 一个excel 文件
// 	//1.下载压缩文件到本地
// 	var localFilePath = util.BuildFilePath(config.GetConfig().KnowledgeDocConfig.DocLocalFilePath, docInfo.DocType)
// 	err := service.DownloadFileToLocal(ctx, docInfo.DocUrl, localFilePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer func() {
// 		log.Infof("remove local file %s", localFilePath)
// 		err1 := os.Remove(localFilePath)
// 		if err1 != nil {
// 			log.Errorf("DoFileExtract local file delete %v", err)
// 		}
// 	}()
// 	//2.读取excel
// 	columnList, err := util.ReadExcelColumn(localFilePath, 1, 1)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(columnList) == 0 {
// 		return nil, errors.New("解析url失败")
// 	}
// 	urlCountLimit := config.GetConfig().UsageLimit.UrlCountLimit
// 	if urlCountLimit <= 0 || len(columnList) < urlCountLimit {
// 		urlCountLimit = len(columnList) // 如果数组长度小于限制，取实际长度
// 	}
// 	columnList = columnList[:urlCountLimit]
// 	//3.执行文档解析
// 	docUrlRespList, err := service.BatchRagDocUrlAnalysis(ctx, columnList)
// 	if err != nil {
// 		return nil, err
// 	}
// 	//4.转换结果
// 	var retList []*model.DocInfo
// 	for _, docUrlREsp := range docUrlRespList {
// 		retList = append(retList, &model.DocInfo{
// 			DocName:     docUrlREsp.FileName,
// 			DocSize:     int64(docUrlREsp.FileSize),
// 			DocUrl:      docUrlREsp.Url,
// 			FilePathMd5: wanwu_util.MD5([]byte(docUrlREsp.Url)),
// 		})
// 	}
// 	return retList, nil
// }

func (f UrlFileDocImportService) CheckDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*model.DocInfo) ([]*CheckFileResult, error) {
	return urlDocImportService.CheckDoc(ctx, importTask, docList)
}

func (f UrlFileDocImportService) ImportDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*CheckFileResult) ([]*model.DocInfo, error) {
	return urlDocImportService.ImportDoc(ctx, importTask, docList)
}
