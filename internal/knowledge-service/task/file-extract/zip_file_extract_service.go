package file_extract

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/UnicomAI/wanwu/pkg/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type ZipFileExtractServiceService struct {
}

var zipFileExtractServiceService = &ZipFileExtractServiceService{}

func init() {
	AddFileExtractService(zipFileExtractServiceService)
}

func (t ZipFileExtractServiceService) ExtractFileType() string {
	return ".zip"
}

func (t ZipFileExtractServiceService) ExtractFile(ctx context.Context, localFilePath string, destDir string) (extractDir string, err error) {
	fileReader, err := zip.OpenReader(localFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err1 := fileReader.Close(); err1 != nil {
			log.Errorf("ZipFileExtractServiceService file close error %v", err)
		}
	}()

	for _, f := range fileReader.Reader.File {
		var decodeFileName string
		if f.Flags == 0 { //本地编码，默认GBK，转换成UTF-8 [EN] Local encoding, default GBK, converted to UTF-8
			i := bytes.NewReader([]byte(f.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := io.ReadAll(decoder)
			decodeFileName = string(content)
		} else {
			decodeFileName = f.Name
		}
		// 构建完整的文件路径 [EN] Build full file path
		destFilePath := filepath.Join(destDir, decodeFileName)
		// 检查是否为目录 [EN] Check if it is a directory
		if f.FileInfo().IsDir() {
			// 创建目录 [EN] Create directory
			if err := os.MkdirAll(destFilePath, f.Mode()); err != nil {
				fmt.Printf("无法创建目录: %v\n", err)
			}
			continue
		}
		log.Infof("ExtractFile file path %s", destFilePath)
		// 我们需要确保所有的文件夹都已经创建好 [EN] We need to make sure all folders have been created
		err = os.MkdirAll(filepath.Dir(destFilePath), f.Mode())
		if err != nil {
			return "", err
		}
		//写入文件 [EN] write file
		err = writeUnzipFile(f, destFilePath)
		if err != nil {
			return "", err
		}
	}
	return destDir, nil
}

// writeUnzipFile 写入文件 [EN] writeUnzipFile writes a file
func writeUnzipFile(zipFile *zip.File, destFilePath string) error {
	//打开目标文件 [EN] Open target file
	destFile, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			log.Errorf("ZipFileExtractServiceService file close error %v", err)
		}
	}()

	//打开源压缩文件 [EN] Open source compressed file
	sourceFile, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Errorf("ZipFileExtractServiceService file close error %v", err)
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
