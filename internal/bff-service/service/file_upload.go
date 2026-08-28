package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	FileUploadCheckFileStatusFailed  = 0
	FileUploadCheckFileStatusSuccess = 1
	FileUploadTmpLocalDir            = "tmp"
)

var (
	UnarchiveFilePrefix = "unarchive"
)

func CheckFile(ctx *gin.Context, r *request.CheckFileReq) (*response.CheckFileResp, error) {
	exist, err := util.FileExist(BuildUploadFilePath(r.FileName, r.Sequence, r.ChunkName))
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", err.Error())
	}
	status := FileUploadCheckFileStatusSuccess
	if !exist {
		status = FileUploadCheckFileStatusFailed
	}
	return &response.CheckFileResp{
		Status: status,
	}, nil
}

func CheckFileList(ctx *gin.Context, r *request.CheckFileListReq) (*response.CheckFileListResp, error) {
	dirPath := BuildUploadFilePathDir(r.ChunkName)
	list, err := util.DirFileList(dirPath, false, false)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_not_exist", err.Error())
	}
	sequences, err := BuildUploadFileSeqList(list, false)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", err.Error())
	}
	return &response.CheckFileListResp{
		UploadedFileSequences: sequences,
	}, nil
}

func UploadFile(ctx *gin.Context, r *request.UploadFileReq) (*response.UploadFileResp, error) {
	var err error
	defer func() {
		if err != nil {
			if err := clearChunkFile(r.FileName, r.Sequence, r.ChunkName); err != nil {
				log.Errorf("upload file but clear chunk file err: %v", err)
				return
			}
			return
		}
	}()
	defer util.PrintPanicStack()

	filePath := BuildUploadFilePath(r.FileName, r.Sequence, r.ChunkName)
	exist, err := util.FileExist(filePath)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_not_exist", err.Error())
	}
	if !exist {
		err = saveFileInfo(ctx, filePath)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_save", err.Error())
		}
	}
	return &response.UploadFileResp{
		Status: FileUploadCheckFileStatusSuccess,
	}, nil
}

func MergeFile(ctx *gin.Context, r *request.MergeFileReq) (*response.MergeFileResp, error) {
	var err error
	var mergeFilePath = BuildMergeFilePath(r.FileName, r.ChunkName)
	defer func() {
		if err != nil {
			if err := util.DeleteFile(mergeFilePath); err != nil {
				log.Errorf("merge file but delete file err: %v", err)
				return
			}
			return
		}
		if err := clearChunkDir(r.ChunkName); err != nil {
			log.Errorf("merge file but clear chunk dir err: %v", err)
			return
		}
	}()
	defer util.PrintPanicStack()

	dir := BuildUploadFilePathDir(r.ChunkName)
	list, err := util.DirFileList(dir, false, true)
	if err != nil || len(list) != r.ChunkTotal {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", err.Error())
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})
	sequences, err := BuildUploadFileSeqList(list, true)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", err.Error())
	} else if len(sequences) != r.ChunkTotal {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", fmt.Sprintf("sequences num %v but total chunk %v", len(sequences), r.ChunkTotal))
	}
	for i := 1; i <= r.ChunkTotal; i++ {
		if i != sequences[i-1] {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", "file upload not completed")
		}
	}
	file, err := util.MergeFile(list, mergeFilePath)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_merge", err.Error())
	}
	open, err := os.Open(file.FilePath)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_open", fmt.Sprintf("open file (%v) err: %v", file.FilePath, err))
	}
	defer func() {
		if err := open.Close(); err != nil {
			log.Errorf("merge file but close file (%v) err: %v", file.FilePath, err)
			return
		}
	}()
	defer util.PrintPanicStack()

	if file.TotalByteCount != r.FileSize {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_merge", fmt.Sprintf("merge file total size (%v) but origin file size (%v)", file.TotalByteCount, r.FileSize))
	}
	fileName, _, err := minio.UploadFileCommon(ctx, open, util.FileExt(file.FilePath), file.TotalByteCount, r.IsExpired)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_merge", fmt.Sprintf("merge file but upload minio err: %v", err))
	}
	filePath, err := minio.GetUploadFileCommon(ctx, fileName, r.IsExpired)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_merge", fmt.Sprintf("merge file but get minio file err: %v", err))
	}
	return &response.MergeFileResp{
		OriginalFileName: r.FileName,
		FileName:         fileName,
		FilePath:         filePath,
	}, nil
}

func CleanFile(ctx *gin.Context, r *request.CleanFileReq) (*response.CleanFileResp, error) {
	err := clearChunkDir(r.ChunkName)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_clear", err.Error())
	}
	return &response.CleanFileResp{
		Status: FileUploadCheckFileStatusSuccess,
	}, nil
}

