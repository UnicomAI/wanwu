package service

import (
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreateOrg(ctx *gin.Context, creatorID, parentID string, orgCreate *request.OrgCreate) (*response.OrgID, error) {
	resp, err := iam.CreateOrg(ctx.Request.Context(), &iam_service.CreateOrgReq{
		CreatorId:  creatorID,
		ParentId:   parentID,
		Name:       orgCreate.Name,
		Remark:     orgCreate.Remark,
		AvatarPath: orgCreate.Avatar.Key,
	})
	if err != nil {
		return nil, err
	}
	return &response.OrgID{OrgID: resp.Id}, nil
}

func ChangeOrg(ctx *gin.Context, orgUpdate *request.OrgUpdate) error {
	_, err := iam.UpdateOrg(ctx.Request.Context(), &iam_service.UpdateOrgReq{
		OrgId:      orgUpdate.OrgID.OrgID,
		Name:       orgUpdate.Name,
		Remark:     orgUpdate.Remark,
		AvatarPath: orgUpdate.Avatar.Key,
	})
	return err
}

func DeleteOrg(ctx *gin.Context, parentID, orgID string) error {
	_, err := iam.DeleteOrg(ctx.Request.Context(), &iam_service.DeleteOrgReq{
		OrgId: orgID,
	})
	return err
}

func GetOrgInfo(ctx *gin.Context, orgID string) (*response.OrgInfo, error) {
	org, err := iam.GetOrgInfo(ctx.Request.Context(), &iam_service.GetOrgInfoReq{
		OrgId: orgID,
	})
	if err != nil {
		return nil, err
	}
	return toOrgInfo(org), nil
}

func GetOrgList(ctx *gin.Context, parentID, name string, pageNo, pageSize int32) (*response.PageResult, error) {
	resp, err := iam.GetOrgList(ctx.Request.Context(), &iam_service.GetOrgListReq{
		ParentId: parentID,
		Name:     name,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	var orgs []*response.OrgInfo
	for _, org := range resp.Orgs {
		orgs = append(orgs, toOrgInfo(org))
	}
	return &response.PageResult{
		List:     orgs,
		Total:    resp.Total,
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

func ChangeOrgStatus(ctx *gin.Context, parentID, orgID string, status bool) error {
	_, err := iam.ChangeOrgStatus(ctx.Request.Context(), &iam_service.ChangeOrgStatusReq{
		OrgId:  orgID,
		Status: status,
	})
	return err
}

func GetAdminOrgSubTree(ctx *gin.Context, userID string) ([]*response.AdminOrgTreeNode, error) {
	resp, err := iam.GetAdminOrgSubTree(ctx.Request.Context(), &iam_service.GetAdminOrgSubTreeReq{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return toAdminOrgTreeNodes(resp.Orgs), nil
}

// --- internal ---

func toOrgIDName(ctx *gin.Context, org *iam_service.IDName) response.IDName {
	if org.Id == config.TopOrgID {
		org.Name = gin_util.I18nKey(ctx, "bff_top_org_name")
	}
	return response.IDName{
		ID:   org.Id,
		Name: org.Name,
	}
}

func toOrgIDNames(ctx *gin.Context, orgs []*iam_service.IDName, isSystemAdmin bool) []response.IDName {
	var ret []response.IDName
	for _, org := range orgs {
		if len(orgs) > 1 && org.Id == config.TopOrgID && !isSystemAdmin {
			continue
		}
		ret = append(ret, toOrgIDName(ctx, org))
	}
	return ret
}

func toAdminOrgTreeNodes(nodes []*iam_service.AdminOrgTreeNode) []*response.AdminOrgTreeNode {
	if len(nodes) == 0 {
		return nil
	}
	var ret []*response.AdminOrgTreeNode
	for _, node := range nodes {
		ret = append(ret, &response.AdminOrgTreeNode{
			OrgID:    node.OrgId,
			Name:     node.Name,
			HasPerm:  node.HasPerm,
			IsSystem: node.OrgId == config.TopOrgID,
			Children: toAdminOrgTreeNodes(node.Children),
		})
	}
	return ret
}

func toOrgInfo(org *iam_service.OrgInfo) *response.OrgInfo {
	return &response.OrgInfo{
		OrgID:     org.OrgId,
		Name:      org.Name,
		Remark:    org.Remark,
		Creator:   toIDName(org.Creator),
		CreatedAt: util.Time2Str(org.CreatedAt),
		Status:    org.Status,
		UserCount: org.UserCount,
		Admins:    org.Admins,
		Avatar:    cacheOrgAvatar(org.AvatarPath),
	}
}
