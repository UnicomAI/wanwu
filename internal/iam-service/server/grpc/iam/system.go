package iam

import (
	"context"

	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
)

func (s *Service) GetSystemIdentity(ctx context.Context, req *iam_service.GetSystemIdentityReq) (*iam_service.SystemIdentity, error) {
	return &iam_service.SystemIdentity{
		TopOrgId:    util.Int2Str(config.TopOrgID()),
		AdminUserId: util.Int2Str(config.AdminUserID()),
		AdminRoleId: util.Int2Str(config.AdminRoleID()),
	}, nil
}
