package minio_service

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	minio_client "github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/minio/minio-go/v7"
)

func DeleteFile(ctx context.Context, minioFilePath string) error {
	minioFilePath = strings.ReplaceAll(minioFilePath, "minio/download/api/", "")
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	log.Infof("DeleteFile bucketName %s objectName %s", bucketName, objectName)
	err := minio_client.FileUpload().DeleteObject(ctx, bucketName, objectName)
	if err != nil {
		log.Errorf("DeleteFile error %s", err)
		return err
	}
	return nil
}

func CopyFile(ctx context.Context, srcFilePath string, destObjectNamePre string, open bool) (string, string, int64, error) {
	srcFilePath = strings.ReplaceAll(srcFilePath, "minio/download/api/", "")
	bucketName, objectName, fileName := SplitFilePath(srcFilePath)
	if len(bucketName) == 0 || len(objectName) == 0 {
		return "", "", 0, errors.New("invalid file path")
	}
	destObjectName := buildObjectName(destObjectNamePre, fileName)
	minioConfig := config.Cfg().Minio

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
	uploadInfo, err := minio_client.FileUpload().Cli().CopyObject(ctx, destOptions, srcOptions)
	if err != nil {
		log.Errorf("minio copy object error %s", err)
		return "", "", 0, err
	}
	var minioUrl = "http://" + minioConfig.EndPoint
	if open {
		minioUrl = minioConfig.DownloadURL
	}
	return minioUrl + "/" + minioConfig.Bucket + "/" + destObjectName, fileName, uploadInfo.Size, nil
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

func SplitFilePath(filePath string) (bucketName string, objectName string, fileName string) {
	if len(filePath) == 0 {
		return "", "", ""
	}
	u, err := url.Parse(filePath)
	if err != nil {
		return "", "", ""
	}
	split := strings.Split(u.Path, "/")
	var pathSplit []string
	for _, s := range split {
		if s == "" {
			continue
		}
		pathSplit = append(pathSplit, s)
	}
	totalLen := len(pathSplit)
	if totalLen > 1 {
		var buffer bytes.Buffer
		for i := 1; i < totalLen; i++ {
			buffer.WriteString(pathSplit[i])
			if i < totalLen-1 {
				buffer.WriteString("/")
			}
		}
		return pathSplit[0], buffer.String(), pathSplit[totalLen-1]
	}
	return "", "", filePath
}

func buildObjectName(dir, fileName string) string {
	if len(dir) == 0 {
		return fileName
	}
	return dir + "/" + fileName
}
