package minio

import (
	"context"
)

var (
	_minioApp *client
)

func InitApp(ctx context.Context, cfg Config, initBucketName string) error {
	if _minioApp == nil {
		c, err := newClient(cfg)
		if err != nil {
			return err
		}
		_minioApp = c
	}
	// 初始化桶
	if err := _minioApp.CreateBucketIfNotExist(ctx, initBucketName); err != nil {
		return err
	}
	return nil
}

func App() *client {
	return _minioApp
}
