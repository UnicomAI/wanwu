package orm

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gromitlee/access"
	"github.com/gromitlee/access/pkg/perm"
	"gorm.io/gorm"
)

func (c *Client) GetTopOrg(ctx context.Context) (uint32, error) {
	org := &model.Org{}
	err := sqlopt.WithParentID(0).Apply(c.db.WithContext(ctx)).First(org).Error
	return org.ID, err
}

func (c *Client) GetOrg(ctx context.Context, orgID uint32) (*OrgInfo, *errs.Status) {
	var ret *OrgInfo
	var err error
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		org := &model.Org{}
		if err = sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
			return toErrStatus("iam_org_get", err.Error())
		}
		ret, err = toOrgInfoTx(tx, org)
		if err != nil {
			return toErrStatus("iam_org_get", err.Error())
		}
		return nil
	})
}

func (c *Client) GetOrgs(ctx context.Context, parentID uint32, name string, offset, limit int32) ([]*OrgInfo, int64, *errs.Status) {
	var ret []*OrgInfo
	var count int64
	return ret, count, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var orgs []*model.Org
		orgsQuery := sqlopt.WithParentID(parentID).Apply(tx).Select("id").Table("orgs")
		if err := sqlopt.LikeName(name).Apply(tx).Where("id IN (?)", orgsQuery).
			Offset(int(offset)).Limit(int(limit)).Order("id DESC").Find(&orgs).
			Offset(-1).Limit(-1).Count(&count).Error; err != nil {
			return toErrStatus("iam_orgs_get", err.Error())
		}
		for _, org := range orgs {
			info, err := toOrgInfoTx(tx, org)
			if err != nil {
				return toErrStatus("iam_orgs_get", err.Error())
			}
			ret = append(ret, info)
		}
		return nil
	})
}

func (c *Client) SelectOrgs(ctx context.Context, userID uint32) ([]IDNameWithAvatar, *errs.Status) {
	var ret []IDNameWithAvatar
	var orgTree *model.OrgNode
	var err error
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// org tree
		orgTree, err = getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_orgs_select", err.Error())
		}
		ret, err = selectOrgs(tx, userID, orgTree)
		if err != nil {
			return toErrStatus("iam_orgs_select", err.Error())
		}
		return nil
	})

}

func (c *Client) GetOrgByOrgIDs(ctx context.Context, orgIDs []uint32) ([]IDFullName, *errs.Status) {
	var ret []IDFullName
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var orgs []*model.Org
		if err := sqlopt.WithIDs(orgIDs).Apply(tx).Find(&orgs).Error; err != nil {
			return toErrStatus("iam_orgs_get_by_ids", err.Error())
		}
		// 组织树，用于获取全名（祖先 - ... - 本级）
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_orgs_get_by_ids", err.Error())
		}
		for _, org := range orgs {
			ret = append(ret, IDFullName{
				IDNameWithAvatar: IDNameWithAvatar{
					ID:         org.ID,
					Name:       org.Name,
					AvatarPath: org.AvatarPath,
				},
				FullName: orgTree.GetFullName(org.ID),
			})
		}
		return nil
	})
}

// GetOrgParentPath 返回从内部顶级组织到指定组织直属父组织的完整父级路径，不包含指定组织。
func (c *Client) GetOrgParentPath(ctx context.Context, orgID uint32) ([]IDNameWithAvatar, *errs.Status) {
	var ret []IDNameWithAvatar
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_org_get", err.Error())
		}
		if orgTree.GetOrg(orgID) == nil {
			return toErrStatus("iam_org_get", gorm.ErrRecordNotFound.Error())
		}
		for _, node := range orgTree.GetParentPath(orgID, true) {
			ret = append(ret, IDNameWithAvatar{
				ID:         node.GetOrgID(),
				Name:       node.GetOrgName(node.GetOrgID()),
				AvatarPath: node.GetAvatarPath(),
			})
		}
		return nil
	})
}

func (c *Client) GetOrgUsersAndSubOrgs(ctx context.Context, orgID uint32) (*OrgUsersResult, *errs.Status) {
	if orgID == 0 {
		ret := &OrgUsersResult{}
		return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
			orgTree, err := getOrgTree(tx)
			if err != nil {
				return toErrStatus("iam_org_users_get", err.Error())
			}
			node := orgTree.GetOrg(config.TopOrgID())
			if node == nil {
				return toErrStatus("iam_org_users_get", gorm.ErrRecordNotFound.Error())
			}
			ret.Orgs = append(ret.Orgs, IDNameWithAvatar{ID: node.GetOrgID(), Name: node.GetOrgName(node.GetOrgID()), AvatarPath: node.GetAvatarPath()})
			return nil
		})
	}
	return c.getOrgUsersAndSubOrgs(ctx, orgID)
}