func DeleteFile(ctx *gin.Context, r *request.DeleteFileReq) (*response.DeleteFileResp, error) {
	list := r.FileList
	if len(list) == 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_not_empty")
	}
	for _, file := range list {
		err := minio.DeleteFileCommon(ctx, file, r.IsExpired)
		if err != nil {
			log.Errorf("delete file (%v) err: %v", file, err)
		}
	}
	return &response.DeleteFileResp{
		Status: FileUploadCheckFileStatusSuccess,
	}, nil
}

func FileUrlConvertBase64(ctx *gin.Context, req *request.FileUrlConvertBase64Req) (string, error) {
	resp, err := http.Get(req.FileUrl)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_http_get", err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_http_get", fmt.Sprintf("StatusCode: %d", resp.StatusCode))
	}
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_read", err.Error())
	}

	base64Str, base64StrWithPrefix, err := util.FileData2Base64(fileData, req.CustomPrefix)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_convert_base64", err.Error())
	}
	if req.AddPrefix {
		return base64StrWithPrefix, nil
	} else {
		return base64Str, nil
	}
}

func UploadFileByBase64(ctx *gin.Context, req *request.UploadFileByBase64Req) (*response.UploadFileByBase64Resp, error) {
	var finalFileName string
	if req.FileName == "" {
		finalFileName = util.GenUUID()
	} else {
		finalFileName = req.FileName
	}

	base64Data := req.File
	var inferredExt string

	// 尝试从标准 Data URL 提取 MIME 类型
	if strings.HasPrefix(base64Data, "data:") && strings.Contains(base64Data, ";base64,") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) == 2 {
			header := parts[0]
			base64Data = parts[1]

			mimeType := strings.TrimPrefix(header, "data:")
			if idx := strings.Index(mimeType, ";"); idx != -1 {
				mimeType = mimeType[:idx]
			}

			if mimeType != "" {
				if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 {
					inferredExt = exts[0]
				}
			}
		}
	}

	var finalExt string
	if req.FileExt != "" {
		ext := strings.TrimPrefix(req.FileExt, ".")
		finalExt = "." + ext
	} else if inferredExt != "" {
		finalExt = inferredExt
	}

	// 当 FileName 包含复合扩展名（如 .tar.gz）时，优先使用完整扩展名，
	// 避免 FileExt/MIME 推断只取末尾扩展名导致 .tar.gz 变成 .gz
	if req.FileName != "" {
		if fileExt := util.FileExt(req.FileName); fileExt != "" && len(fileExt) > len(finalExt) {
			finalExt = fileExt
		}
	}

	if finalExt != "" {
		finalFileName += finalExt
	}

	// 解码 base64 数据
	fileData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_base64_decode", err.Error())
	}

	// 使用 minio 上传文件
	reader := bytes.NewReader(fileData)
	_, _, err = minio.UploadFile(ctx, minio.BucketFileUpload, minio.DirFileExpire, finalFileName, reader, int64(len(fileData)))
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_minio", err.Error())
	}

	objectPath := path.Join(minio.BucketFileUpload, minio.DirFileExpire, finalFileName)
	filePath, _ := url.JoinPath(config.Cfg().Minio.DownloadURL, objectPath)

	return &response.UploadFileByBase64Resp{
		Url: filePath,
		Uri: objectPath,
	}, nil
}

