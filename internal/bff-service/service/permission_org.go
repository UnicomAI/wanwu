package service

import (
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
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
	return toOrgInfo(ctx, org), nil
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
		orgs = append(orgs, toOrgInfo(ctx, org))
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

func GetAdminOrgSelect(ctx *gin.Context, userID string) (*response.Select, error) {
	resp, err := iam.GetAdminOrgSubTree(ctx.Request.Context(), &iam_service.GetAdminOrgSubTreeReq{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	// 从树中提取所有有管理权限的组织（hasPerm=true），展平为列表
	var selects []response.IDNameWithAvatar
	collectAdminOrgs(resp.Orgs, &selects)
	return &response.Select{Select: selects}, nil
}

// --- internal ---

func toOrgIDNamesWithAvatar(ctx *gin.Context, orgs []*iam_service.IDNameWithAvatar, isSystemAdmin bool) []response.IDNameWithAvatar {
	var ret []response.IDNameWithAvatar
	for _, org := range orgs {
		if len(orgs) > 1 && org.Id == config.TopOrgID && !isSystemAdmin {
			continue
		}
		ret = append(ret, toOrgIDNameWithAvatar(ctx, org))
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
			Avatar:   cacheOrgAvatar(node.AvatarPath),
			Children: toAdminOrgTreeNodes(node.Children),
		})
	}
	return ret
}

// collectAdminOrgs 递归遍历树，收集所有 hasPerm=true 的组织到扁平列表
func collectAdminOrgs(nodes []*iam_service.AdminOrgTreeNode, out *[]response.IDNameWithAvatar) {
	for _, node := range nodes {
		if node.HasPerm {
			*out = append(*out, response.IDNameWithAvatar{
				ID:     node.OrgId,
				Name:   node.Name,
				Avatar: cacheOrgAvatar(node.AvatarPath),
			})
		}
		collectAdminOrgs(node.Children, out)
	}
}

func toOrgInfo(ctx *gin.Context, org *iam_service.OrgInfo) *response.OrgInfo {
	return &response.OrgInfo{
		OrgID:     org.OrgId,
		Name:      org.Name,
		Remark:    org.Remark,
		Creator:   toUserIDNameWithAvatar(org.Creator),
		CreatedAt: util.Time2Str(org.CreatedAt),
		Status:    org.Status,
		UserCount: org.UserCount,
		Admins:    org.Admins,
		Avatar:    cacheOrgAvatar(org.AvatarPath),
	}
}
