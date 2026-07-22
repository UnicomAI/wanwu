package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

var knowledgeBiz = &middleware.AdminCenterBiz{
	BizId:   "knowledgeId",
	BizType: constant.BizModuleResourceKnowledge,
}

var workflowBiz = &middleware.AdminCenterBiz{
	BizId:   "workflowId",
	BizType: constant.BizModuleAppWorkflow,
}

var skillBiz = &middleware.AdminCenterBiz{
	BizId:   "skillId",
	BizType: constant.BizModuleResourceSkill,
}

var modelBiz = &middleware.AdminCenterBiz{
	BizId:   "modelId",
	BizType: constant.BizModuleModel,
}

var ragBiz = &middleware.AdminCenterBiz{
	BizId:   "ragId",
	BizType: constant.BizModuleAppRag,
}

var mcpBiz = &middleware.AdminCenterBiz{
	BizId:   "mcpId",
	BizType: constant.BizModuleResourceMCP,
}
var toolBiz = &middleware.AdminCenterBiz{
	BizId:   "customToolId",
	BizType: constant.BizModuleResourceTool,
}
var promptBiz = &middleware.AdminCenterBiz{
	BizId:   "customPromptId",
	BizType: constant.BizModuleResourcePrompt,
}

var assistantBiz = &middleware.AdminCenterBiz{
	BizId:   "assistantId",
	BizType: constant.BizModuleAppAgent,
}

var sensitiveWordBiz = &middleware.AdminCenterBiz{
	BizId:   "tableId",
	BizType: constant.BizModuleResourceSafety,
}

