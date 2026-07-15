package import_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	file_extract "github.com/UnicomAI/wanwu/internal/knowledge-service/task/file-extract"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
)

type FileDocImportService struct{}

var fileDocImportService = &FileDocImportService{}

func init() {
	AddDocImportService(fileDocImportService)
}

func (f FileDocImportService) ImportType() int {
	return model.FileImportType
}

func (f FileDocImportService) AnalyzeDoc(ctx context.Context, importTask *model.KnowledgeImportTask, importDocInfo *model.DocImportInfo) ([]*model.DocInfo, error) {
	var docFileList []*model.DocInfo
	for _, docInfo := range importDocInfo.DocInfoList {
		isCompressed, err := checkCompressedFile(docInfo)
		log.Infof("AnalyzeDoc %v, %v, err %v", docInfo, isCompressed, err)
		if err != nil {
			docFileList = append(docFileList, docInfo)
			continue
		}
		docList, err := buildDocList(ctx, isCompressed, docInfo)
		if err != nil {
			return nil, err
		}
		if len(docList) > 0 {
			docFileList = append(docFileList, docList...)
		}
	}
	return docFileList, nil
}

func (f FileDocImportService) CheckDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*model.DocInfo) ([]*CheckFileResult, error) {
	var resultList []*CheckFileResult
	knowledge, err := orm.SelectKnowledgeById(ctx, importTask.KnowledgeId, "", "")
	var multimodal = false
	if err != nil {
		//查询失败不影响主流程
		log.Errorf("check doc select knowledge error %v", err)
	} else {
		multimodal = knowledge.Category == model.CategoryMultimodal
	}
	fileTypeMap := BuildFileTypeMap(multimodal)
	for _, docInfo := range docList {
		checkResult, checkMessage := checkOneFile(ctx, importTask, docInfo, fileTypeMap)
		var status = model.DocInit
		if !checkResult {
			status = model.DocFail
		}
		resultList = append(resultList, &CheckFileResult{
			Status:     status,
			ErrMessage: checkMessage,
			DocInfo:    docInfo,
		})
	}
	return resultList, nil
}

func (f FileDocImportService) ImportDoc(ctx context.Context, importTask *model.KnowledgeImportTask, docList []*CheckFileResult) ([]*model.DocInfo, error) {
	var result = false
	var retList []*model.DocInfo
	for _, docInfo := range docList {
		err := orm.CreateKnowledgeDoc(ctx, buildKnowledgeDoc(importTask, docInfo), importTask)
		if err != nil {
			log.Errorf("import doc fail %v", err)
			continue
		}
		result = true
		retList = append(retList, docInfo.DocInfo)
	}
	if !result {
		log.Errorf("import doc fail success")
		return nil, errors.New("import fail")
	}
	return retList, nil
}

// checkOneFile 单个文件校验
func checkOneFile(ctx context.Context, importTask *model.KnowledgeImportTask, doc *model.DocInfo, fileTypeMap map[string]bool) (bool, string) {
	//1.文件类型校验
	if !fileTypeMap[doc.DocType] {
		log.Errorf("文件%s格式%s不支持", doc.DocName, doc.DocType)
		return false, util.KnowledgeImportFileFormatErr
	}
	//2.文件大小校验
	err := CheckSingleFileSize(doc)
	if err != nil {
		log.Errorf("文件 '%s' 大小超过限制(%v)", doc.DocName, err)
		return false, util.KnowledgeImportFileSizeErr
	}
	//3.文件名合法性校验
	if !IsSafeFileName(doc.DocName) {
		log.Errorf("文件 '%s' 文件名非法", doc.DocName)
		return false, util.KnowledgeImportInvalidNameErr
	}
	//3.1 文件名长度校验：RAG 落盘时以文档名为本地文件名，过长会触发系统"文件名过长"报错
	if !wanwu_util.ValidateFileNameLength(doc.DocName) {
		log.Errorf("文件 '%s' 文件名过长", doc.DocName)
		return false, util.KnowledgeImportNameTooLongErr
	}
	//4.文档重名校验
	err = orm.CheckKnowledgeDocSameName(ctx, importTask.UserId, importTask.KnowledgeId, doc.DocName, "", "")
	if err != nil {
		log.Errorf("文件 '%s' 判断文档重名失败(%v)", doc.DocName, err)
		return false, util.KnowledgeImportSameNameErr
	}
	return true, ""
}

// IsSafeFileName 校验文件名是否合法
// 文件名中不能包含斜杠 /、反斜杠 \、或连续两个点 ..
func IsSafeFileName(name string) bool {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	if strings.Contains(name, "..") {
		return false
	}
	return true
}

