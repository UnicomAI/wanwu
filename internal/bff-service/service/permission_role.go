package service

import (
	"fmt"
	"strings"

	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetRoleTemplate(ctx *gin.Context, userID, orgID string) (*response.RoleTemplate, error) {
	resp, err := GetUserPermission(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	routes := mid.CollectRoutes()
	ret := &response.RoleTemplate{}
	for _, r := range routes {
		if ok, sub := cutRoute(r, resp.OrgPermission.Permissions); ok {
			ret.Routes = append(ret.Routes, sub)
		}
	}
	return ret, nil
}

// GetRoleTemplateExcludeAdminCenter 与 GetRoleTemplate 相同，但过滤掉管理员中心（admin_center）
// 相关的权限分类，用于角色模板接口返回给前端的权限选择树。
// 注意：GetRoleInfo/GetRoleList 仍使用 GetRoleTemplate（不过滤），以保证已持有管理员中心
// 权限的普通角色在详情/列表中能正常解析出权限中文名。
func GetRoleTemplateExcludeAdminCenter(ctx *gin.Context, userID, orgID string) (*response.RoleTemplate, error) {
	ret, err := GetRoleTemplate(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	ret.Routes = filterAdminCenterRoutes(ret.Routes)
	return ret, nil
}

// filterAdminCenterRoutes 过滤掉 Perm 以 "admin_center" 为前缀的路由（含其整个子树）。
// admin_center 顶层分类下挂有 oauth、setting 等子分类，此处一并剔除。
func filterAdminCenterRoutes(routes []response.Route) []response.Route {
	var ret []response.Route
	for _, r := range routes {
		if r.Perm == "admin_center" || strings.HasPrefix(r.Perm, "admin_center.") {
			continue
		}
		r.Children = filterAdminCenterRoutes(r.Children)
		ret = append(ret, r)
	}
	return ret
}
func CreateRole(ctx *gin.Context, creatorID, creatorOrgID, orgID string, roleCreate *request.RoleCreate) (*response.RoleID, error) {
	// creator permission
	creatorPermission, err := GetUserPermission(ctx, creatorID, creatorOrgID)
	if err != nil {
		return nil, err
	}
	// req
	req := &iam_service.CreateRoleReq{
		CreatorId:  creatorID,
		OrgId:      orgID,
		IsGlobal:   roleCreate.IsGlobal,
		Name:       roleCreate.Name,
		Remark:     roleCreate.Remark,
		AvatarPath: roleCreate.Avatar.Key,
	}
	routes := mid.CollectPerms()
	for _, perm := range roleCreate.Permissions {
		var exist bool
		for _, p := range creatorPermission.OrgPermission.Permissions {
			if p.Perm != perm {
				continue
			}
			exist = true
			break
		}
		if !exist {
			return nil, fmt.Errorf("当前用户没有 %v 权限", perm)
		}
		for _, r := range routes {
			if perm == r.Tag || strings.HasPrefix(perm, r.Tag+".") {
				var add bool
				for _, p := range req.Perms {
					if p.Perm == r.Tag {
						add = true
						break
					}
				}
				if !add {
					req.Perms = append(req.Perms, &iam_service.Perm{Perm: r.Tag})
				}
			}
		}
	}
	// create role
	resp, err := iam.CreateRole(ctx.Request.Context(), req)
	if err != nil {
		return nil, err
	}
	return &response.RoleID{RoleID: resp.Id}, nil
}

func ChangeRole(ctx *gin.Context, userID string, roleUpdate *request.RoleUpdate) error {
	// 内置管理员角色不允许编辑
	roleInfo, err := iam.GetRoleInfo(ctx.Request.Context(), &iam_service.GetRoleInfoReq{
		OrgId:  roleUpdate.OrgID,
		RoleId: roleUpdate.RoleID,
	})
	if err != nil {
		return err
	}
	if roleInfo.IsAdmin {
		return fmt.Errorf("组织内置管理员角色不允许编辑")
	}
	// creator permission
	userPermission, err := GetUserPermission(ctx, userID, roleUpdate.OrgID)
	if err != nil {
		return err
	}
	// req
	req := &iam_service.UpdateRoleReq{
		OrgId:      roleUpdate.OrgID,
		RoleId:     roleUpdate.RoleID,
		Name:       roleUpdate.Name,
		Remark:     roleUpdate.Remark,
		AvatarPath: roleUpdate.Avatar.Key,
	}
	routes := mid.CollectPerms()
	for _, perm := range roleUpdate.Permissions {
		var exist bool
		for _, p := range userPermission.OrgPermission.Permissions {
			if p.Perm != perm {
				continue
			}
			exist = true
			break
		}
		if !exist {
			return fmt.Errorf("当前用户没有 %v 权限", perm)
		}
		for _, r := range routes {
			if perm == r.Tag || strings.HasPrefix(perm, r.Tag+".") {
				var add bool
				for _, p := range req.Perms {
					if p.Perm == r.Tag {
						add = true
						break
					}
				}
				if !add {
					req.Perms = append(req.Perms, &iam_service.Perm{Perm: r.Tag})
				}
			}
		}
	}
	_, err = iam.UpdateRole(ctx.Request.Context(), req)
	return err
}

func DeleteRole(ctx *gin.Context, roleDelete *request.RoleDelete) error {
	_, err := iam.DeleteRole(ctx.Request.Context(), &iam_service.DeleteRoleReq{
		OrgId:  roleDelete.OrgID,
		RoleId: roleDelete.RoleID,
	})
	return err
}

func GetRoleInfo(ctx *gin.Context, userID, orgID, roleID string) (*response.RoleInfo, error) {
	role, err := iam.GetRoleInfo(ctx.Request.Context(), &iam_service.GetRoleInfoReq{
		OrgId:  orgID,
		RoleId: roleID,
	})
	if err != nil {
		return nil, err
	}
	template, err := GetRoleTemplate(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	return toRoleInfo(role, template), nil
}

func GetRoleList(ctx *gin.Context, userID, orgID, name string) (*response.ListResult, error) {
	// 传大 pageSize 拉取全量数据，BFF 层不再分页
	const fetchAllPageSize int32 = 9999
	resp, err := iam.GetRoleList(ctx.Request.Context(), &iam_service.GetRoleListReq{
		OrgId:    orgID,
		Name:     name,
		PageNo:   1,
		PageSize: fetchAllPageSize,
	})
	if err != nil {
		return nil, err
	}
	template, err := GetRoleTemplate(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	var roles []*response.RoleInfo
	for _, role := range resp.Roles {
		roles = append(roles, toRoleInfo(role, template))
	}
	// also include global roles
	globalResp, err := iam.GetGlobalRoleList(ctx.Request.Context(), &iam_service.GetGlobalRoleListReq{
		Name:     name,
		OrgId:    orgID,
		PageNo:   1,
		PageSize: fetchAllPageSize,
	})
	if err != nil {
		return nil, err
	}
	for _, role := range globalResp.Roles {
		roles = append(roles, toRoleInfo(role, template))
	}
	return &response.ListResult{
		List:  roles,
		Total: resp.Total + globalResp.Total,
	}, nil
}

func ChangeRoleStatus(ctx *gin.Context, orgID, roleID string, status bool) error {
	_, err := iam.ChangeRoleStatus(ctx.Request.Context(), &iam_service.ChangeRoleStatusReq{
		OrgId:  orgID,
		RoleId: roleID,
		Status: status,
	})
	return err
}

func GetRoleUsers(ctx *gin.Context, roleID, name, orgID string, pageNo, pageSize int32) (*response.PageResult, error) {
	resp, err := iam.GetRoleUsers(ctx.Request.Context(), &iam_service.GetRoleUsersReq{
		RoleId:   roleID,
		Name:     name,
		OrgId:    orgID,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	var users []response.RoleUser
	for _, u := range resp.Users {
		ru := response.RoleUser{
			UserID:   u.UserId,
			UserName: u.UserName,
			Phone:    u.Phone,
			Email:    u.Email,
			Avatar:   cacheUserAvatar(u.AvatarPath),
		}
		for _, g := range u.Orgs {
			ru.Orgs = append(ru.Orgs, response.IDNameWithAvatar{
				ID:     g.Id,
				Name:   g.Name,
				Avatar: cacheOrgAvatar(g.AvatarPath),
			})
		}
		users = append(users, ru)
	}
	return &response.PageResult{
		List:     users,
		Total:    resp.Total,
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

func RemoveRoleUser(ctx *gin.Context, roleID, userID, orgID string) error {
	_, err := iam.RemoveRoleUser(ctx.Request.Context(), &iam_service.RemoveRoleUserReq{
		RoleId: roleID,
		UserId: userID,
		OrgId:  orgID,
	})
	return err
}

// IsGlobalRole 检查指定角色是否是全局角色
func IsGlobalRole(ctx *gin.Context, orgID, roleID string) (bool, error) {
	resp, err := iam.GetRoleInfo(ctx.Request.Context(), &iam_service.GetRoleInfoReq{
		RoleId: roleID,
	})
	if err != nil {
		// 查不到角色信息，说明角色不存在，视为非全局角色
		return false, nil
	}
	return resp.IsGlobal, nil
}

// --- internal ---

func toRoleIDName(ctx *gin.Context, role *iam_service.RoleIDName) response.RoleIDName {
	ret := response.RoleIDName{
		ID:       role.Id,
		Name:     role.Name,
		Avatar:   cacheRoleAvatar(role.AvatarPath),
		IsGlobal: role.IsGlobal,
	}
	if role.IsAdmin {
		if role.IsSystem {
			ret.Name = gin_util.I18nKey(ctx, "bff_role_system_admin_name")
		} else {
			ret.Name = gin_util.I18nKey(ctx, "bff_role_org_admin_name")
		}
	}
	return ret
}

func toRoleIDNames(ctx *gin.Context, roles []*iam_service.RoleIDName) []response.RoleIDName {
	var ret []response.RoleIDName
	for _, role := range roles {
		ret = append(ret, toRoleIDName(ctx, role))
	}
	return ret
}

func toRoleInfo(role *iam_service.RoleInfo, template *response.RoleTemplate) *response.RoleInfo {
	ret := &response.RoleInfo{
		RoleID:       role.RoleId,
		Name:         role.Name,
		Remark:       role.Remark,
		CreatedAt:    util.Time2Str(role.CreatedAt),
		Creator:      toUserIDNameWithAvatar(role.Creator),
		Status:       role.Status,
		IsAdmin:      role.IsAdmin,
		IsGlobal:     role.IsGlobal,
		OrgName:      role.OrgName,
		UserCount:    role.UserCount,
		Avatar:       cacheRoleAvatar(role.AvatarPath),
		RoleTemplate: template,
	}
	if role.IsAdmin {
		ret.Permissions = toPermissions(true, false, false, nil)
		return ret
	}
	if role.IsGlobal {
		// 全局角色：直接返回角色存储的权限，Name 用全量权限表解析（不受查看者权限裁剪）
		permNameMap := collectPermNameMap()
		for _, perm := range role.Perms {
			ret.Permissions = append(ret.Permissions, response.Permission{
				Perm: perm.Perm,
				Name: permNameMap[perm.Perm],
			})
		}
		return ret
	}
	for _, perm := range role.Perms {
		for _, route := range template.Routes {
			if ok, name := inRoute(route, perm.Perm); ok {
				ret.Permissions = append(ret.Permissions, response.Permission{
					Perm: perm.Perm,
					Name: name,
				})
				break
			}
		}
	}
	return ret
}

// collectPermNameMap 用全量权限表构建 perm tag -> 中文名 的映射。
// mid.CollectPerms 返回所有 PermNeedCheck 路由（含 setting/operation/ontology 等系统级权限），
// 未经查看者权限裁剪，用于全局角色权限名解析。
func collectPermNameMap() map[string]string {
	ret := make(map[string]string)
	for _, p := range mid.CollectPerms() {
		ret[p.Tag] = p.Name
	}
	return ret
}

func inRoute(r response.Route, perm string) (bool, string) {
	if r.Perm == perm {
		return true, r.Name
	}
	for _, child := range r.Children {
		if ok, name := inRoute(child, perm); ok {
			return true, name
		}
	}
	return false, ""
}

func cutRoute(r route.Route, perms []response.Permission) (bool, response.Route) {
	var exist bool
	var ret response.Route
	for _, perm := range perms {
		if perm.Perm == r.Tag {
			exist = true
			ret.Perm = perm.Perm
			ret.Name = perm.Name
			break
		}
	}
	if !exist {
		return false, ret
	}
	for _, sub := range r.Subs {
		if ok, child := cutRoute(sub, perms); ok {
			ret.Children = append(ret.Children, child)
		}
	}
	return true, ret
}
