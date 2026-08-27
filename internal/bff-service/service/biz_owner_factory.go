package service

import (
	"errors"
	"github.com/UnicomAI/wanwu/api/proto/common"

	"github.com/gin-gonic/gin"
)

type BizConversationLog struct {
}

type BizService interface {
	BizType() string
	SearchBizOwner(ctx *gin.Context, bizId string) (userId, orgId string, err error)
	SearchConversationLog(ctx *gin.Context, bizId, sourceFrom string) (*common.ConversationLog, error)
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

func ConversationLog(ctx *gin.Context, bizType, bizId, sourceFrom string) (*common.ConversationLog, error) {
	owner := builderMap[bizType]
	if owner == nil {
		return nil, errors.New("no biz conversation found")
	}
	return owner.SearchConversationLog(ctx, bizId, sourceFrom)
}
