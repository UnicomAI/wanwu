package service

import (
	"context"
	"errors"
)

type BizService interface {
	BizType() string
	SearchBizOwner(ctx context.Context, bizId string) (userId, orgId string, err error)
}

var builderMap = make(map[string]BizService)

func InitBizService(bizOwner BizService) {
	builderMap[bizOwner.BizType()] = bizOwner
}

func OwnerInfo(ctx context.Context, bizType, bizId string) (userId, orgId string, err error) {
	owner := builderMap[bizType]
	if owner == nil {
		return "", "", errors.New("no biz owner found")
	}
	return owner.SearchBizOwner(ctx, bizId)
}
