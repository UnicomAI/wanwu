package minio

import (
	"bytes"
	"context"
	"net/url"
	"strings"
)

var (
	_minioSafety *client
)

func InitSafety(ctx context.Context, cfg Config, initBucketName string) error {
	if _minioSafety == nil {
		c, err := newClient(cfg)
		if err != nil {
			return err
		}
		_minioSafety = c
	}
	// 创建存储桶并设置存储策略 [EN] Create a bucket and set storage policy
	if err := _minioSafety.CreateBucketIfNotExist(ctx, initBucketName); err != nil {
		return err
	}

	return nil
}

func Safety() *client {
	return _minioSafety
}

func SplitFilePath(filePath string) (bucketName string, objectName string, fileName string) {
	if len(filePath) == 0 {
		return "", "", ""
	}
	u, err := url.Parse(filePath)
	if err != nil {
		return "", "", ""
	}
	//此处拿到的path是以"/"开头的，因此split的时候split[0]="",数据从split[1]开始 [EN] The path obtained here starts with "/", so when splitting, split[0]="", the data starts from split[1]
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

func DownloadFileToMemory(ctx context.Context, minioFilePath string) ([]byte, error) {
	_, objectName, _ := SplitFilePath(minioFilePath)
	return Safety().GetObject(ctx, BucketFileUpload, objectName)
}
