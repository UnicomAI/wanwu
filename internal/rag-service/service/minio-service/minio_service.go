package minio_service

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	minio_client "github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/minio/minio-go/v7"
)

// Enabled minio 初始化失败时降级为不转存不清理，调用方据此跳过
func Enabled() bool {
	return minio_client.FileUpload() != nil && config.Cfg().Minio != nil
}

func DeleteFile(ctx context.Context, minioFilePath string) error {
	if !Enabled() {
		return nil
	}
	minioFilePath = strings.ReplaceAll(minioFilePath, "minio/download/api/", "")
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	if len(bucketName) == 0 || len(objectName) == 0 {
		return nil
	}
	log.Infof("DeleteFile bucketName %s objectName %s", bucketName, objectName)
	if err := minio_client.FileUpload().DeleteObject(ctx, bucketName, objectName); err != nil {
		log.Errorf("DeleteFile error %s", err)
		return err
	}
	return nil
}

// CopyFile 把源文件拷到 destObjectNamePre 目录下，返回新地址；open 为 true 返回对外下载地址
func CopyFile(ctx context.Context, srcFilePath string, destObjectNamePre string, open bool) (string, error) {
	if !Enabled() {
		return "", errors.New("minio not init")
	}
	srcFilePath = strings.ReplaceAll(srcFilePath, "minio/download/api/", "")
	bucketName, objectName, fileName := SplitFilePath(srcFilePath)
	if len(bucketName) == 0 || len(objectName) == 0 {
		return "", errors.New("invalid file path")
	}
	minioConfig := config.Cfg().Minio
	destObjectName := buildObjectName(destObjectNamePre, fileName)
	_, err := minio_client.FileUpload().Cli().CopyObject(ctx,
		minio.CopyDestOptions{Bucket: minioConfig.Bucket, Object: destObjectName},
		minio.CopySrcOptions{Bucket: bucketName, Object: objectName})
	if err != nil {
		log.Errorf("minio copy object error %s", err)
		return "", err
	}
	minioUrl := "http://" + minioConfig.EndPoint
	if open {
		minioUrl = minioConfig.DownloadURL
	}
	return minioUrl + "/" + minioConfig.Bucket + "/" + destObjectName, nil
}

func SplitFilePath(filePath string) (bucketName string, objectName string, fileName string) {
	if len(filePath) == 0 {
		return "", "", ""
	}
	u, err := url.Parse(filePath)
	if err != nil {
		return "", "", ""
	}
	var pathSplit []string
	for _, s := range strings.Split(u.Path, "/") {
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