func registerAdminCenter(apiV1 *gin.RouterGroup) {
	// user
	mid.Sub("admin_center").Reg(apiV1, "/user", http.MethodPost, v1.CreateUser, "创建用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/batch", http.MethodPost, v1.CreateUserByFile, "批量导入用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/batch", http.MethodDelete, v1.BatchDeleteUser, "批量删除用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user", http.MethodPut, v1.ChangeUser, "编辑用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user", http.MethodDelete, v1.DeleteUser, "删除用户", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/list", http.MethodGet, v1.GetUserList, "获取用户列表", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/status", http.MethodPut, v1.ChangeUserStatus, "修改用户状态", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/user/admin/password", http.MethodPut, v1.AdminChangeUserPassword, "重置用户密码（by 管理员）", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/other/select", http.MethodGet, v1.GetOrgUserNotSelect, "获取不在组织中的用户列表（用于下拉选择）", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/select", http.MethodGet, v1.GetRoleSelect, "获取组织角色列表（用于下拉选择）", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/user", http.MethodPost, v1.AddOrgUser, "邀请用户加入组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/users", http.MethodPost, v1.GetUsersByOrgIDs, "根据组织ID列表获取用户ID")
	// org
	mid.Sub("admin_center").Reg(apiV1, "/org", http.MethodPost, v1.CreateOrg, "创建下级组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org", http.MethodPut, v1.ChangeOrg, "编辑下级组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org", http.MethodDelete, v1.DeleteOrg, "删除下级组织", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/info", http.MethodGet, v1.GetOrgInfo, "获取组织信息", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/list", http.MethodGet, v1.GetOrgList, "获取下级组织列表", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/status", http.MethodPut, v1.ChangeOrgStatus, "修改下级组织状态", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/org/tree", http.MethodGet, v1.GetAdminOrgSubTree, "获取管理员组织及下级组织列表（包含上级组织）")
	mid.Sub("admin_center").Reg(apiV1, "/org/admin/select", http.MethodGet, v1.GetAdminOrgSelect, "获取管理员组织及下级组织列表（不含上级组织）")
	// role
	mid.Sub("admin_center").Reg(apiV1, "/role/template", http.MethodGet, v1.GetRoleTemplate, "获取角色模板（用于创建角色）", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role", http.MethodPost, v1.CreateRole, "创建角色", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role", http.MethodPut, v1.ChangeRole, "编辑角色", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role", http.MethodDelete, v1.DeleteRole, "删除角色", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/info", http.MethodGet, v1.GetRoleInfo, "获取角色信息", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/list", http.MethodGet, v1.GetRoleList, "获取角色列表", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/status", http.MethodPut, v1.ChangeRoleStatus, "修改角色状态", middleware.CheckOrgAdmin)
	mid.Sub("admin_center").Reg(apiV1, "/role/users", http.MethodGet, v1.GetRoleUsers, "获取角色关联用户列表", middleware.CheckOrgAdmin)
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

	apiAdminCenter := apiV1.Group("/admin/center")
	// knowledge
	mid.Sub("admin_center.knowledge").Reg(apiAdminCenter, "/knowledge/page/list", http.MethodPost, v1.AdminKnowledgePageList, "知识库全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, knowledgeBiz))
	mid.Sub("admin_center.knowledge").Reg(apiAdminCenter, "/knowledge/base", http.MethodPost, v1.AdminKnowledgeBase, "知识库基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, knowledgeBiz))
	mid.Sub("admin_center.knowledge").Reg(apiAdminCenter, "/knowledge/file/list", http.MethodPost, v1.AdminKnowledgeFileList, "知识库文档列表", middleware.AuthAdminCenter(middleware.AdminBizCheck, knowledgeBiz))
	mid.Sub("admin_center.knowledge").Reg(apiAdminCenter, "/knowledge/qa/pair/list", http.MethodPost, v1.AdminKnowledgeQAPairList, "问答对列表", middleware.AuthAdminCenter(middleware.AdminBizCheck, knowledgeBiz))
	mid.Sub("admin_center.knowledge").Reg(apiAdminCenter, "/knowledge/file/detail", http.MethodPost, v1.AdminKnowledgeFileDetail, "知识库文档详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, knowledgeBiz))
	// workflow
	mid.Sub("admin_center.workflow").Reg(apiAdminCenter, "/workflow/page/list", http.MethodPost, v1.AdminWorkflowPageList, "工作流分页列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, workflowBiz))
	// skill
	mid.Sub("admin_center.skill").Reg(apiAdminCenter, "/skill/page/list", http.MethodPost, v1.AdminSkillPageList, "skill分页列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, skillBiz))
	mid.Sub("admin_center.skill").Reg(apiAdminCenter, "/skill/base", http.MethodPost, v1.AdminSkillBase, "skill详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, skillBiz))
	mid.Sub("admin_center.skill").Reg(apiAdminCenter, "/skill/detail", http.MethodPost, v1.AdminSkillDetail, "skill详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, skillBiz))
	mid.Sub("admin_center.skill").Reg(apiAdminCenter, "/skill/version/list", http.MethodPost, v1.AdminSkillVersionList, "已发布skill版本列表", middleware.AuthAdminCenter(middleware.AdminBizCheck, skillBiz))
	//model 模型管理
	mid.Sub("admin_center.model").Reg(apiAdminCenter, "/model/page/list", http.MethodPost, v1.AdminModelPageList, "模型全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, modelBiz))
	mid.Sub("admin_center.model").Reg(apiAdminCenter, "/model/base", http.MethodPost, v1.AdminModelBase, "模型基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, modelBiz))
	mid.Sub("admin_center.model").Reg(apiAdminCenter, "/model/detail", http.MethodPost, v1.AdminModelDetail, "模型详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, modelBiz))
	//rag 知识问答
	mid.Sub("admin_center.rag").Reg(apiAdminCenter, "/rag/page/list", http.MethodPost, v1.AdminRagPageList, "知识问答全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, ragBiz))
	mid.Sub("admin_center.rag").Reg(apiAdminCenter, "/rag/base", http.MethodPost, v1.AdminRagBase, "知识问答基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, ragBiz))
	mid.Sub("admin_center.rag").Reg(apiAdminCenter, "/rag/detail", http.MethodPost, v1.AdminRagDetail, "知识问答详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, ragBiz))
	// mcp
	mid.Sub("admin_center.mcp").Reg(apiAdminCenter, "/mcp/page/list", http.MethodPost, v1.AdminMCPPageList, "MCP全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, mcpBiz))
	mid.Sub("admin_center.mcp").Reg(apiAdminCenter, "/mcp/base", http.MethodPost, v1.AdminMCPBase, "MCP基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, mcpBiz))
	mid.Sub("admin_center.mcp").Reg(apiAdminCenter, "/mcp/custom/detail", http.MethodPost, v1.AdminMCPDetail, "导入MCP详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, mcpBiz))
	mid.Sub("admin_center.mcp").Reg(apiAdminCenter, "/mcp/server/detail", http.MethodPost, v1.AdminMCPServerDetail, "创建MCP详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, mcpBiz))
	mid.Sub("admin_center.mcp").Reg(apiAdminCenter, "/mcp/tool/list", http.MethodPost, v1.AdminMCPToolList, "MCP工具列表", middleware.AuthAdminCenter(middleware.AdminBizCheck, mcpBiz))
	// tool
	mid.Sub("admin_center.tool").Reg(apiAdminCenter, "/tool/page/list", http.MethodPost, v1.AdminToolPageList, "工具全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, toolBiz))
	mid.Sub("admin_center.tool").Reg(apiAdminCenter, "/tool/base", http.MethodPost, v1.AdminToolBase, "工具基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, toolBiz))
	mid.Sub("admin_center.tool").Reg(apiAdminCenter, "/tool/detail", http.MethodPost, v1.AdminToolDetail, "工具详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, toolBiz))
	// prompt
	mid.Sub("admin_center.prompt").Reg(apiAdminCenter, "/prompt/page/list", http.MethodPost, v1.AdminPromptPageList, "提示词全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, promptBiz))
	mid.Sub("admin_center.prompt").Reg(apiAdminCenter, "/prompt/base", http.MethodPost, v1.AdminPromptBase, "提示词基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, promptBiz))
	mid.Sub("admin_center.prompt").Reg(apiAdminCenter, "/prompt/detail", http.MethodPost, v1.AdminPromptDetail, "提示词详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, promptBiz))
	// assistant
	mid.Sub("admin_center.assistant").Reg(apiAdminCenter, "/assistant/page/list", http.MethodPost, v1.AdminAssistantPageList, "智能体全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, assistantBiz))
	mid.Sub("admin_center.assistant").Reg(apiAdminCenter, "/assistant/base", http.MethodPost, v1.AdminAssistantBase, "智能体基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, assistantBiz))
	mid.Sub("admin_center.assistant").Reg(apiAdminCenter, "/assistant/detail", http.MethodPost, v1.AdminAssistantDetail, "智能体业务详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, assistantBiz))
	// sensitive
	mid.Sub("admin_center.sensitive").Reg(apiAdminCenter, "/sensitive/page/list", http.MethodPost, v1.AdminSensitiveWordPageList, "敏感词全局列表", middleware.AuthAdminCenter(middleware.AdminOrgCheck, sensitiveWordBiz))
	mid.Sub("admin_center.sensitive").Reg(apiAdminCenter, "/sensitive/base", http.MethodPost, v1.AdminSensitiveWordBase, "敏感词基础信息", middleware.AuthAdminCenter(middleware.AdminBizCheck, sensitiveWordBiz))
	mid.Sub("admin_center.sensitive").Reg(apiAdminCenter, "/sensitive/detail", http.MethodPost, v1.AdminSensitiveWordDetail, "敏感词业务详情", middleware.AuthAdminCenter(middleware.AdminBizCheck, sensitiveWordBiz))
}
