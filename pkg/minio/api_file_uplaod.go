package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/minio/minio-go/v7"
)

var (
	BucketFileUpload = "file-upload"
	DirFileExpire    = "file-expire"
	DirFileNotExpire = "file-not-expire"
	StoreExpireDays  = 1
	// BucketAAA = "aaa-upload"
	// BucketBBB = "bbb-upload"
)

var (
	_minioFileUpload *client
	// _minioAAA *client
	// _minioBBB *client
)

func InitFileUpload(ctx context.Context, cfg Config) error {
	if _minioFileUpload != nil {
		return fmt.Errorf("minio fileupload client already init")
	}
	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	_minioFileUpload = c
	if _, err = _minioFileUpload.createBucketIfAbsent(ctx, BucketFileUpload); err != nil {
		return err
	}
	if err = _minioFileUpload.SetPathExpireByDay(ctx, BucketFileUpload, DirFileExpire, StoreExpireDays); err != nil {
		return err
	}
	return nil
}

func FileUpload() *client {
	return _minioFileUpload
}

func UploadFileCommon(ctx context.Context, reader io.Reader, fileType string, objectSize int64, isExpired bool) (string, int64, error) {
	if isExpired {
		return UploadFileCommonWithNotExpire(ctx, reader, fileType, objectSize)
	}
	return UploadFileCommonWithExpire(ctx, reader, fileType, objectSize)
}

func UploadFileCommonWithExpire(ctx context.Context, reader io.Reader, fileType string, objectSize int64) (string, int64, error) {
	fileName := util.GenUUID() + fileType
	_, i, err := UploadFile(ctx, BucketFileUpload, DirFileExpire, fileName, reader, objectSize)
	return fileName, i, err
}

func UploadFileCommonWithNotExpire(ctx context.Context, reader io.Reader, fileType string, objectSize int64) (string, int64, error) {
	fileName := util.GenUUID() + fileType
	_, i, err := UploadFile(ctx, BucketFileUpload, DirFileNotExpire, fileName, reader, objectSize)
	return fileName, i, err
}

func DeleteFileCommon(ctx context.Context, fileName string, isExpired bool) error {
	if isExpired {
		return DeleteFileCommonWithNotExpire(ctx, fileName)
	}
	return DeleteFileCommonWithExpire(ctx, fileName)
}

func DeleteFileCommonWithExpire(ctx context.Context, fileName string) error {
	objectName := buildObjectName(DirFileExpire, fileName)
	return _minioFileUpload.cli.RemoveObject(ctx, BucketFileUpload, objectName, minio.RemoveObjectOptions{})
}

func DeleteFileCommonWithNotExpire(ctx context.Context, fileName string) error {
	objectName := buildObjectName(DirFileNotExpire, fileName)
	return _minioFileUpload.cli.RemoveObject(ctx, BucketFileUpload, objectName, minio.RemoveObjectOptions{})
}

func GetUploadFileCommon(ctx context.Context, fileName string, isExpired bool) (string, error) {
	if isExpired {
		return GetUploadFileWithNotExpire(ctx, fileName)
	}
	return GetUploadFileWithExpire(ctx, fileName)
}

func GetUploadFileWithExpire(ctx context.Context, fileName string) (string, error) {
	objectName := buildObjectName(DirFileExpire, fileName)
	object, err := _minioFileUpload.cli.StatObject(ctx, BucketFileUpload, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", err
	}
	return buildFilePath(BucketFileUpload, object.Key), nil
}

func GetUploadFileWithNotExpire(ctx context.Context, fileName string) (string, error) {
	objectName := buildObjectName(DirFileNotExpire, fileName)
	object, err := _minioFileUpload.cli.StatObject(ctx, BucketFileUpload, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", err
	}
	return buildFilePath(BucketFileUpload, object.Key), nil
}

func UploadFile(ctx context.Context, bucketName string, dir string, fileName string, reader io.Reader, objectSize int64) (string, int64, error) {
	var uploadInfo minio.UploadInfo
	objectName := buildObjectName(dir, fileName)
	uploadInfo, err := _minioFileUpload.cli.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{})

	if err != nil {
		//log-config.Fatalln(err)
		return "", 0, err
	}
	if len(uploadInfo.Location) == 0 {
		return buildFilePath(bucketName, objectName), uploadInfo.Size, nil
	}
	return uploadInfo.Location, uploadInfo.Size, nil
}

func buildObjectName(dir, fileName string) string {
	return dir + "/" + fileName
}

func buildFilePath(bucketName, objectName string) string {
	return "http://" + _minioFileUpload.config.Endpoint + "/" + bucketName + "/" + objectName
}

func (c *client) createBucketIfAbsent(ctx context.Context, bucketName string) (bool, error) {
	exists, err := c.cli.BucketExists(ctx, bucketName)
	if err != nil {
		log.Errorf("BucketExists %s error %s", BucketFileUpload, err)
		return false, err
	}

	if !exists {
		// 创建存储桶 [EN] Create bucket
		err = c.cli.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Errorf("create bucket %s error %s", BucketFileUpload, err)
			return false, err
		}
		// 设置存储桶策略以公开访问。 [EN] Set bucket policy for public access.
		policy := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Deny",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:GetBucketLocation",
                "s3:ListBucket",
                "s3:ListBucketMultipartUploads"
            ],
            "Resource": [
                "arn:aws:s3:::` + bucketName + `"
            ]
        },
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:ListMultipartUploadParts",
                "s3:PutObject",
                "s3:AbortMultipartUpload",
                "s3:DeleteObject",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::` + bucketName + `/*"
            ]
        }
    ]
}`
		// 设置存储桶策略。 [EN] Set bucket policy.
		err = c.cli.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return false, err
		}
	}
	return !exists, nil
}
