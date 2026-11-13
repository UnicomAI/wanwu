package file_extract

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type TarGzFileExtractServiceService struct {
}

var tarGzFileExtractServiceService = &TarGzFileExtractServiceService{}

func init() {
	AddFileExtractService(tarGzFileExtractServiceService)
}

func (t TarGzFileExtractServiceService) ExtractFileType() string {
	return ".tar.gz"
}

func (t TarGzFileExtractServiceService) ExtractFile(ctx context.Context, localFilePath string, destDir string) (extractDir string, err error) {
	// 打开.tar.gz文件
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Errorf("error closing file: %v", err)
		}
	}()
	// 创建gzip读取器
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Errorf("error gzip new reader file: %v", err)
		return "", err
	}
	defer func() {
		err = gzipReader.Close()
		if err != nil {
			log.Errorf("error gzip close reader: %v", err)
		}
	}()

	// 创建tar读取器
	tarReader := tar.NewReader(gzipReader)
	// 确保目标目录存在
	if err = os.MkdirAll(destDir, 0755); err != nil {
		log.Errorf("error make dir: %v", err)
		return "", err
	}

	// 遍历tar包中的文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 已遍历完全部的文件
		}
		if err != nil {
			log.Errorf("error tar reader: %v", err)
			return "", err
		}
		// 安全检查：把不可信条目名解析为 destDir 下的绝对路径，
		// 字符串级 CleanArchiveEntryName + 文件系统级 filepath.Rel 双重校验，
		// 与 unarchive.go 的 ArchiveEntryPath 标准对齐。
		path, err := util.ArchiveEntryPath(destDir, header.Name)
		if err != nil {
			log.Errorf("skip unsafe tar entry: %s, error: %v", header.Name, err)
			continue
		}
		// 根据文件类型创建文件夹或者写文件
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			log.Errorf("error make dir: %v", err)
			return "", err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg:
			err = writeFile(path, header, tarReader)
			if err != nil {
				return "", err
			}
		default:
			// 忽略其他类型的文件
			continue
		}
	}
	return destDir, nil
}

// writeFile 写入文件
func writeFile(filePath string, header *tar.Header, tarReader *tar.Reader) error {
	// 打开文件进行写入
	writer, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer func() {
		err = writer.Close()
		log.Errorf("error make dir: %v", err)
	}()

	// 从tar包中读取文件内容并写入到文件中
	if _, err := io.Copy(writer, tarReader); err != nil {
		return err
	}
	return nil
}
