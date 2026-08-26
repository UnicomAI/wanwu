package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerAgentSkill(apiV1 *gin.RouterGroup) {
	// 自定义 skill（CRUD）
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/list", http.MethodGet, v1.GetCustomSkillList, "获取自定义skill列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/detail", http.MethodGet, v1.GetCustomSkillDetail, "获取自定义skill详情")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/check", http.MethodPost, v1.CheckCustomSkill, "校验自定义skill zip包")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom", http.MethodDelete, v1.DeleteCustomSkill, "删除自定义skill")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/config", http.MethodPost, v1.CreateCustomSkillConfig, "新增自定义skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/config", http.MethodPut, v1.UpdateCustomSkillConfig, "编辑自定义skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/config", http.MethodDelete, v1.DeleteCustomSkillConfig, "删除自定义skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/version/download", http.MethodGet, v1.DownloadCustomSkillVersion, "下载自定义skill指定版本")

	// 内置 skill
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/builtin/list", http.MethodGet, v1.GetBuiltinSkillList, "获取内置skill列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/builtin/detail", http.MethodGet, v1.GetBuiltinSkillDetail, "获取内置skill详情")
	mid.Sub("resource.skill").Reg(apiV1, "/builtin/skill/download", http.MethodGet, v1.DownloadBuiltinSkill, "下载内置skill")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/builtin/config", http.MethodPost, v1.CreateBuiltinSkillConfig, "新增内置skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/builtin/config", http.MethodPut, v1.UpdateBuiltinSkillConfig, "编辑内置skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/builtin/config", http.MethodDelete, v1.DeleteBuiltinSkillConfig, "删除内置skill配置")

	// 内置/我创建/我添加的skill
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/select", http.MethodGet, v1.GetSkillSelect, "智能体skills下拉列表")

	// 资源库 - 我添加的 skill（acquired skill）
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/list", http.MethodGet, v1.GetAcquiredSkillList, "获取我添加的skill列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired", http.MethodDelete, v1.DeleteAcquiredSkill, "删除我添加的skill")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/detail", http.MethodGet, v1.GetAcquiredSkillDetail, "获取我添加的skill详情")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/download", http.MethodGet, v1.DownloadAcquiredSkill, "下载我添加的skill")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/version/list", http.MethodGet, v1.GetAcquiredSkillVersionList, "获取我添加的skill历史版本")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/config", http.MethodPost, v1.CreateAcquiredSkillConfig, "新增我添加的skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/config", http.MethodPut, v1.UpdateAcquiredSkillConfig, "编辑我添加的skill配置")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/acquired/config", http.MethodDelete, v1.DeleteAcquiredSkillConfig, "删除我添加的skill配置")

	// workspace 文件管理
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/files", http.MethodGet, v1.GetSkillWorkspaceFiles, "获取Skill工作区文件列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/file", http.MethodGet, v1.GetSkillWorkspaceFile, "读取Skill工作区文件")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/download", http.MethodGet, v1.DownloadSkillWorkspace, "下载Skill工作区文件")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/file", http.MethodPut, v1.UpdateSkillWorkspaceFile, "更新Skill工作区文件")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/file", http.MethodDelete, v1.DeleteSkillWorkspaceFile, "删除Skill工作区文件")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/search", http.MethodPost, v1.SearchSkillWorkspace, "搜索Skill工作区内容")

	// skill 内容查询（已发布返回最新发布版本，从未发布回退草稿工作区）
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/content/files", http.MethodGet, v1.GetSkillContentFiles, "获取Skill内容文件列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/content/file", http.MethodGet, v1.GetSkillContentFile, "读取Skill内容文件")

	// git 版本管理
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/log", http.MethodGet, v1.GetSkillWorkspaceGitLog, "获取Skill工作区Git提交历史")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/diff", http.MethodGet, v1.GetSkillWorkspaceGitDiff, "获取Skill工作区Git diff")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/file", http.MethodGet, v1.GetSkillWorkspaceGitFile, "获取Skill工作区Git历史文件内容")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/file-diff", http.MethodGet, v1.GetSkillWorkspaceGitFileDiff, "获取Skill工作区Git单文件diff")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/status", http.MethodGet, v1.GetGitStatus, "获取Skill工作区Git状态")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/add", http.MethodPost, v1.GitAdd, "暂存Skill工作区文件")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/reset", http.MethodPost, v1.GitReset, "取消暂存Skill工作区文件")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/restore", http.MethodPost, v1.GitRestore, "恢复Skill工作区到指定commit")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/discard", http.MethodPost, v1.GitDiscardWorkingTree, "放弃Skill工作区未暂存更改")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/commit", http.MethodPost, v1.GitCommit, "提交Skill工作区变更")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/diff-working", http.MethodGet, v1.GetGitDiffWorkingTree, "获取Skill工作区未暂存diff")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/workspace/git/diff-staged", http.MethodGet, v1.GetGitDiffStaged, "获取Skill工作区已暂存diff")
}
