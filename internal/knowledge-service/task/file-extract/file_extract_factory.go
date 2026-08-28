package file_extract

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type DocFileInfo struct {
	DocInfo      *model.DocInfo
	DocLocalPath string
}

var fileExtractServiceMap = make(map[string]FileExtractService)

func AddFileExtractService(service FileExtractService) {
	fileExtractServiceMap[service.ExtractFileType()] = service
}

// DoFileExtract 执行文件导入
func DoFileExtract(ctx context.Context, doc *model.DocInfo) (docList []*model.DocInfo, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			var errMsg = fmt.Sprintf("do doc import task panic: %v", err1)
			log.Errorf(errMsg)
			err = errors.New(errMsg)
		}
	}()
	//1.获取服务service
	fileExtractService, ok := fileExtractServiceMap[doc.DocType]
	if !ok {
		log.Errorf("DoFileExtract not found doc type %s", doc.DocType)
		//没找到处理器不算处理错误
		return nil, errors.New("DoFileExtract not found doc type")
	}
	//2.下载压缩文件到本地
	var localFilePath = util.BuildFilePath(config.GetConfig().KnowledgeDocConfig.DocLocalFilePath, doc.DocType)
	err = service.DownloadFileToLocal(ctx, doc.DocUrl, localFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err1 := os.Remove(localFilePath)
		if err != nil {
			log.Errorf("DoFileExtract local file delete %v", err1)
		}
	}()
	//3.执行压缩文件解压
	var destDir = util.ReplaceLast(localFilePath, doc.DocType, "") + "-dir"
	extractDir, err := fileExtractService.ExtractFile(ctx, localFilePath, destDir)
	if err != nil {
		return nil, err
	}
	//4.执行文件上传
	return uploadFile(ctx, extractDir, destDir)
}

// uploadFile 上传文件到minio
func uploadFile(ctx context.Context, extractDir, destDir string) (docList []*model.DocInfo, err error) {
	defer func() {
		log.Infof("remove file %s ", extractDir)
		err1 := os.RemoveAll(extractDir)
		if err1 != nil {
			log.Errorf("uploadFile remove path %s error %v", extractDir, err)
		}
	}()
	var fileList []*DocFileInfo // 存储解压后所有文件路径的切片
	// 使用filepath.Walk遍历目录
	err = filepath.Walk(extractDir, walkFunc(destDir, &fileList))
	if err != nil {
		log.Infof("upload file to %s", err)
		return nil, err
	}
	fileLen := len(fileList)
	if fileLen == 0 {
		return nil, errors.New("无法解析出可用文档")
	}
	//压缩文件总量截取
	limitNum := int(config.GetConfig().UsageLimit.MaxNumberOfFilesInCompressed)
	if limitNum > 0 && fileLen > limitNum {
		fileList = fileList[0:limitNum]
	}
	//解压后文件大小限制
	//dir := config.GetConfig().Minio.KnowledgeDir
	bucketName := config.GetConfig().Minio.Bucket
	//循环遍历所有文件,上传minio
	for _, fileInfo := range fileList {
		//上传文档
		_, minioFilePath, _, err := service.UploadLocalFile(ctx, "", bucketName, util.BuildFilePath("", fileInfo.DocInfo.DocType), fileInfo.DocLocalPath)
		if err != nil {
			log.Errorf("upload file err: %v", err)
			continue
		}
		fileInfo.DocInfo.DocUrl = minioFilePath
		docList = append(docList, fileInfo.DocInfo)
	}
	return docList, nil
}

//// checkFile 校验压缩文件，压缩后的文件大小限制过滤
//func checkFile(docType string, docSize int64) bool {
//	fileTypeList := strings.Split(config.GetConfig().UsageLimit.FileTypes, ";")
//	for _, suffix := range fileTypeList {
//		if suffix == "" || strings.TrimSpace(suffix) == "" {
//			continue
//		}
//		if docType == suffix {
//			//压缩文件大小校验
//			limitSize := config.GetConfig().UsageLimit.FileSizeLimit
//			if limitSize != -1 && docSize > limitSize {
//				return false
//			}
//			return true
//		}
//	}
//	return false
//}

// 使用闭包来捕获files变量，配合filepath.Walk来使用
func walkFunc(destDir string, files *[]*DocFileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf(fmt.Sprintf("walkFunc error accessing path %s file: %v", path, err)) // 处理错误
			return err                                                                      // 返回nil以继续遍历
		}
		log.Infof("walkFunc accessing destDir %s path %s ", destDir, path)
		if !info.IsDir() {
			*files = append(*files, &DocFileInfo{
				DocLocalPath: path,
				DocInfo: &model.DocInfo{
					DocName:     info.Name(),
					DocType:     filepath.Ext(info.Name()),
					DocSize:     info.Size(),
					DirFilePath: strings.TrimPrefix(path, destDir),
				},
			}) // 添加文件路径到切片
		}
		return nil // 返回nil表示没有错误，可以继续遍历
	}
}
