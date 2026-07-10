package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerAdminCenter(apiV1 *gin.RouterGroup) {
	// user
	mid.Sub("admin_center").Reg(apiV1, "/user", http.MethodPost, v1.CreateUser, "创建用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/batch", http.MethodPost, v1.CreateUserByFile, "批量导入用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user", http.MethodPut, v1.ChangeUser, "编辑用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user", http.MethodDelete, v1.DeleteUser, "删除用户")
	mid.Sub("admin_center").Reg(apiV1, "/user/list", http.MethodGet, v1.GetUserList, "获取用户列表")
	mid.Sub("admin_center").Reg(apiV1, "/user/status", http.MethodPut, v1.ChangeUserStatus, "修改用户状态", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/admin/password", http.MethodPut, v1.AdminChangeUserPassword, "重置用户密码（by 管理员）", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/other/select", http.MethodGet, v1.GetOrgUserNotSelect, "获取不在组织中的用户列表（用于下拉选择）")
	mid.Sub("admin_center").Reg(apiV1, "/role/select", http.MethodGet, v1.GetRoleSelect, "获取组织角色列表（用于下拉选择）")
	mid.Sub("admin_center").Reg(apiV1, "/org/user", http.MethodPost, v1.AddOrgUser, "邀请用户加入组织", middleware.CheckOrgAdmin)
	// org
	mid.Sub("admin_center").Reg(apiV1, "/org", http.MethodPost, v1.CreateOrg, "创建下级组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org", http.MethodPut, v1.ChangeOrg, "编辑下级组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org", http.MethodDelete, v1.DeleteOrg, "删除下级组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/info", http.MethodGet, v1.GetOrgInfo, "获取组织信息")
	mid.Sub("admin_center").Reg(apiV1, "/org/list", http.MethodGet, v1.GetOrgList, "获取下级组织列表")
	mid.Sub("admin_center").Reg(apiV1, "/org/status", http.MethodPut, v1.ChangeOrgStatus, "修改下级组织状态", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/tree", http.MethodGet, v1.GetAdminOrgSubTree, "获取管理员组织及下级组织列表")
	// role
	mid.Sub("admin_center").Reg(apiV1, "/role/template", http.MethodGet, v1.GetRoleTemplate, "获取角色模板（用于创建角色）")
	mid.Sub("admin_center").Reg(apiV1, "/role", http.MethodPost, v1.CreateRole, "创建角色", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role", http.MethodPut, v1.ChangeRole, "编辑角色", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role", http.MethodDelete, v1.DeleteRole, "删除角色", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/info", http.MethodGet, v1.GetRoleInfo, "获取角色信息")
	mid.Sub("admin_center").Reg(apiV1, "/role/list", http.MethodGet, v1.GetRoleList, "获取角色列表")
	mid.Sub("admin_center").Reg(apiV1, "/role/status", http.MethodPut, v1.ChangeRoleStatus, "修改角色状态", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/users", http.MethodGet, v1.GetRoleUsers, "获取角色关联用户列表")
	mid.Sub("admin_center").Reg(apiV1, "/role/user", http.MethodDelete, v1.RemoveRoleUser, "移除角色关联用户", middleware.CheckOrgAdmin)
	// setting
	mid.Sub("admin_center.setting").Reg(apiV1, "/custom/tab", http.MethodPost, v1.UploadCustomTab, "标签页自定义配置", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.setting").Reg(apiV1, "/custom/login", http.MethodPost, v1.UploadCustomLogin, "登录页自定义配置", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.setting").Reg(apiV1, "/custom/home", http.MethodPost, v1.UploadCustomHome, "平台自定义配置", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.setting").Reg(apiV1, "/custom/general-agent", http.MethodPost, v1.UploadCustomGeneralAgent, "通用智能体自定义配置", middleware.CheckSystemAdmin)
	// oauth
	mid.Sub("admin_center.oauth").Reg(apiV1, "/oauth/app", http.MethodPost, v1.CreateOauthApp, "创建OAuth应用", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.oauth").Reg(apiV1, "/oauth/app", http.MethodDelete, v1.DeleteOauthApp, "删除OAuth应用", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.oauth").Reg(apiV1, "/oauth/app", http.MethodPut, v1.UpdateOauthApp, "修改OAuth应用信息", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.oauth").Reg(apiV1, "/oauth/app/list", http.MethodGet, v1.GetOauthAppList, "获取OAuth应用列表", middleware.CheckSystemAdmin)
	mid.Sub("admin_center.oauth").Reg(apiV1, "/oauth/app/status", http.MethodPut, v1.UpdateOauthAppStatus, "更新OAuth应用状态", middleware.CheckSystemAdmin)
}
