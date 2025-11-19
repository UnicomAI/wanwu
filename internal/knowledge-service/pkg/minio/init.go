package minio

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/pkg/minio"
)

var minioClient = ClientMinio{}

type ClientMinio struct {
}

func init() {
	pkg.AddContainer(minioClient)
}

func (c ClientMinio) LoadType() string {
	return "minioClient"
}

func (c ClientMinio) Load() error {
	minioConfig := config.GetConfig().Minio
	//初始化知识库内部使用minio-bucket [EN] Initialize the knowledge base internally using minio-bucket
	err := minio.InitKnowledge(context.Background(), minio.Config{
		Endpoint: minioConfig.EndPoint,
		User:     minioConfig.User,
		Password: minioConfig.Password,
	}, minioConfig.Bucket, false)
	if err != nil {
		return err
	}
	//初始化对外公开bucket [EN] Initialize the public bucket
	err = minio.InitKnowledge(context.Background(), minio.Config{
		Endpoint: minioConfig.EndPoint,
		User:     minioConfig.User,
		Password: minioConfig.Password,
	}, minioConfig.PublicRagBucket, true)
	if err != nil {
		return err
	}
	return nil
}

func (c ClientMinio) StopPriority() int {
	return pkg.DefaultPriority
}

func (c ClientMinio) Stop() error {
	return nil
}
