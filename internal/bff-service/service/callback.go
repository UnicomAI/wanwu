package service

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/gin-gonic/gin"
)

// GetDeployInfo 查询部署信息
func GetDeployInfo(ctx *gin.Context) (interface{}, error) {
	return map[string]string{
		"webBaseUrl": config.Cfg().Minio.DownloadURL,
	}, nil
}