// 校验单个文件大小限制
func CheckSingleFileSize(doc *model.DocInfo) error {
	limitConfig := config.GetConfig().UsageLimit
	var fileLimit int64
	switch doc.DocType {
	case ".docx", ".doc":
		fileLimit = limitConfig.DocxSizeLimit
	case ".txt":
		fileLimit = limitConfig.TxtSizeLimit
	case ".pdf":
		fileLimit = limitConfig.PdfSizeLimit
	case ".xlsx":
		fileLimit = limitConfig.ExcelSizeLimit
	case ".csv":
		fileLimit = limitConfig.CsvSizeLimit
	case ".pptx":
		fileLimit = limitConfig.PptxSizeLimit
	case ".html":
		fileLimit = limitConfig.HtmlSizeLimit
	case ".tar.gz":
		fileLimit = limitConfig.CompressedSizeLimit
	case ".zip":
		fileLimit = limitConfig.CompressedSizeLimit

	default:
		fileLimit = limitConfig.MaxFileSize
	}
	return checkOneTypeFile(fileLimit, doc.DocSize, doc.DocName, doc.DocType)
}

func checkOneTypeFile(fileLimit int64, fileSize int64, fileName string, fileType string) error {
	if fileLimit != -1 && fileSize > fileLimit {
		return fmt.Errorf("%s文件'%s'大小%d超过限制%d", fileType, fileName, fileSize, fileLimit)
	}
	return nil
}

// checkCompressedFile 校验是否是压缩文件
func checkCompressedFile(doc *model.DocInfo) (bool, error) {
	compressedFileTypeList := strings.Split(config.GetConfig().UsageLimit.CompressedFileType, ";")
	limitSize := config.GetConfig().UsageLimit.CompressedSizeLimit
	for _, suffix := range compressedFileTypeList {
		if suffix == "" || strings.TrimSpace(suffix) == "" {
			continue
		}
		if doc.DocType == suffix {
			err := checkOneTypeFile(limitSize, doc.DocSize, doc.DocName, doc.DocType)
			return true, err
		}
	}
	return false, nil
}

// buildDocList 构造文档列表
func buildDocList(ctx context.Context, isCompress bool, file *model.DocInfo) ([]*model.DocInfo, error) {
	var docList []*model.DocInfo
	if !isCompress {
		copyFile, _, _, err := service.CopyFile(ctx, file.DocUrl, "", false)
		if err != nil {
			return nil, err
		}
		file.DocUrl = copyFile
		docList = append(docList, file)
		return docList, nil
	}
	//执行文件解压
	list, err := file_extract.DoFileExtract(ctx, file)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		docList = append(docList, list...)
	}
	return docList, nil
}

func buildKnowledgeDoc(importTask *model.KnowledgeImportTask, checkFileResult *CheckFileResult) *model.KnowledgeDoc {
	docInfo := checkFileResult.DocInfo
	return &model.KnowledgeDoc{
		DocId:        wanwu_util.NewID(),
		ImportTaskId: importTask.ImportId,
		KnowledgeId:  importTask.KnowledgeId,
		FilePath:     docInfo.DocUrl,
		FilePathMd5:  wanwu_util.MD5([]byte(docInfo.DocUrl)),
		DirFilePath:  docInfo.DirFilePath,
		Name:         docInfo.DocName,
		FileType:     docInfo.DocType,
		FileSize:     docInfo.DocSize,
		CreatedAt:    time.Now().UnixMilli(),
		UpdatedAt:    time.Now().UnixMilli(),
		UserId:       importTask.UserId,
		OrgId:        importTask.OrgId,
		Status:       checkFileResult.Status,
		ErrorMsg:     checkFileResult.ErrMessage,
	}
}

func BuildFileTypeMap(multiModal bool) map[string]bool {
	var fileTypeMap = make(map[string]bool)
	var filterTypeMap = make(map[string]bool)
	//非多模态过滤对应类型
	if !multiModal {
		fileTypes := strings.Split(config.GetConfig().UsageLimit.AudioTypes, ";")
		for _, fileType := range fileTypes {
			filterTypeMap["."+fileType] = true
		}
		fileTypes = strings.Split(config.GetConfig().UsageLimit.VideoTypes, ";")
		for _, fileType := range fileTypes {
			filterTypeMap["."+fileType] = true
		}
		fileTypes = strings.Split(config.GetConfig().UsageLimit.ImageTypes, ";")
		for _, fileType := range fileTypes {
			filterTypeMap["."+fileType] = true
		}
	}
	fileTypes := strings.Split(config.GetConfig().UsageLimit.FileTypes, ";")
	for _, fileType := range fileTypes {
		if filterTypeMap[fileType] {
			continue
		}
		fileTypeMap[fileType] = true
	}
	compressedFileTypeList := strings.Split(config.GetConfig().UsageLimit.CompressedFileType, ";")
	for _, fileType := range compressedFileTypeList {
		fileTypeMap[fileType] = true
	}
	return fileTypeMap
}
