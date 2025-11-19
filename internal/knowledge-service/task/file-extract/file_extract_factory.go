package file_extract

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

// DoFileExtract 执行文件导入 [EN] DoFileExtract executes file import
func DoFileExtract(ctx context.Context, doc *model.DocInfo) (docList []*model.DocInfo, err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			var errMsg = fmt.Sprintf("do doc import task panic: %v", err1)
			log.Errorf(errMsg)
			err = errors.New(errMsg)
		}
	}()
	//1.获取服务service [EN] 1. Get service service
	fileExtractService, ok := fileExtractServiceMap[doc.DocType]
	if !ok {
		log.Errorf("DoFileExtract not found doc type %s", doc.DocType)
		//没找到处理器不算处理错误 [EN] Not finding the processor does not count as a processing error
		return nil, errors.New("DoFileExtract not found doc type")
	}
	//2.下载压缩文件到本地 [EN] 2. Download the compressed file to local
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
	//3.执行压缩文件解压 [EN] 3. Execute compressed file decompression
	var destDir = util.ReplaceLast(localFilePath, doc.DocType, "") + "-dir"
	extractDir, err := fileExtractService.ExtractFile(ctx, localFilePath, destDir)
	if err != nil {
		return nil, err
	}
	//4.执行文件上传 [EN] 4. Perform file upload
	return uploadFile(ctx, extractDir)
}

// uploadFile 上传文件到minio [EN] uploadFile uploads files to minio
func uploadFile(ctx context.Context, extractDir string) (docList []*model.DocInfo, err error) {
	defer func() {
		log.Infof("remove file %s ", extractDir)
		err1 := os.RemoveAll(extractDir)
		if err1 != nil {
			log.Errorf("uploadFile remove path %s error %v", extractDir, err)
		}
	}()
	var fileList []*DocFileInfo // 存储解压后所有文件路径的切片 [EN] Store slices of all file paths after decompression
	// 使用filepath.Walk遍历目录 [EN] Traverse directories using filepath.Walk
	err = filepath.Walk(extractDir, walkFunc(&fileList))
	if err != nil {
		log.Infof("upload file to %s", err)
		return nil, err
	}
	fileLen := len(fileList)
	if fileLen == 0 {
		return nil, errors.New("无法解析出可用文档")
	}
	//压缩文件总量截取 [EN] Total amount of compressed files intercepted
	limitNum := int(config.GetConfig().UsageLimit.MaxNumberOfFilesInCompressed)
	if limitNum > 0 && fileLen > limitNum {
		fileList = fileList[0:limitNum]
	}
	//解压后文件大小限制 [EN] File size limit after decompression
	dir := config.GetConfig().Minio.KnowledgeDir
	//循环遍历所有文件,上传minio [EN] Loop through all files and upload minio
	for _, fileInfo := range fileList {
		//上传文档 [EN] Upload documents
		_, minioFilePath, _, err := service.UploadLocalFile(ctx, dir, util.BuildFilePath("", fileInfo.DocInfo.DocType), fileInfo.DocLocalPath)
		if err != nil {
			log.Errorf("upload file err: %v", err)
			continue
		}
		fileInfo.DocInfo.DocUrl = minioFilePath
		docList = append(docList, fileInfo.DocInfo)
	}
	return docList, nil
}

//// checkFile 校验压缩文件，压缩后的文件大小限制过滤 [EN] // checkFile verifies the compressed file and filters the compressed file size limit
//func checkFile(docType string, docSize int64) bool {
//	fileTypeList := strings.Split(config.GetConfig().UsageLimit.FileTypes, ";")
//	for _, suffix := range fileTypeList {
//		if suffix == "" || strings.TrimSpace(suffix) == "" {
//			continue
//		}
//		if docType == suffix {
//			//压缩文件大小校验 [EN] //Compressed file size verification
//			limitSize := config.GetConfig().UsageLimit.FileSizeLimit
//			if limitSize != -1 && docSize > limitSize {
//				return false
//			}
//			return true
//		}
//	}
//	return false
//}

// 使用闭包来捕获files变量，配合filepath.Walk来使用 [EN] Use closures to capture files variables and use them with filepath.Walk
func walkFunc(files *[]*DocFileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf(fmt.Sprintf("walkFunc error accessing path %s file: %v", path, err)) // 处理错误 [EN] handling errors
			return err                                                                      // 返回nil以继续遍历 [EN] Return nil to continue traversing
		}
		log.Infof("walkFunc accessing path %s ", path)
		if !info.IsDir() {
			*files = append(*files, &DocFileInfo{
				DocLocalPath: path,
				DocInfo: &model.DocInfo{
					DocName: info.Name(),
					DocType: filepath.Ext(info.Name()),
					DocSize: info.Size(),
				},
			}) // 添加文件路径到切片 [EN] Add file path to slice
		}
		return nil // 返回nil表示没有错误，可以继续遍历 [EN] Returning nil means there is no error and you can continue traversing
	}
}