// UnarchiveFile 解压压缩包并将文件上传到 MinIO，返回目录树和访问路径
func UnarchiveFile(ctx *gin.Context, r *request.UnarchiveFileReq) (*response.UnarchiveFileResp, error) {
	// 1. 从 MinIO 下载压缩包到本地临时目录
	tmpDir := filepath.Join(FileUploadTmpLocalDir, UnarchiveFilePrefix, util.GenUUID())
	if err := util.MkDir(tmpDir); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_create_tmp_dir", err.Error())
	}
	// 确保临时目录最终被清理
	defer func() {
		if err := util.DeleteDir(tmpDir); err != nil {
			log.Errorf("cleanup tmp dir %s error: %v", tmpDir, err)
		}
	}()

	// 下载压缩包
	archiveData, _, err := minio_util.DownloadFile(ctx.Request.Context(), r.FileUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_download", fmt.Sprintf("download archive file error: %v", err))
	}

	// 从 URL 中获取原始文件名
	_, _, originalFileName := minio_util.SplitMinioPath(r.FileUrl)

	// 检查是否为支持的压缩格式
	if !util.IsSupportedArchive(originalFileName) {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_format", fmt.Sprintf("unsupported archive format: %s", originalFileName))
	}

	// 保存到本地临时文件
	localArchivePath := filepath.Join(tmpDir, originalFileName)
	if err := os.WriteFile(localArchivePath, archiveData, 0644); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_save", fmt.Sprintf("save archive file error: %v", err))
	}

	// 2. 解压到临时子目录
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := util.MkDir(extractDir); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_create_extract_dir", err.Error())
	}

	if err := util.UnarchiveFile(ctx.Request.Context(), localArchivePath, extractDir); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_extract", fmt.Sprintf("extract archive error: %v", err))
	}

	// 3. 将解压后的文件上传到 MinIO
	// localDir 以 "/." 结尾：不包含最后一级目录名，与 TarDir/ZipDir 的 "/." 模式一致
	uploadPrefix := path.Join(minio.DirFileExpire, UnarchiveFilePrefix, util.GenUUID())

	unarchivedFiles, err := minio.UploadDirectory(ctx.Request.Context(), minio.BucketFileUpload, uploadPrefix, extractDir+"/.")
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_unarchive_upload", fmt.Sprintf("upload extracted files to minio error: %v", err))
	}

	// 4. 构建目录树和统计
	children := buildUnarchiveFileTree(unarchivedFiles)

	var totalSize int64
	var fileCount int
	for _, f := range unarchivedFiles {
		if !f.IsDir {
			totalSize += f.Size
			fileCount++
		}
	}

	return &response.UnarchiveFileResp{
		ObjectPath: path.Join(minio.BucketFileUpload, uploadPrefix),
		Children:   children,
		TotalFiles: fileCount,
		TotalSize:  totalSize,
	}, nil
}

// buildUnarchiveFileTree 从扁平文件列表构建目录树结构
// 输入由 UploadDirectory 返回，filepath.Walk 保证目录条目先于其子文件出现，且每个目录/文件均有独立条目
func buildUnarchiveFileTree(files []minio.DirUploadFile) []response.UnarchiveFileNode {
	root := &response.UnarchiveFileNode{
		Children: make([]response.UnarchiveFileNode, 0),
	}

	for _, f := range files {
		parts := strings.Split(f.RelativePath, "/")
		current := root

		for i, part := range parts {
			if part == "" {
				continue
			}

			// 在当前节点的子节点中查找
			found := false
			for idx := range current.Children {
				if current.Children[idx].Name == part {
					current = &current.Children[idx]
					found = true
					break
				}
			}

			if !found {
				node := response.UnarchiveFileNode{
					Name:     part,
					Type:     "directory",
					Children: make([]response.UnarchiveFileNode, 0),
				}
				if i == len(parts)-1 {
					node.RelativePath = f.RelativePath
					node.ObjectPath = f.ObjectPath
					if !f.IsDir {
						node.Type = "file"
						node.Size = f.Size
						node.MinioUrl = f.MinioUrl
						node.DownloadUrl = f.DownloadUrl
					}
				} else {
					node.RelativePath = strings.Join(parts[:i+1], "/")
				}
				current.Children = append(current.Children, node)
				current = &current.Children[len(current.Children)-1]
			}
		}
	}

	sortUnarchiveNodes(root)
	return root.Children
}

// --- internal ---
func saveFileInfo(ctx *gin.Context, filePath string) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return fmt.Errorf("read file (%v) err: %v", filePath, err)
	}
	files := form.File["files"]
	if len(files) == 0 {
		return fmt.Errorf("file (%v) not exist", filePath)
	}
	fileInfo := files[0]
	err = ctx.SaveUploadedFile(fileInfo, filePath)
	if err != nil {
		return fmt.Errorf("save file (%v) err: %v", filePath, err)
	}
	return nil
}

func clearChunkDir(chunkName string) error {
	dir := BuildFilePathDir(chunkName)
	exist, err := util.FileExist(dir)
	if err != nil {
		return fmt.Errorf("check dir (%v) err: %v", dir, err)
	}
	if exist {
		err = util.DeleteDir(dir)
		if err != nil {
			return fmt.Errorf("delete dir (%v) err: %v", dir, err)
		}
	}
	return nil
}