func (c *Client) SearchOrgAndUser(ctx context.Context, name string) (*OrgUsersResult, *errs.Status) {
	return c.queryOrgUsers(ctx, name)
}

// getOrgUsersAndSubOrgs 根据组织 ID 获取直属下级组织和当前组织用户。
// 未传组织 ID 时仅返回系统顶级组织；系统顶级组织不返回用户。
func (c *Client) getOrgUsersAndSubOrgs(ctx context.Context, orgID uint32) (*OrgUsersResult, *errs.Status) {
	ret := &OrgUsersResult{}
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_org_users_get", err.Error())
		}
		if orgID != 0 {
			node := orgTree.GetOrg(orgID)
			if node == nil {
				return toErrStatus("iam_org_users_get", gorm.ErrRecordNotFound.Error())
			}
			var orgs []*model.Org
			if err := sqlopt.WithParentID(orgID).Apply(tx).Order("id ASC").Find(&orgs).Error; err != nil {
				return toErrStatus("iam_org_users_get", err.Error())
			}
			for _, o := range orgs {
				ret.Orgs = append(ret.Orgs, IDNameWithAvatar{ID: o.ID, Name: o.Name, AvatarPath: o.AvatarPath})
			}
			if orgID != config.TopOrgID() {
				var ous []*model.OrgUser
				if err := sqlopt.WithOrgID(orgID).Apply(tx).Find(&ous).Error; err != nil {
					return toErrStatus("iam_org_users_get", err.Error())
				}
				ids := make([]uint32, 0, len(ous))
				for _, ou := range ous {
					ids = append(ids, ou.UserID)
				}
				ret.Users, _ = c.SelectUsersByUserIDs(ctx, ids)
			}
		}
		return nil
	})
}

// queryOrgUsers 根据名称在全系统模糊搜索组织和用户，并组装用户所属组织路径。
func (c *Client) queryOrgUsers(ctx context.Context, name string) (*OrgUsersResult, *errs.Status) {
	ret := &OrgUsersResult{}
	if name == "" {
		return nil, toErrStatus("iam_org_users_search", "name is empty")
	}
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_org_users_get", err.Error())
		}
		var orgs []*model.Org
		if err := tx.Where("name LIKE ?", "%"+name+"%").Order("id ASC").Find(&orgs).Error; err != nil {
			return toErrStatus("iam_org_users_search", err.Error())
		}
		for _, o := range orgs {
			ret.SearchOrgs = append(ret.SearchOrgs, IDNameWithAvatar{ID: o.ID, Name: o.Name, AvatarPath: o.AvatarPath})
		}
		var ous []*model.OrgUser
		if err := tx.Find(&ous).Error; err != nil {
			return toErrStatus("iam_org_users_search", err.Error())
		}
		var users []*model.User
		if err := tx.Where("name LIKE ?", "%"+name+"%").Find(&users).Error; err != nil {
			return toErrStatus("iam_org_users_search", err.Error())
		}
		um := map[uint32]*model.User{}
		for _, u := range users {
			um[u.ID] = u
		}
		for _, ou := range ous {
			u := um[ou.UserID]
			if u == nil {
				continue
			}
			n := orgTree.GetOrg(ou.OrgID)
			if n == nil {
				continue
			}
			var p []IDNameWithAvatar
			for _, x := range append(orgTree.GetParentPath(ou.OrgID, true), n) {
				p = append(p, IDNameWithAvatar{ID: x.GetOrgID(), Name: x.GetOrgName(x.GetOrgID()), AvatarPath: x.GetAvatarPath()})
			}
			ret.SearchUsers = append(ret.SearchUsers, OrgUserSearchItem{User: IDNameWithAvatar{ID: u.ID, Name: u.Name, AvatarPath: u.AvatarPath}, Orgs: p})
		}
		return nil
	})
}

func (c *Client) GetOrgAndSubOrgSelectByUser(ctx context.Context, userID, orgID uint32) ([]IDNameWithAvatar, *errs.Status) {
	var result []IDNameWithAvatar
	return result, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// 获取组织树
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_orgs_select", err.Error())
		}
		crurentOrgTree := orgTree.GetOrg(orgID)
		orgs, err := selectOrgs(tx, userID, crurentOrgTree)
		if err != nil {
			return toErrStatus("iam_orgs_select", err.Error())
		}
		for _, org := range orgs {
			result = append(result, IDNameWithAvatar{ID: org.ID, Name: org.Name, AvatarPath: org.AvatarPath})
		}
		return nil
	})
}

