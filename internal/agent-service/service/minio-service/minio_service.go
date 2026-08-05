package minio_service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_http_client "github.com/UnicomAI/wanwu/internal/agent-service/pkg/http"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	minio_client "github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/minio/minio-go/v7"
)

// URLType 表示URL的类型
type URLType string

const (
	URLTypeMinioDirect    URLType = "MINIO_DIRECT"    // 直接MinIO URL
	URLTypeMinioPresigned URLType = "MINIO_PRESIGNED" // MinIO预签名URL
	URLTypeOther          URLType = "OTHER"           // 其他URL
)

func DownloadFile(ctx context.Context, minioFilePath string) ([]byte, error) {
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	object, err := minio_client.Agent().GetObject(ctx, bucketName, objectName)
	if err != nil {
		log.Errorf("DownloadFile error %s", err)
		return nil, err
	}
	return object, nil
}

func DownloadFileObject(ctx context.Context, minioFilePath string) (*minio.Object, error) {
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	object, err := minio_client.Agent().Cli().GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("DownloadFile error %s", err)
		return nil, err
	}
	return object, nil
}

func DeleteFile(ctx context.Context, minioFilePath string) error {
	bucketName, objectName, _ := SplitFilePath(minioFilePath)
	err := minio_client.Agent().DeleteObject(ctx, bucketName, objectName)
	if err != nil {
		log.Errorf("DeleteFile error %s", err)
		return err
	}
	return nil
}

// UploadLocalFile 根据文件路径上传文件
func UploadLocalFile(ctx context.Context, minioDir string, minioFileName string, srcFilePath string, open bool) (string, string, int64, error) {
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
	filePath, fileSize, err := UploadFile(ctx, minioDir, minioFileName, srcFile, fileUploadSize, open)
	return minioFileName, filePath, fileSize, err
}

func CopyFile(ctx context.Context, srcFilePath string, destObjectNamePre string, open bool) (string, string, int64, error) {
	srcFilePath = strings.ReplaceAll(srcFilePath, "minio/download/api/", "")
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
	uploadInfo, err := minio_client.Agent().Cli().CopyObject(ctx, destOptions, srcOptions)
	if err != nil {
		log.Errorf("minio copy object error %s", err)
		return "", "", 0, err
	}
	configInfo := config.GetConfig()
	var minioUrl = "http://" + configInfo.Minio.EndPoint
	if open {
		minioUrl = configInfo.Minio.DownloadUrl
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

func UploadFile(ctx context.Context, dir string, fileName string, reader io.Reader, objectSize int64, open bool) (string, int64, error) {
	bucketName := config.GetConfig().Minio.Bucket
	// 上传文件。
	//milli := time.Now().UnixMilli()
	var uploadInfo minio.UploadInfo
	objectName := buildObjectName(dir, fileName)
	contentType := getContentType(objectName)
	putObjectOptions := minio.PutObjectOptions{}
	if len(contentType) > 0 {
		putObjectOptions.ContentType = contentType
	}
	uploadInfo, err := minio_client.Agent().Cli().PutObject(ctx, bucketName, objectName, reader, objectSize, putObjectOptions)

	//log_config.LogRpcJsonNoParams("minio", "PutObject", err, milli)
	if err != nil {
		//log-config.Fatalln(err)
		return "", 0, err
	}
	configInfo := config.GetConfig()
	var minioUrl = "http://" + configInfo.Minio.EndPoint
	if open {
		minioUrl = configInfo.Minio.DownloadUrl
	}
	return minioUrl + "/" + bucketName + "/" + objectName, uploadInfo.Size, nil
	//return uploadInfo.Location, uploadInfo.Size, nil
}

func DownloadFileToLocal(ctx context.Context, minioFilePath string, localPath string) error {
	urlType := getURLType(minioFilePath)
	if urlType == URLTypeMinioPresigned {
		return DownloadFileToLocalDirect(ctx, minioFilePath, localPath)
	}
	//替换掉对外的url
	minioFilePath = strings.ReplaceAll(minioFilePath, "/minio/download/api", "")
	sourceBucketName, objectName, _ := SplitFilePath(minioFilePath)
	object, err := minio_client.Agent().Cli().GetObject(ctx, sourceBucketName, objectName, minio.GetObjectOptions{})
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

func DownloadFileToLocalDirect(ctx context.Context, minioFilePath string, localPath string) error {
	resp, err := agent_http_client.GetClient().GetOriResp(ctx, &http_client.HttpRequestParams{
		Url:        minioFilePath,
		Timeout:    2 * time.Minute,
		MonitorKey: "http_download_service",
		LogLevel:   http_client.LogParams,
	})
	if err != nil {
		log.Errorf("DownloadFileToLocalDirect error %s", err)
		return err
	}

	defer func() {
		err2 := resp.Body.Close()
		if err2 != nil {
			log.Errorf("DownloadFile %s close resp.Body error: %s", minioFilePath, err2)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
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
	_, err = io.Copy(outFile, resp.Body)
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
	//此处拿到的path是以"/"开头的，因此split的时候split[0]="",数据从split[1]开始
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

// IsMinioPresignedURL 判断是否是MinIO预签名URL
func IsMinioPresignedURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// 预签名URL通常包含这些查询参数
	query := parsedURL.Query()
	presignedParams := []string{
		"X-Amz-Algorithm",
		"X-Amz-Credential",
		"X-Amz-Date",
		"X-Amz-Expires",
		"X-Amz-SignedHeaders",
		"X-Amz-Signature",
		"AWSAccessKeyId", // 旧版本参数
		"Signature",      // 旧版本参数
	}

	for _, param := range presignedParams {
		if query.Get(param) != "" {
			return true
		}
	}

	return false
}

// GetURLType 获取URL类型
func getURLType(rawURL string) URLType {
	if IsMinioPresignedURL(rawURL) {
		return URLTypeMinioPresigned
	}
	return URLTypeMinioDirect
}