func clearChunkFile(fileName string, sequence int, chunkName string) error {
	filePath := BuildUploadFilePath(fileName, sequence, chunkName)
	exist, err := util.FileExist(filePath)
	if err != nil {
		return fmt.Errorf("check file (%v) err: %v", filePath, err)
	}
	if exist {
		err = util.DeleteFile(filePath)
		if err != nil {
			return fmt.Errorf("delete file (%v) err: %v", filePath, err)
		}
	}
	dirPath := BuildUploadFilePathDir(chunkName)
	exist, err = util.FileExist(dirPath)
	if err != nil {
		return fmt.Errorf("check dir (%v) err: %v", dirPath, err)
	}
	if exist {
		dir, err := os.ReadDir(dirPath)
		if err != nil {
			return fmt.Errorf("read dir (%v) err: %v", dirPath, err)
		}
		if len(dir) == 0 {
			err = util.DeleteDir(dirPath)
			if err != nil {
				return fmt.Errorf("delete dir (%v) err: %v", dirPath, err)
			}
		}
	}
	return nil
}

func BuildFilePathDir(baseFileName string) string {
	fileMd5 := util.MD5([]byte(baseFileName))
	return fmt.Sprintf("%s/%s", FileUploadTmpLocalDir, fileMd5)
}

func BuildUploadFilePath(baseFileName string, sequence int, chunkName string) string {
	fileName := fmt.Sprintf("%010d_%s", sequence, baseFileName)
	return fmt.Sprintf("%s/upload/%s", BuildFilePathDir(chunkName), fileName)
}

func BuildUploadFilePathDir(chunkName string) string {
	return fmt.Sprintf("%s/upload", BuildFilePathDir(chunkName))
}

func BuildMergeFilePath(baseFileName string, chunkName string) string {
	mergeFileName := util.GenUUID() + util.FileExt(baseFileName)
	return fmt.Sprintf("%s/merge/%s", BuildFilePathDir(chunkName), mergeFileName)
}

func BuildUploadFileSeqList(filePathList []string, fullPath bool) ([]int, error) {
	var retList []int
	for _, filePath := range filePathList {
		seq, _, err := BuildChunkSequence(filePath, fullPath)
		if err != nil {
			return nil, err
		}
		retList = append(retList, seq)
	}
	return retList, nil
}

func BuildChunkSequence(storeFile string, fullPath bool) (int, string, error) {
	if fullPath {
		lastIndex := strings.LastIndex(storeFile, "/")
		if lastIndex < 0 {
			return 0, "", fmt.Errorf("store file (%v) is not full path", storeFile)
		}
		storeFile = storeFile[lastIndex+1:]
	}
	splitIndex := strings.Index(storeFile, "_")
	if splitIndex < 0 {
		return 0, "", fmt.Errorf("store file (%v) is empty", storeFile)
	}
	sequence, err := strconv.Atoi(storeFile[0:splitIndex])
	if err != nil {
		return 0, "", fmt.Errorf("store file (%v) is invalid", storeFile)
	}
	return sequence, storeFile[splitIndex+1:], nil
}

func DirectUploadFiles(ctx *gin.Context, r *request.DirectUploadFilesReq) (*response.DirectUploadFilesResp, error) {
	var uploadFiles []*response.DirectUploadFileInfo
	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_save", err.Error())
	}
	files := form.File["files"]
	if len(files) <= 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", fmt.Errorf("file is empty").Error())
	}
	for _, file := range files {
		uploadFileInfo, err := directUploadFile(ctx, r, file)
		if err != nil {
			return nil, err
		}
		uploadFiles = append(uploadFiles, uploadFileInfo)
	}
	return &response.DirectUploadFilesResp{
		Files: uploadFiles,
	}, nil
}

func directUploadFile(ctx *gin.Context, r *request.DirectUploadFilesReq, file *multipart.FileHeader) (*response.DirectUploadFileInfo, error) {
	open, err := file.Open()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_file_open", err.Error())
	}
	defer func() {
		if err = open.Close(); err != nil {
			log.Errorf("close file (%v) err: %v", file, err)
			return
		}
	}()
	defer util.PrintPanicStack()
	fileName, _, err := minio.UploadFileCommon(ctx, open, util.FileExt(file.Filename), file.Size, r.IsExpired)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_minio", fmt.Sprintf("upload minio err: %v", err))
	}
	filePath, err := minio.GetUploadFileCommon(ctx, fileName, r.IsExpired)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_get_minio_path", fmt.Sprintf("get minio file err: %v", err))
	}
	return &response.DirectUploadFileInfo{
		FileName: file.Filename,
		FilePath: filePath,
		FileSize: file.Size,
		FileId:   fileName,
	}, nil
}

// sortUnarchiveNodes 递归排序目录树节点
func sortUnarchiveNodes(node *response.UnarchiveFileNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].Type != node.Children[j].Type {
			return node.Children[i].Type == "directory"
		}
		return node.Children[i].Name < node.Children[j].Name
	})
	for i := range node.Children {
		sortUnarchiveNodes(&node.Children[i])
	}
}