// GetAdminOrgIDs 查询用户有管理员权限的组织及其所有子孙组织的ID集合。
// 系统管理员返回所有组织。
func (c *Client) GetAdminOrgIDs(ctx context.Context, userID uint32) ([]uint32, *errs.Status) {
	nodes, err := c.GetAdminOrgSubTree(ctx, userID)
	if err != nil {
		return nil, err
	}
	var orgIDs []uint32
	collectOrgIDs(nodes, &orgIDs)
	return orgIDs, nil
}

// collectOrgIDs 从管理员组织树中收集所有 hasPerm=true 的组织ID
func collectOrgIDs(nodes []*AdminOrgTreeNode, ids *[]uint32) {
	for _, n := range nodes {
		if n.HasPerm {
			*ids = append(*ids, n.ID)
		}
		if len(n.Children) > 0 {
			collectOrgIDs(n.Children, ids)
		}
	}
}

func (c *Client) GetAdminOrgSelect(ctx context.Context, userID uint32) ([]IDNameWithAvatar, *errs.Status) {
	// 复用 GetAdminOrgSubTree 获取管理员组织树
	nodes, status := c.GetAdminOrgSubTree(ctx, userID)
	if status != nil {
		return nil, status
	}
	// 在新的事务中获取组织树（用于解析全路径名）
	var selects []IDNameWithAvatar
	if st := c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_orgs_select", err.Error())
		}
		selects = collectAdminOrgSelects(nodes, orgTree)
		return nil
	}); st != nil {
		return nil, st
	}
	return selects, nil
}

// collectAdminOrgSelects 递归遍历管理员组织树，收集所有 hasPerm=true 的组织到扁平列表，
// 并使用 orgTree.GetFullName 设置全路径名（如 "总公司 - 部门A - 小组1"）
func collectAdminOrgSelects(nodes []*AdminOrgTreeNode, orgTree *model.OrgNode) []IDNameWithAvatar {
	var out []IDNameWithAvatar
	var dfs func([]*AdminOrgTreeNode)
	dfs = func(ns []*AdminOrgTreeNode) {
		for _, n := range ns {
			if n.HasPerm {
				out = append(out, IDNameWithAvatar{
					ID:         n.ID,
					Name:       orgTree.GetFullName(n.ID),
					AvatarPath: n.AvatarPath,
				})
			}
			dfs(n.Children)
		}
	}
	dfs(nodes)
	return out
}

func (c *Client) GetAdminOrgSubTree(ctx context.Context, userID uint32) ([]*AdminOrgTreeNode, *errs.Status) {
	var ret []*AdminOrgTreeNode
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// 获取组织树
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_admin_org_sub_tree", err.Error())
		}
		// 查询用户的管理员角色对应的组织ID集合
		var userRoles []*model.UserRole
		orgRolesQuery := sqlopt.WithStatus(true).Apply(tx).Select("role_id").Table("org_roles")
		if err := sqlopt.WithUserID(userID).Apply(tx).Where("role_id IN (?)", orgRolesQuery).Find(&userRoles).Error; err != nil {
			return toErrStatus("iam_admin_org_sub_tree", err.Error())
		}
		adminOrgIDs := make(map[uint32]bool)
		for _, ur := range userRoles {
			if ur.IsAdmin {
				adminOrgIDs[ur.OrgID] = true
			}
		}
		// 如果用户是系统管理员，也加入顶级组织
		if userID == config.AdminUserID() {
			adminOrgIDs[config.TopOrgID()] = true
		}
		// 从顶级组织开始构建管理员组织树
		ret = buildAdminOrgSubTree(orgTree, adminOrgIDs)
		return nil
	})
}

func (c *Client) GetFirstClassOrgAndSubs(ctx context.Context, userID, orgID uint32) ([]IDNameWithAvatar, *errs.Status) {
	var result []IDNameWithAvatar
	return result, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_orgs_select", err.Error())
		}
		targetOrg := orgTree.GetOrg(orgID).GetFirstClassOrg()
		if targetOrg == nil {
			return nil
		}
		var dfs func(*model.OrgNode)
		dfs = func(node *model.OrgNode) {
			result = append(result, IDNameWithAvatar{
				ID:         node.GetOrgID(),
				Name:       node.GetFullName(node.GetOrgID()),
				AvatarPath: node.GetAvatarPath(),
			})
			for _, child := range node.GetSubs(node.GetOrgID()) {
				dfs(child)
			}
		}
		dfs(targetOrg)
		return nil
	})
}

