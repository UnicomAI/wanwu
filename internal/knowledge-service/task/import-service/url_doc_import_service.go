package import_service

import (
	"context"
	"errors"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
)

const (
	UrlFileType = "url"
)

type UrlDocImportService struct{}

var urlDocImportService = &UrlDocImportService{}

func init() {
	AddDocImportService(urlDocImportService)
}

func (f UrlDocImportService) ImportType() int {
	return model.UrlImportType
}

func (f UrlDocImportService) AnalyzeDoc(ctx context.Context, importTask *model.KnowledgeImportTask, importDocInfo *model.DocImportInfo) ([]*model.DocInfo, error) {
	return importDocInfo.DocInfoList, nil
}

func (f UrlDocImportService) CheckDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*model.DocInfo) ([]*CheckFileResult, error) {
	var resultList []*CheckFileResult
	var urlRepeatMap = make(map[string]bool)
	for _, docInfo := range docList {
		if urlRepeatMap[docInfo.FilePathMd5] { //如果统一文件内存在同一url则直接失败
			resultList = append(resultList, &CheckFileResult{
				Status:     model.DocFail,
				ErrMessage: util.KnowledgeImportSameNameErr,
				DocInfo:    docInfo,
			})
			continue
		}
		//文档重名校验
		checkResult, checkMessage := checkUrlFile(ctx, importTask.UserId, importTask.KnowledgeId, docInfo.DocUrl)
		var status = model.DocInit
		if !checkResult {
			status = model.DocFail
		}
		resultList = append(resultList, &CheckFileResult{
			Status:     status,
			ErrMessage: checkMessage,
			DocInfo:    docInfo,
		})
		if len(docInfo.FilePathMd5) > 0 {
			urlRepeatMap[docInfo.FilePathMd5] = true
		}
	}
	return resultList, nil
}

func (f UrlDocImportService) ImportDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*CheckFileResult) ([]*model.DocInfo, error) {
	var result = false
	var retList []*model.DocInfo
	for _, docInfo := range docList {
		err := orm.CreateKnowledgeUrlDoc(ctx, buildKnowledgeUrlDoc(importTask, docInfo), importTask)
		if err != nil {
			log.Errorf("import doc fail %v", err)
			continue
		}
		result = true
		retList = append(retList, docInfo.DocInfo)
	}
	if !result {
		log.Errorf("import doc fail non success")
		return nil, errors.New("import fail")
	}
	return retList, nil
}

func checkUrlFile(ctx context.Context, userId string, knowledgeId string, docUrl string) (bool, string) {
	err := orm.CheckKnowledgeDocSameName(ctx, userId, knowledgeId, "", docUrl, "")
	if err != nil {
		log.Errorf("文件 '%s' 判断文档重名失败(%v)", docUrl, err)
		return false, util.KnowledgeImportSameNameErr
	}
	return true, ""
}

func buildKnowledgeUrlDoc(importTask *model.KnowledgeImportTask, docInfo *CheckFileResult) *model.KnowledgeDoc {
	var fileSize = docInfo.DocInfo.DocSize
	if docInfo.DocInfo.DocSize == 0 {
		fileSize = 10 // 10b,经bff转换后为0.01kb
	}
	return &model.KnowledgeDoc{
		DocId:        wanwu_util.NewID(),
		ImportTaskId: importTask.ImportId,
		KnowledgeId:  importTask.KnowledgeId,
		FilePath:     docInfo.DocInfo.DocUrl,
		FilePathMd5:  util.MD5(docInfo.DocInfo.DocUrl),
		Name:         docInfo.DocInfo.DocName,
		Status:       docInfo.Status,
		ErrorMsg:     db.LongText(docInfo.ErrMessage),
		FileType:     UrlFileType,
		FileSize:     fileSize,
		CreatedAt:    time.Now().UnixMilli(),
		UpdatedAt:    time.Now().UnixMilli(),
		UserId:       importTask.UserId,
		OrgId:        importTask.OrgId,
	}
}
