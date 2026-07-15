package iam

import (
	"context"
	"strconv"

	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/iam-service/client"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

type Service struct {
	iam_service.UnimplementedIAMServiceServer
	cli client.IClient
}

func NewService(cli client.IClient) *Service {
	return &Service{
		cli: cli,
	}
}

func (s *Service) InitData() error {
	ctx := context.Background()
	var topOrgID, adminRoleID, adminUserID uint32
	var err error
	var status *errs.Status
	// 系统内置唯一顶级组织（挂载全部一级组织）
	if topOrgID, err = s.cli.GetTopOrg(ctx); err == gorm.ErrRecordNotFound {
		if topOrgID, status = s.cli.CreateOrg(ctx, &model.Org{
			Status: true,
			Name:   "--- 系统 ---",
		}); status != nil {
			return errStatus(errs.Code_IAMGeneral, status)
		}
	} else if err != nil {
		return err
	}
	config.InitTopOrgID(topOrgID)
	log.Infof("system top org: %v", config.TopOrgID())
	// 系统顶级组织内置管理员角色
	if adminRoleID, err = s.cli.GetAdminRole(ctx); err != nil {
		return err
	}
	config.InitAdminRoleID(adminRoleID)
	log.Infof("system admin role: %v", config.AdminRoleID())
	// 系统内置管理员
	if adminUserID, err = s.cli.GetAdminUser(ctx); err == gorm.ErrRecordNotFound {
		if adminUserID, status = s.cli.CreateUser(ctx, &model.User{
			IsAdmin:  true,
			Status:   true,
			Name:     "admin",
			Nick:     "admin",
			Password: "Wanwu123456",
		}, topOrgID, []uint32{adminRoleID}); status != nil {
			return errStatus(errs.Code_IAMGeneral, status)
		}
	} else if err != nil {
		return err
	}
	config.InitAdminUserID(adminUserID)
	log.Infof("system admin user: %v", config.AdminUserID())
	return nil
}

// --- internal method ---
func errStatus(code errs.Code, status *errs.Status) error {
	return grpc_util.ErrorStatusWithKey(code, status.TextKey, status.Args...)
}

func toErrStatus(key string, args ...string) *errs.Status {
	return &errs.Status{
		TextKey: key,
		Args:    args,
	}
}

// --- internal function ---

func toIDName(idName orm.IDName) *iam_service.IDName {
	return &iam_service.IDName{Id: strconv.Itoa(int(idName.ID)), Name: idName.Name, NameStatus: idName.NameStatus}
}

func toIDNames(idNames []orm.IDName) []*iam_service.IDName {
	var ret []*iam_service.IDName
	for _, idName := range idNames {
		ret = append(ret, toIDName(idName))
	}
	return ret
}

func toProtoIDNameWithAvatars(idNames []orm.IDNameWithAvatar) []*iam_service.IDNameWithAvatar {
	var ret []*iam_service.IDNameWithAvatar
	for _, idName := range idNames {
		ret = append(ret, &iam_service.IDNameWithAvatar{
			Id:         strconv.Itoa(int(idName.ID)),
			Name:       idName.Name,
			NameStatus: idName.NameStatus,
			AvatarPath: idName.AvatarPath,
		})
	}
	return ret
}

func toIDFullName(idFullName orm.IDFullName) *iam_service.IDFullName {
	return &iam_service.IDFullName{Id: strconv.Itoa(int(idFullName.ID)), Name: idFullName.Name, FullName: idFullName.FullName}
}

func toIDFullNames(idFullNames []orm.IDFullName) []*iam_service.IDFullName {
	var ret []*iam_service.IDFullName
	for _, idFullName := range idFullNames {
		ret = append(ret, toIDFullName(idFullName))
	}
	return ret
}

func toRoleIDName(roleIDName orm.RoleIDName) *iam_service.RoleIDName {
	return &iam_service.RoleIDName{Id: strconv.Itoa(int(roleIDName.ID)), Name: roleIDName.Name, IsAdmin: roleIDName.IsAdmin, IsSystem: roleIDName.IsSystem, IsGlobal: roleIDName.IsGlobal}
}

func toRoleIDNames(roleIDNames []orm.RoleIDName) []*iam_service.RoleIDName {
	var ret []*iam_service.RoleIDName
	for _, roleIDName := range roleIDNames {
		ret = append(ret, toRoleIDName(roleIDName))
	}
	return ret
}

func toOffset(req iReq) int32 {
	if req.GetPageNo() < 1 || req.GetPageSize() < 0 {
		return -1
	}
	return (req.GetPageNo() - 1) * req.GetPageSize()
}

type iReq interface {
	GetPageNo() int32 // 从1开始
	GetPageSize() int32
}
