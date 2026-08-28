package minio

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// DirUploadFile 目录上传到 MinIO 的文件信息
type DirUploadFile struct {
	RelativePath string // 相对于解压根目录的路径
	ObjectName   string // MinIO 中的 objectName
	ObjectPath   string // MinIO 中的完整路径（包含 bucketName）
	MinioUrl     string // 内部 MinIO 下载地址
	DownloadUrl  string // 外部下载地址（如果 DownloadURL 配置了外部访问地址）
	IsDir        bool   // 是否为目录
	Size         int64  // 文件大小（目录为0）
}

// UploadDirectory 将本地目录递归上传到 MinIO
// bucketName: MinIO 存储桶名称
// prefix: MinIO 对象名前缀（如 "file-expire/unarchive/xxx"，不需要以 / 结尾，函数内部会自动补充）
// localDir: 本地目录路径，支持两种模式（与 TarDir 保持一致）：
//   - "/path/to/dir"：包含最后一级目录名，MinIO 对象路径为 "prefix/dir/file1.txt"
//   - "/path/to/dir/."：不包含最后一级目录名，MinIO 对象路径为 "prefix/file1.txt"
//
// 返回所有文件（含目录）的信息列表
// 注意：隐藏文件和目录（以 "." 开头）、macOS 元数据目录（__MACOSX）和资源分叉文件（._ 前缀）将被自动跳过
func UploadDirectory(ctx context.Context, bucketName string, prefix string, localDir string) ([]DirUploadFile, error) {
	var files []DirUploadFile

	// 确保前缀以 / 结尾
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// 检测是否以 /. 结尾，表示不包含基础目录名（与 TarDir 逻辑一致）
	skipBase := strings.HasSuffix(localDir, string(os.PathSeparator)+".")
	localDir = filepath.Clean(localDir)

	var baseName string
	if skipBase {
		// 不包含基础目录名，文件直接在 prefix 下
		baseName = ""
	} else {
		// 包含基础目录名，文件在 prefix/dir/ 下
		baseName = filepath.Base(localDir)
	}

	err := filepath.Walk(localDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localDir, filePath)
		if err != nil {
			return err
		}

		// 跳过根目录自身
		if relPath == "." {
			return nil
		}

		// 跳过隐藏文件和系统元数据
		// 对于隐藏目录，使用 filepath.SkipDir 跳过整个目录树
		entryName := info.Name()
		if util.IsHiddenEntry(entryName) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 跳过符号链接，避免上传预期目录之外的文件
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// 统一使用正斜杠作为路径分隔符（MinIO 对象名使用正斜杠）
		relPath = filepath.ToSlash(relPath)

		// 根据模式拼接对象名
		var objectName string
		if skipBase {
			objectName = path.Join(prefix, relPath)
		} else {
			objectName = path.Join(prefix, baseName, relPath)
		}

		if info.IsDir() {
			files = append(files, DirUploadFile{
				RelativePath: relPath,
				ObjectName:   objectName,
				ObjectPath:   path.Join(bucketName, objectName),
				MinioUrl:     "",
				DownloadUrl:  "",
				IsDir:        true,
				Size:         0,
			})
			return nil
		}

		// 上传文件
		file, err := os.Open(filePath)
		if err != nil {
			log.Errorf("open file %s error: %v", filePath, err)
			return err
		}

		_, size, uploadErr := UploadFile(ctx, bucketName, path.Dir(objectName), path.Base(objectName), file, info.Size())
		closeErr := file.Close()

		if uploadErr != nil {
			log.Errorf("upload file %s to minio error: %v", filePath, uploadErr)
			return uploadErr
		}
		if closeErr != nil {
			log.Errorf("close file %s error: %v", filePath, closeErr)
		}

		files = append(files, DirUploadFile{
			RelativePath: relPath,
			ObjectName:   objectName,
			ObjectPath:   path.Join(bucketName, objectName),
			MinioUrl:     buildMinioUrl(bucketName, objectName),
			DownloadUrl:  buildDownloadUrl(bucketName, objectName),
			IsDir:        false,
			Size:         size,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
