package knowledge_permission

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_permission_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-permission-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *Service) SelectKnowledgeUserPermission(ctx context.Context, req *knowledgebase_permission_service.KnowledgeUserPermissionReq) (*knowledgebase_permission_service.KnowledgeUserPermissionResp, error) {
	permissionList, err := orm.SelectUserKnowledgePermissionList(ctx, req.KnowledgeId)
	if err != nil {
		log.Errorf(fmt.Sprintf("SelectKnowledgeUserPermission 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionSelectedFailed)
	}
	return buildKnowledgePermissionResp(permissionList), nil
}

func (s *Service) CheckKnowledgeUserPermission(ctx context.Context, req *knowledgebase_permission_service.CheckKnowledgeUserPermissionReq) (*emptypb.Empty, error) {
	permission, err := orm.SelectUserKnowledgePermission(ctx, req.UserId, req.OrgId, req.KnowledgeId)
	if err != nil {
		log.Errorf(fmt.Sprintf("CheckKnowledgeUserPermission 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	if permission.PermissionType < int(req.PermissionType) {
		log.Errorf(fmt.Sprintf("CheckKnowledgeUserPermission 权限不足失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) AddKnowledgeUser(ctx context.Context, req *knowledgebase_permission_service.AddKnowledgeUserReq) (*emptypb.Empty, error) {
	err := orm.BatchCreateKnowledgeIdPermission(db.GetHandle(ctx), buildKnowledgePermissionList(req))
	if err != nil {
		log.Errorf(fmt.Sprintf("AddKnowledgeUser 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeAddPermissionFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) EditKnowledgeUser(ctx context.Context, req *knowledgebase_permission_service.EditKnowledgeUserReq) (*emptypb.Empty, error) {
	editList := buildKnowledgePermissionEditList(req)
	if err := checkKnowledgePermission(ctx, req.UserId, req.OrgId, req.KnowledgeId, editList[0]); err != nil {
		log.Errorf(fmt.Sprintf("EditKnowledgeUser checkKnowledgeDelete 权限不足失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	err := db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		if err := orm.BatchEditKnowledgePermission(tx, editList); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Errorf(fmt.Sprintf("EditKnowledgeUser 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeEditPermissionFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteKnowledgeUser(ctx context.Context, req *knowledgebase_permission_service.DeleteKnowledgeUserReq) (*emptypb.Empty, error) {
	permission, err := orm.SelectKnowledgePermissionById(ctx, req.PermissionId)
	if err != nil {
		log.Errorf(fmt.Sprintf("DeleteKnowledgeUser 权限不足失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	if err = checkKnowledgePermission(ctx, req.UserId, req.OrgId, req.KnowledgeId, permission); err != nil {
		log.Errorf(fmt.Sprintf("DeleteKnowledgeUser checkKnowledgeDelete 权限不足失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	err = orm.DeleteKnowledgePermissionById(ctx, permission)
	if err != nil {
		log.Errorf(fmt.Sprintf("DeleteKnowledgeUser 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeDeletePermissionFailed)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) TransferKnowledgeAdminUser(ctx context.Context, req *knowledgebase_permission_service.TransferKnowledgeAdminUserReq) (*emptypb.Empty, error) {
	editList, addList := buildKnowledgePermissionTransferList(req)
	//In theory, this operation should add a distributed lock in the knowledge base dimension.
	err := db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		if len(editList) > 0 {
			if err := orm.BatchEditKnowledgePermission(tx, editList); err != nil {
				return err
			}
		}
		if len(addList) > 0 {
			if err := orm.BatchCreateKnowledgeIdPermission(tx, addList); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Errorf(fmt.Sprintf("TransferKnowledgeAdminUser 失败(%v)  参数(%v)", err, req))
		return nil, util.ErrCode(errs.Code_KnowledgeTransferPermissionFailed)
	}
	return &emptypb.Empty{}, nil
}

// checkKnowledgePermission Check user operation permissions
func checkKnowledgePermission(ctx context.Context, userId, orgId, knowledgeId string, operation *model.KnowledgePermission) error {
	permission, err := orm.SelectUserKnowledgePermission(ctx, userId, orgId, knowledgeId)
	if err != nil {
		log.Errorf(fmt.Sprintf("DeleteKnowledgeUser 删除失败(%v)  参数(userId %s orgId %s knowledgeId %s)", err, userId, orgId, knowledgeId))
		return util.ErrCode(errs.Code_KnowledgePermissionDeny)
	}
	//System administrators can operate
	if permission.PermissionType == model.PermissionTypeSystem {
		return nil
	}
	//Authorized administrators can only operate ordinary users
	if permission.PermissionType == model.PermissionTypeGrant && operation.PermissionType < model.PermissionTypeGrant {
		return nil
	}

	//The rest have no permissions
	return util.ErrCode(errs.Code_KnowledgePermissionDeny)
}

// buildKnowledgePermissionResp builds the knowledge base permission list and queries the returned data
func buildKnowledgePermissionResp(knowledgePermissionList []*model.KnowledgePermission) *knowledgebase_permission_service.KnowledgeUserPermissionResp {
	var knowledgeUserInfoList []*knowledgebase_permission_service.KnowledgeUserInfo
	for _, knowledgePermission := range knowledgePermissionList {
		knowledgeUserInfoList = append(knowledgeUserInfoList, &knowledgebase_permission_service.KnowledgeUserInfo{
			UserId:         knowledgePermission.UserId,
			OrgId:          knowledgePermission.OrgId,
			PermissionType: int32(knowledgePermission.PermissionType),
			PermissionId:   knowledgePermission.PermissionId,
		})
	}
	return &knowledgebase_permission_service.KnowledgeUserPermissionResp{
		KnowledgeUserList: knowledgeUserInfoList,
	}
}

// buildKnowledgePermissionList build knowledge base permission list
func buildKnowledgePermissionList(req *knowledgebase_permission_service.AddKnowledgeUserReq) []*model.KnowledgePermission {
	var dataList []*model.KnowledgePermission
	milli := time.Now().UnixMilli()
	for _, info := range req.KnowledgeUserList {
		dataList = append(dataList, &model.KnowledgePermission{
			PermissionId:   generator.GetGenerator().NewID(),
			KnowledgeId:    req.KnowledgeId,
			GrantOrgId:     req.OrgId,
			GrantUserId:    req.UserId,
			PermissionType: int(info.PermissionType),
			UserId:         info.UserId,
			OrgId:          info.OrgId,
			CreatedAt:      milli,
			UpdatedAt:      milli,
		})
	}
	return dataList
}

// buildKnowledgePermissionEditList Build knowledge base permission list
func buildKnowledgePermissionEditList(req *knowledgebase_permission_service.EditKnowledgeUserReq) (editList []*model.KnowledgePermission) {
	editList = append(editList, &model.KnowledgePermission{
		PermissionId:   req.KnowledgeUser.PermissionId,
		PermissionType: int(req.KnowledgeUser.PermissionType),
		GrantUserId:    req.UserId,
		GrantOrgId:     req.OrgId,
	})
	return editList
}

// buildKnowledgePermissionTransferList Build knowledge base permission transfer list
func buildKnowledgePermissionTransferList(req *knowledgebase_permission_service.TransferKnowledgeAdminUserReq) (editList, addList []*model.KnowledgePermission) {
	editList = append(editList, &model.KnowledgePermission{
		PermissionId:   req.PermissionId,
		PermissionType: model.PermissionTypeEdit,
		GrantUserId:    req.UserId,
		GrantOrgId:     req.OrgId,
	})
	//Normally, this is a new addition
	milli := time.Now().UnixMilli()
	addList = append(addList, &model.KnowledgePermission{
		PermissionId:   generator.GetGenerator().NewID(),
		KnowledgeId:    req.KnowledgeId,
		GrantOrgId:     req.OrgId,
		GrantUserId:    req.UserId,
		PermissionType: model.PermissionTypeSystem,
		UserId:         req.KnowledgeUser.UserId,
		OrgId:          req.KnowledgeUser.OrgId,
		CreatedAt:      milli,
		UpdatedAt:      milli,
	})
	return editList, addList
}