func (c *Client) CreateOrg(ctx context.Context, org *model.Org) (uint32, *errs.Status) {
	if org.ID != 0 {
		return 0, toErrStatus("iam_org_create", "create org but id err")
	}
	return org.ID, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		return createOrgTx(tx, org)
	})
}

func createOrgTx(tx *gorm.DB, org *model.Org) *errs.Status {
	var roleName string
	// check parents
	if org.ParentID != 0 {
		// 正常创建组织
		if err := sqlopt.WithID(org.ParentID).Apply(tx).First(&model.Org{}).Error; err != nil {
			return toErrStatus("iam_org_create", err.Error())
		}
		// check creator
		if err := sqlopt.WithID(org.CreatorID).Apply(tx).First(&model.User{}).Error; err != nil {
			return toErrStatus("iam_org_create", err.Error())
		}
		// check name 组织名在上级组织的所有下级组织内唯一
		if err := sqlopt.SQLOptions(
			sqlopt.WithParentID(org.ParentID),
			sqlopt.WithName(org.Name),
		).Apply(tx).First(&model.Org{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				err = errors.New("already exist")
			}
			return toErrStatus("iam_org_create", err.Error())
		}
		roleName = "组织管理员"
	} else {
		// 创建系统内唯一顶级组织，此时系统内不能存在任何组织
		if err := tx.First(&model.Org{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				err = errors.New("already exist")
			}
			return toErrStatus("iam_org_create", err.Error())
		}
		// check creator
		if org.CreatorID != 0 {
			return toErrStatus("iam_org_create", "create top org but creator not empty")
		}
		roleName = "超级管理员"
	}
	// create org
	if err := tx.Create(org).Error; err != nil {
		return toErrStatus("iam_org_create", err.Error())
	}
	// create role
	roleID, err := createRole(tx, org.ID, org.CreatorID, roleName, "", "", true, nil)
	if err != nil {
		return toErrStatus("iam_org_create", err.Error())
	}
	if org.CreatorID != 0 {
		// create org user
		if err := tx.Create(&model.OrgUser{
			OrgID:  org.ID,
			UserID: org.CreatorID,
		}).Error; err != nil {
			return toErrStatus("iam_org_create", err.Error())
		}
		// create user role
		if err := tx.Create(&model.UserRole{
			OrgID:   org.ID,
			UserID:  org.CreatorID,
			RoleID:  roleID,
			IsAdmin: true,
		}).Error; err != nil {
			return toErrStatus("iam_org_create", err.Error())
		}
	}
	return nil
}

func (c *Client) UpdateOrg(ctx context.Context, org *model.Org) *errs.Status {
	if org.ID == 0 {
		return toErrStatus("iam_org_update", "update org but id err")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		orgTable := &model.Org{}
		// check parent
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(org.ID),
		).Apply(tx).First(orgTable).Error; err != nil {
			return toErrStatus("iam_org_update", err.Error())
		}
		// check name
		var orgs []*model.Org
		if err := sqlopt.SQLOptions(
			sqlopt.WithParentID(orgTable.ParentID),
			sqlopt.WithName(org.Name),
		).Apply(tx).Find(&orgs).Error; err != nil {
			return toErrStatus("iam_org_update", err.Error())
		}
		if len(orgs) > 0 {
			for _, o := range orgs {
				if o.ID != org.ID {
					return toErrStatus("iam_org_update", fmt.Sprintf("check name %v but already exist", org.Name))
				}
			}
		}
		// update org
		if err := tx.Model(org).Updates(map[string]interface{}{
			"name":        org.Name,
			"remark":      org.Remark,
			"avatar_path": org.AvatarPath,
		}).Error; err != nil {
			return toErrStatus("iam_org_update", err.Error())
		}
		return nil
	})
}

func (c *Client) DeleteOrg(ctx context.Context, orgID uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org
		if orgID == config.TopOrgID() {
			return toErrStatus("iam_org_delete", "cannot delete top org")
		}
		// org tree
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_org_delete", err.Error())
		}
		// delete org
		if err := deleteOrg(tx, orgID, orgTree); err != nil {
			return toErrStatus("iam_org_delete", err.Error())
		}
		return nil
	})
}

