package service

import (
	"errors"
	"github.com/gin-gonic/gin"
)

type BizService interface {
	BizType() string
	SearchBizOwner(ctx *gin.Context, bizId string) (userId, orgId string, err error)
}

var builderMap = make(map[string]BizService)

func InitBizService(bizOwner BizService) {
	builderMap[bizOwner.BizType()] = bizOwner
}

func OwnerInfo(ctx *gin.Context, bizType, bizId string) (userId, orgId string, err error) {
	owner := builderMap[bizType]
	if owner == nil {
		return "", "", errors.New("no biz owner found")
	}
	return owner.SearchBizOwner(ctx, bizId)
}
