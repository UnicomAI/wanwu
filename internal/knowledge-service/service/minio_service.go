package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	minio_client "github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/minio/minio-go/v7"
)

func DownloadFile(ctx context.Context, minioFilePath string) ([]byte, error) {
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	object, err := minio_client.Knowledge().GetObject(ctx, bucketName, objectName)
	if err != nil {
		log.Errorf("DownloadFile error %s", err)
		return nil, err
	}
	return object, nil
}

func DownloadFileObject(ctx context.Context, minioFilePath string) (*minio.Object, error) {
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	object, err := minio_client.Knowledge().Cli().GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("DownloadFile error %s", err)
		return nil, err
	}
	return object, nil
}

func DeleteFile(ctx context.Context, minioFilePath string) error {
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	err := minio_client.Knowledge().DeleteObject(ctx, bucketName, objectName)
	if err != nil {
		log.Errorf("DeleteFile error %s", err)
		return err
	}
	return nil
}

// UploadLocalFile uploads files based on file path
func UploadLocalFile(ctx context.Context, minioDir string, minioFileName string, srcFilePath string) (string, string, int64, error) {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		log.Errorf("UploadLocalFile open file error :%s", err)
		return "", "", 0, err
	}
	defer func() {
		err2 := srcFile.Close()
		if err2 != nil {
			log.Errorf("UploadFilePath close file error :%s", err2)
		}
	}()
	fileInfo, err := os.Stat(srcFilePath)
	var fileUploadSize int64 = -1
	if err == nil {
		fileUploadSize = fileInfo.Size()
	}
	filePath, fileSize, err := UploadFile(ctx, minioDir, minioFileName, srcFile, fileUploadSize)
	return minioFileName, filePath, fileSize, err
}

func CopyFile(ctx context.Context, srcFilePath string, destObjectNamePre string) (string, string, int64, error) {
	bucketName, objectName, fileName := SplitFilePath(srcFilePath)
	if len(bucketName) == 0 || len(objectName) == 0 {
		return "", "", 0, errors.New("invalid file path")
	}
	destObjectName := buildObjectName(destObjectNamePre, fileName)
	minioConfig := config.GetConfig().Minio

	destOptions := minio.CopyDestOptions{
		Bucket: minioConfig.Bucket,
		Object: destObjectName,
	}
	contentType := getContentType(destObjectName)
	if len(contentType) > 0 {
		destOptions.ReplaceMetadata = true
		destOptions.UserMetadata = map[string]string{
			"Content-Type": contentType,
		}
	}
	srcOptions := minio.CopySrcOptions{
		Bucket: bucketName,
		Object: objectName,
	}
	uploadInfo, err := minio_client.Knowledge().Cli().CopyObject(ctx, destOptions, srcOptions)
	if err != nil {
		log.Errorf("minio copy object error %s", err)
		return "", "", 0, err
	}
	if len(uploadInfo.Location) == 0 {
		return "http://" + minioConfig.EndPoint + "/" + minioConfig.Bucket + "/" + destObjectName, fileName, uploadInfo.Size, nil
	}
	return uploadInfo.Location, fileName, uploadInfo.Size, nil
}

func getContentType(uri string) (contentType string) {
	//_ = mime.AddExtensionType(".svg", "image/svg+xml")
	//_ = mime.AddExtensionType(".svgz", "image/svg+xml")
	//_ = mime.AddExtensionType(".webp", "image/webp")
	//_ = mime.AddExtensionType(".ico", "image/x-icon")
	//fileExtension := path.Base(uri)
	//ext := path.Ext(fileExtension)
	//contentType = mime.TypeByExtension(ext)
	return ""
}

func UploadFile(ctx context.Context, dir string, fileName string, reader io.Reader, objectSize int64) (string, int64, error) {
	bucketName := config.GetConfig().Minio.Bucket
	// Upload files.
	//milli := time.Now().UnixMilli()
	var uploadInfo minio.UploadInfo
	objectName := buildObjectName(dir, fileName)
	contentType := getContentType(objectName)
	putObjectOptions := minio.PutObjectOptions{}
	if len(contentType) > 0 {
		putObjectOptions.ContentType = contentType
	}
	uploadInfo, err := minio_client.Knowledge().Cli().PutObject(ctx, bucketName, objectName, reader, objectSize, putObjectOptions)

	//log_config.LogRpcJsonNoParams("minio", "PutObject", err, milli)
	if err != nil {
		//log-config.Fatalln(err)
		return "", 0, err
	}
	if len(uploadInfo.Location) == 0 {
		configInfo := config.GetConfig()
		return "http://" + configInfo.Minio.EndPoint + "/" + bucketName + "/" + objectName, uploadInfo.Size, nil
	}
	return uploadInfo.Location, uploadInfo.Size, nil
}

func DownloadFileToLocal(ctx context.Context, minioFilePath string, localPath string) error {
	sourceBucketName, objectName, _ := SplitFilePath(minioFilePath)
	object, err := minio_client.Knowledge().Cli().GetObject(ctx, sourceBucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("DownloadFileToLocal error %s", err)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer func() {
		err2 := outFile.Close()
		if err2 != nil {
			log.Errorf("DownloadFile %s close outFile error: %s", minioFilePath, err2)
		}
	}()
	defer func() {
		err2 := object.Close()
		if err2 != nil {
			log.Errorf("DownloadFile %s close error: %s", minioFilePath, err2)
		}
	}()
	_, err = io.Copy(outFile, object)

	if util.FileEOF(err) {
		return nil
	}
	return err
}

func SplitFilePath(filePath string) (bucketName string, objectName string, fileName string) {
	if len(filePath) == 0 {
		return "", "", ""
	}
	u, err := url.Parse(filePath)
	if err != nil {
		return "", "", ""
	}
	//The path obtained here starts with "/", so when splitting, split[0]="", the data starts from split[1]
	split := strings.Split(u.Path, "/")
	totalLen := len(split)
	if totalLen > 2 {
		var buffer bytes.Buffer
		for i := 2; i < totalLen; i++ {
			buffer.WriteString(split[i])
			if i < totalLen-1 {
				buffer.WriteString("/")
			}
		}
		return split[1], buffer.String(), split[totalLen-1]
	}
	return "", "", filePath
}

func buildObjectName(dir, fileName string) string {
	if len(dir) == 0 {
		return fileName
	}
	return dir + "/" + fileName
}