func deleteOrg(tx *gorm.DB, orgID uint32, orgTree *model.OrgNode) error {
	// delete sub org
	for _, sub := range orgTree.GetSubs(orgID) {
		if err := deleteOrg(tx, sub.GetOrgID(), orgTree); err != nil {
			return fmt.Errorf("delete org %v err: %v", orgID, err)
		}
	}
	// delete user role
	if err := sqlopt.WithOrgID(orgID).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("delete user role err: %v", err)
	}
	// delete org user
	if err := sqlopt.WithOrgID(orgID).Apply(tx).Delete(&model.OrgUser{}).Error; err != nil {
		return fmt.Errorf("delete org user err: %v", err)
	}
	// delete org role
	var orgRoles []*model.OrgRole
	if err := sqlopt.WithOrgID(orgID).Apply(tx).Find(&orgRoles).Error; err != nil {
		return fmt.Errorf("get org role err: %v", err)
	}
	if err := sqlopt.WithOrgID(orgID).Apply(tx).Delete(&model.OrgRole{}).Error; err != nil {
		return fmt.Errorf("delete org role err: %v", err)
	}
	// delete role
	for _, orgRole := range orgRoles {
		if err := access.RBAC0DeleteRole(tx, perm.Role(orgRole.RoleID)); err != nil {
			return fmt.Errorf("delete role %v err: %v", orgRole.OrgID, err)
		}
	}
	// delete org
	if err := sqlopt.WithID(orgID).Apply(tx).Delete(&model.Org{}).Error; err != nil {
		return fmt.Errorf("delete org %v err: %v", orgID, err)
	}
	return nil
}

func (c *Client) ChangeOrgStatus(ctx context.Context, orgID uint32, status bool) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org
		if orgID == config.TopOrgID() {
			return toErrStatus("iam_org_change_status", "cannot change top org status")
		}
		// change status
		if err := changeOrgStatus(tx, orgID, status); err != nil {
			return toErrStatus("iam_org_change_status", err.Error())
		}
		return nil
	})
}

func changeOrgStatus(tx *gorm.DB, orgID uint32, status bool) error {
	if status {
		// enable: check parent
		org := &model.Org{}
		if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
			return fmt.Errorf("change org %v status %v get org err: %v", orgID, status, err)
		}
		parent := &model.Org{}
		if err := sqlopt.WithID(org.ParentID).Apply(tx).First(parent).Error; err != nil {
			return fmt.Errorf("change org %v status %v get parent %v err: %v", orgID, status, org.ParentID, err)
		}
		if !parent.Status {
			return fmt.Errorf("change org %v status %v but parent %v status false", orgID, status, org.ParentID)
		}
	} else {
		// disable: change sub orgs status
		var orgs []*model.Org
		if err := sqlopt.WithParentID(orgID).Apply(tx).Find(&orgs).Error; err != nil {
			return fmt.Errorf("change org %v status %v get sub orgs err: %v", orgID, status, err)
		}
		// change subs
		for _, sub := range orgs {
			if err := changeOrgStatus(tx, sub.ID, status); err != nil {
				return err
			}
		}
	}
	// change status
	if err := sqlopt.WithID(orgID).Apply(tx).Model(&model.Org{}).Updates(map[string]interface{}{
		"status": status,
	}).Error; err != nil {
		return fmt.Errorf("change org %v status %v err: %v", orgID, status, err)
	}
	return nil
}

func (c *Client) AddOrgUser(ctx context.Context, orgID, userID, roleID uint32) *errs.Status {

	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org user
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithUserID(userID),
		).Apply(tx).First(&model.OrgUser{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				err = errors.New("already exist")
			}
			return toErrStatus("iam_org_user_add", util.Int2Str(orgID), util.Int2Str(userID), util.Int2Str(roleID), err.Error())
		}
		// check role
		var isAdmin bool
		if roleID != 0 {
			// check org role first
			orgRole := &model.OrgRole{}
			if err := sqlopt.SQLOptions(
				sqlopt.WithOrgID(orgID),
				sqlopt.WithRoleID(roleID),
			).Apply(tx).First(orgRole).Error; err == nil {
				// org role found
				isAdmin = orgRole.IsAdmin
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("iam_org_user_add", util.Int2Str(orgID),
					util.Int2Str(userID), util.Int2Str(roleID), err.Error())
			} else {
				// not found in org_roles, check global_roles
				globalRole := &model.GlobalRole{}
				if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
					return toErrStatus("iam_org_user_add", util.Int2Str(orgID),
						util.Int2Str(userID), util.Int2Str(roleID), err.Error())
				}
				if !globalRole.Status {
					return toErrStatus("iam_org_user_add", util.Int2Str(orgID),
						util.Int2Str(userID), util.Int2Str(roleID), "global role disabled")
				}
				// global role: isAdmin stays false
			}
		}
		// create org user
		if err := tx.Create(&model.OrgUser{
			OrgID:  orgID,
			UserID: userID,
		}).Error; err != nil {
			return toErrStatus("iam_org_user_add", util.Int2Str(orgID),
				util.Int2Str(userID), util.Int2Str(roleID), err.Error())
		}
		// create user role
		if roleID != 0 {
			if err := tx.Create(&model.UserRole{
				OrgID:   orgID,
				UserID:  userID,
				RoleID:  roleID,
				IsAdmin: isAdmin,
			}).Error; err != nil {
				return toErrStatus("iam_org_user_add", util.Int2Str(orgID),
					util.Int2Str(userID), util.Int2Str(roleID), err.Error())
			}
		}
		return nil
	})
}

func (c *Client) RemoveOrgUser(ctx context.Context, orgID, userID uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org
		if orgID == config.TopOrgID() {
			return toErrStatus("iam_org_user_remove_top")
		}
		// check creator: 不能将组织创建者从该组织中移除
		org := &model.Org{}
		if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
			return toErrStatus("iam_org_user_remove", util.Int2Str(orgID), util.Int2Str(userID), err.Error())
		}
		if org.CreatorID == userID {
			return toErrStatus("iam_org_user_remove", util.Int2Str(orgID), util.Int2Str(userID), "cannot remove org creator")
		}
		// delete user role
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithUserID(userID),
		).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
			return toErrStatus("iam_org_user_remove", util.Int2Str(orgID),
				util.Int2Str(userID), err.Error())
		}
		// delete org user
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithUserID(userID),
		).Apply(tx).Delete(&model.OrgUser{}).Error; err != nil {
			return toErrStatus("iam_org_user_remove", util.Int2Str(orgID),
				util.Int2Str(userID), err.Error())
		}
		return nil
	})
}

func (c *Client) BatchRemoveOrgUser(ctx context.Context, orgID uint32, userIDs []uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		if orgID == config.TopOrgID() {
			return toErrStatus("iam_org_user_remove_top")
		}
		// delete user role
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithUsers(userIDs),
		).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
			return toErrStatus("iam_org_user_batch_remove", util.Int2Str(orgID), err.Error())
		}
		// delete org user
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithUsers(userIDs),
		).Apply(tx).Delete(&model.OrgUser{}).Error; err != nil {
			return toErrStatus("iam_org_user_batch_remove", util.Int2Str(orgID), err.Error())
		}
		return nil
	})
}

// UserOrgPair 用户-组织二元组（消息中心受众计算的基本元素）
type UserOrgPair struct {
	UserID uint32
	OrgID  uint32
}

// UserOrgMembership 单用户在指定组织的成员关系
type UserOrgMembership struct {
	Exists   bool  // org_users 有行
	JoinedAt int64 // org_users.created_at（毫秒）
	Active   bool  // org_users.status != disable 且 users.status = true
}

// FilterValidUserOrgPairs 过滤出真实存在且有效的 (userID, orgID) 二元组。
// 有效 = org_users 有行 且 org_users.status != disable 且 users.status = true
// （判据与 IsUserOrgAdmin / checkUserIsAdmin 保持一致）。
// 实现取 orgIDs × userIDs 超集后内存过滤——二元组数受调用方上限（数百）约束，
// 超集无膨胀风险，且比 (org_id,user_id) IN ((..),(..)) 的行构造器更方言中立。
func (c *Client) FilterValidUserOrgPairs(ctx context.Context, pairs []UserOrgPair) ([]UserOrgPair, *errs.Status) {
	if len(pairs) == 0 {
		return nil, nil
	}
	orgSet := make(map[uint32]struct{}, len(pairs))
	userSet := make(map[uint32]struct{}, len(pairs))
	want := make(map[UserOrgPair]struct{}, len(pairs))
	for _, p := range pairs {
		orgSet[p.OrgID] = struct{}{}
		userSet[p.UserID] = struct{}{}
		want[p] = struct{}{}
	}
	orgIDs := make([]uint32, 0, len(orgSet))
	for id := range orgSet {
		orgIDs = append(orgIDs, id)
	}
	userIDs := make([]uint32, 0, len(userSet))
	for id := range userSet {
		userIDs = append(userIDs, id)
	}

	var rows []UserOrgPair
	if err := c.db.WithContext(ctx).
		Table("org_users").
		Joins("JOIN users ON users.id = org_users.user_id").
		Where("org_users.org_id IN ?", orgIDs).
		Where("org_users.user_id IN ?", userIDs).
		Where("org_users.status IS NULL OR org_users.status != ?", sqlopt.OrgUserStatusDisabled).
		Where("users.status = ?", true).
		Select("org_users.user_id AS user_id, org_users.org_id AS org_id").
		Scan(&rows).Error; err != nil {
		return nil, toErrStatus("iam_org_user_pairs_validate", err.Error())
	}

	ret := make([]UserOrgPair, 0, len(rows))
	for _, r := range rows {
		if _, ok := want[r]; ok {
			ret = append(ret, r)
		}
	}
	return ret, nil
}

// GetUserOrgMembership 查询单用户在指定组织的成员关系（joinedAt + 状态）。
// 供消息中心读侧 «vis» 的"新成员不追溯历史消息"屏蔽使用。
func (c *Client) GetUserOrgMembership(ctx context.Context, userID, orgID uint32) (*UserOrgMembership, *errs.Status) {
	var orgUser model.OrgUser
	err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
	).Apply(c.db.WithContext(ctx)).First(&orgUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &UserOrgMembership{Exists: false}, nil
		}
		return nil, toErrStatus("iam_org_user_membership_get", util.Int2Str(userID), err.Error())
	}
	var user model.User
	if err := sqlopt.WithID(userID).Apply(c.db.WithContext(ctx)).Select("id", "status").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &UserOrgMembership{Exists: false}, nil
		}
		return nil, toErrStatus("iam_org_user_membership_get", util.Int2Str(userID), err.Error())
	}
	return &UserOrgMembership{
		Exists:   true,
		JoinedAt: orgUser.CreatedAt,
		Active:   user.Status && orgUser.Status != sqlopt.OrgUserStatusDisabled,
	}, nil
}

func toOrgInfoTx(tx *gorm.DB, org *model.Org) (*OrgInfo, error) {
	ret := &OrgInfo{
		ID:         org.ID,
		Name:       org.Name,
		Remark:     org.Remark,
		Status:     org.Status,
		CreatedAt:  org.CreatedAt,
		AvatarPath: org.AvatarPath,
	}
	// creator
	if org.CreatorID != 0 {
		creator, err := getCreatorTx(tx, org.CreatorID)
		if err != nil {
			return nil, err
		}
		ret.Creator = creator
	}
	// user count
	var UserCount int64
	if err := sqlopt.WithOrgID(org.ID).Apply(tx).Model(&model.OrgUser{}).Count(&UserCount).Error; err != nil {
		return nil, fmt.Errorf("get org %v user count err: %v", org.ID, err)
	}
	ret.UserCount = UserCount
	// admins: find all admin user names for this org
	var adminUserRoles []*model.UserRole
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(org.ID),
		sqlopt.WithAdmin(true),
	).Apply(tx).Find(&adminUserRoles).Error; err != nil {
		return nil, fmt.Errorf("get org %v admin users err: %v", org.ID, err)
	}
	if len(adminUserRoles) > 0 {
		var adminUserIDs []uint32
		for _, ur := range adminUserRoles {
			adminUserIDs = append(adminUserIDs, ur.UserID)
		}
		var adminUsers []*model.User
		if err := sqlopt.WithIDs(adminUserIDs).Apply(tx).Find(&adminUsers).Error; err != nil {
			return nil, fmt.Errorf("get org %v admin user names err: %v", org.ID, err)
		}
		for _, u := range adminUsers {
			ret.Admins = append(ret.Admins, u.Name)
		}
	}
	return ret, nil
}

func getOrgTree(tx *gorm.DB) (*model.OrgNode, error) {
	// all org
	var orgs []*model.Org
	if err := tx.Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("get org tree all org err: %v", err)
	}
	// all org admin role
	var orgAdmins []*model.OrgRole
	if err := sqlopt.WithAdmin(true).Apply(tx).Find(&orgAdmins).Error; err != nil {
		return nil, fmt.Errorf("get org tree all org admin role err: %v", err)
	}
	return model.NewOrgTree(orgs, orgAdmins)
}

func selectOrgs(tx *gorm.DB, userID uint32, orgTree *model.OrgNode) ([]IDNameWithAvatar, error) {
	// user role
	var userRoles []*model.UserRole
	orgRolesQuery := sqlopt.WithStatus(true).Apply(tx).Select("role_id").Table("org_roles")
	if err := sqlopt.WithUserID(userID).Apply(tx).Where("role_id IN (?)", orgRolesQuery).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("get user role err: %v", err)
	}
	// user org
	var userOrgs []*model.OrgUser
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
	).Apply(tx).Find(&userOrgs).Error; err != nil {
		return nil, fmt.Errorf("get org user err: %v", err)
	}
	// select org
	var ret []IDNameWithAvatar
	for _, org := range orgTree.Select(userOrgs, userRoles) {
		ret = append(ret, IDNameWithAvatar{ID: org.ID, Name: org.Name, AvatarPath: org.AvatarPath})
	}
	return ret, nil
}

// buildAdminOrgSubTree 构建管理员组织树
// 系统管理员：以系统（顶级组织）为根节点返回完整组织树，所有节点 hasPerm=true
// 普通用户：可见范围 = 管理员组织的祖先路径 + 管理员组织及其所有下级；hasPerm=true = 管理员组织及其所有下级
func buildAdminOrgSubTree(orgTree *model.OrgNode, adminOrgIDs map[uint32]bool) []*AdminOrgTreeNode {
	if orgTree == nil {
		return nil
	}

	// 判断是否是系统管理员（adminOrgIDs 包含顶级组织）
	isSysAdmin := adminOrgIDs[orgTree.GetOrgID()]

	// 计算 adminScopeIDs: 管理员组织及其所有后代（这些 hasPerm=true）
	adminScopeIDs := make(map[uint32]bool)
	for orgID := range adminOrgIDs {
		for _, id := range orgTree.CollectDescendants(orgID) {
			adminScopeIDs[id] = true
		}
	}

	if isSysAdmin {
		// 系统管理员：以系统（顶级组织）为根节点，返回完整组织树，所有节点 hasPerm=true
		if node := buildFullOrgNode(orgTree); node != nil {
			return []*AdminOrgTreeNode{node}
		}
		return nil
	}

	// 普通用户：计算可见节点集合
	visibleIDs := make(map[uint32]bool)
	for id := range adminScopeIDs {
		visibleIDs[id] = true
	}
	// 添加管理员组织的所有祖先（不含根节点）
	for orgID := range adminOrgIDs {
		for _, ancestorID := range orgTree.GetAncestorIDs(orgID, false) {
			visibleIDs[ancestorID] = true
		}
	}

	// 从根的子节点开始构建过滤后的树
	var result []*AdminOrgTreeNode
	for _, sub := range orgTree.GetSubs(orgTree.GetOrgID()) {
		if node := buildFilteredOrgNode(sub, visibleIDs, adminScopeIDs); node != nil {
			result = append(result, node)
		}
	}
	return result
}

// buildFullOrgNode 构建完整的组织树节点（系统管理员使用），所有节点 hasPerm=true
func buildFullOrgNode(node *model.OrgNode) *AdminOrgTreeNode {
	if node == nil {
		return nil
	}
	orgID := node.GetOrgID()
	result := &AdminOrgTreeNode{
		ID:         orgID,
		Name:       node.GetOrgName(orgID),
		AvatarPath: node.GetAvatarPath(),
		HasPerm:    true,
	}
	for _, sub := range node.GetSubs(orgID) {
		if child := buildFullOrgNode(sub); child != nil {
			result.Children = append(result.Children, child)
		}
	}
	return result
}

// buildFilteredOrgNode 构建过滤后的组织树节点（普通用户使用）
// visibleIDs: 可见的组织ID集合（管理员组织 + 后代 + 祖先）
// adminScopeIDs: hasPerm=true 的组织ID集合（管理员组织 + 后代）
func buildFilteredOrgNode(node *model.OrgNode, visibleIDs, adminScopeIDs map[uint32]bool) *AdminOrgTreeNode {
	if node == nil {
		return nil
	}
	orgID := node.GetOrgID()
	// 不可见的节点直接跳过
	if !visibleIDs[orgID] {
		return nil
	}
	result := &AdminOrgTreeNode{
		ID:         orgID,
		Name:       node.GetOrgName(orgID),
		AvatarPath: node.GetAvatarPath(),
		HasPerm:    adminScopeIDs[orgID],
	}
	for _, sub := range node.GetSubs(orgID) {
		if child := buildFilteredOrgNode(sub, visibleIDs, adminScopeIDs); child != nil {
			result.Children = append(result.Children, child)
		}
	}
	return result
}
