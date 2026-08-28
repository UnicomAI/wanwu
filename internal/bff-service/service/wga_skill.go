package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// --- internal wga skill ---

// checkWgaSkillConfig 校验wga Skill配置（用于更新配置）
func checkWgaSkillConfig(ctx *gin.Context, userId, orgId string, skillList []*assistant_service.WgaConfigSkill) ([]*assistant_service.WgaConfigSkill, error) {
	if len(skillList) == 0 {
		return skillList, nil
	}

	var customSkillIds []string
	var acquiredSkillIds []string
	for _, s := range skillList {
		switch s.SkillType {
		case constant.SkillTypeCustom:
			customSkillIds = append(customSkillIds, s.SkillId)
		case constant.SkillTypeAcquired:
			acquiredSkillIds = append(acquiredSkillIds, s.SkillId)
		default:
			log.Warnf("[wga] invalid skill type: %s, skillId: %s, skipping", s.SkillType, s.SkillId)
		}
	}

	// 校验 custom 技能
	validCustomIds, err := getValidSkillIds(ctx, customSkillIds)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "custom skill not found")
	}

	// 校验 acquired 技能
	acquiredSkillMap, err := getAcquiredSkillByIDMap(ctx, acquiredSkillIds)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "acquired skill not found")
	}

	// 过滤出有效的技能列表
	var validSkillList []*assistant_service.WgaConfigSkill
	for _, s := range skillList {
		switch s.SkillType {
		case constant.SkillTypeCustom:
			if validCustomIds[s.SkillId] {
				validSkillList = append(validSkillList, s)
			} else {
				log.Warnf("[wga] custom skill not found: %s, userId: %s, orgId: %s, skipping", s.SkillId, userId, orgId)
			}
		case constant.SkillTypeAcquired:
			if _, exists := acquiredSkillMap[s.SkillId]; exists {
				validSkillList = append(validSkillList, s)
			} else {
				log.Warnf("[wga] acquired skill not found: %s, userId: %s, orgId: %s, skipping", s.SkillId, userId, orgId)
			}
		}
	}
	return validSkillList, nil
}

func buildWgaSkillOptions(ctx *gin.Context, userId, orgId, threadId, runId string, skillList []*assistant_service.WgaConfigSkill) ([]wga_option.Option, error) {
	if len(skillList) == 0 {
		return nil, nil
	}

	var customSkillIds []string
	var acquiredSkillIds []string
	for _, s := range skillList {
		switch s.SkillType {
		case constant.SkillTypeCustom:
			customSkillIds = append(customSkillIds, s.SkillId)
		case constant.SkillTypeAcquired:
			acquiredSkillIds = append(acquiredSkillIds, s.SkillId)
		default:
			return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("invalid skill type: %s", s.SkillType))
		}
	}

	var opts []wga_option.Option

	// 处理 custom 技能
	if len(customSkillIds) > 0 {
		customOpts, err := buildCustomSkillOptions(ctx, userId, orgId, threadId, runId, customSkillIds)
		if err != nil {
			return nil, err
		}
		opts = append(opts, customOpts...)
	}

	// 处理 acquired 技能
	if len(acquiredSkillIds) > 0 {
		acquiredOpts, err := buildAcquiredSkillOptions(ctx, userId, orgId, threadId, runId, acquiredSkillIds)
		if err != nil {
			return nil, err
		}
		opts = append(opts, acquiredOpts...)
	}

	return opts, nil
}

// buildCustomSkillOptions 构建自定义技能选项
func buildCustomSkillOptions(ctx *gin.Context, userId, orgId, threadId, runId string, skillIds []string) ([]wga_option.Option, error) {
	resp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: skillIds,
	})
	if err != nil {
		return nil, err
	}

	var opts []wga_option.Option
	for _, s := range resp.SkillDetails {
		var skillTempDir string

		if s.ObjectPath != "" {
			// 已发布的 skill：从 minio 下载
			skillTempDir, err = downloadAndUnzipSkill(ctx, s.SkillId, s.ObjectPath, threadId, runId)
			if err != nil {
				return nil, err
			}
		} else {
			// 未发布的 skill（草稿）：从 workspace 目录获取
			skillDir, err := findFirstCustomSkillDir(s.SkillId)
			if err != nil {
				return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to find skill workspace for %s: %v", s.SkillId, err))
			}
			if err := ensureNoSymlink(skillDir); err != nil {
				return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("custom skill %s contains symlink: %v", s.SkillId, err))
			}
			skillTempDir = skillDir
		}

		variables, err := getCustomSkillVariables(ctx, s.SkillId)
		if err != nil {
			return nil, err
		}
		opts = append(opts, wga_option.WithSkill(wga_option.Skill{Dir: skillTempDir, Variables: getWgaSkillVariables(variables)}))
	}

	return opts, nil
}

// buildAcquiredSkillOptions 构建已添加技能选项
func buildAcquiredSkillOptions(ctx *gin.Context, userId, orgId, threadId, runId string, skillIds []string) ([]wga_option.Option, error) {
	acquiredSkillMap, err := getAcquiredSkillByIDMap(ctx, skillIds)
	if err != nil {
		return nil, err
	}

	var opts []wga_option.Option
	for _, acquiredSkillId := range skillIds {
		skill, exists := acquiredSkillMap[acquiredSkillId]
		if !exists {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("acquired skill not found: %s", acquiredSkillId))
		}

		publish := skill.GetSkill()
		if publish == nil || publish.GetObjectPath() == "" {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("acquired skill has no published package: %s", acquiredSkillId))
		}

		// 从 minio 下载
		skillTempDir, err := downloadAndUnzipSkill(ctx, acquiredSkillId, publish.GetObjectPath(), threadId, runId)
		if err != nil {
			return nil, err
		}

		variables, err := getAcquiredSkillVariables(ctx, acquiredSkillId)
		if err != nil {
			return nil, err
		}
		opts = append(opts, wga_option.WithSkill(wga_option.Skill{Dir: skillTempDir, Variables: getWgaSkillVariables(variables)}))
	}

	return opts, nil
}

// downloadAndUnzipSkill 从 minio 下载技能包并解压到临时目录
func downloadAndUnzipSkill(ctx *gin.Context, skillId, objectPath, threadId, runId string) (string, error) {
	skillUrl, _ := url.JoinPath("http://", config.Cfg().Minio.Endpoint, objectPath)
	b, skillZipName, err := minio_util.DownloadFile(ctx.Request.Context(), skillUrl)
	if err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to download skill file from %s: %v", skillUrl, err))
	}
	skillTempDir := filepath.Join(os.TempDir(), "wga", threadId, runId, "skills", skillId)
	if err := os.MkdirAll(skillTempDir, 0755); err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to create skill temp dir %s: %v", skillTempDir, err))
	}
	skillZipPath := filepath.Join(skillTempDir, skillZipName)
	if err := os.WriteFile(skillZipPath, b, 0644); err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to write skill zip %s: %v", skillZipPath, err))
	}
	if _, err := util.UnzipDir(ctx.Request.Context(), skillZipPath, skillTempDir); err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to unzip skill %s: %v", skillZipPath, err))
	}
	if err := util.DeleteFile(skillZipPath); err != nil {
		log.Warnf("failed to delete skill zip file %s: %v", skillZipPath, err)
	}
	return skillTempDir, nil
}

// getValidSkillIds 批量获取有效的Skill ID映射
func getValidSkillIds(ctx *gin.Context, skillIds []string) (map[string]bool, error) {
	if len(skillIds) == 0 {
		return make(map[string]bool), nil
	}
	resp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: skillIds,
	})
	if err != nil {
		return nil, err
	}
	validIds := make(map[string]bool)
	for _, s := range resp.SkillDetails {
		validIds[s.SkillId] = true
	}
	return validIds, nil
}

func getWgaSkillVariables(variables []*mcp_service.Variable) []wga_option.SkillVariable {
	var ret []wga_option.SkillVariable
	for _, v := range variables {
		if v == nil {
			continue
		}
		ret = append(ret, wga_option.SkillVariable{
			Name:          v.Name,
			Description:   v.Desc,
			VariableKey:   v.VariableKey,
			VariableValue: v.VariableValue,
		})
	}
	return ret
}

// buildWgaSkillVariablesMessage 构建技能变量的系统消息，用于 Skill Preview Agent 模式
func buildWgaSkillVariablesMessage(customSkill *mcp_service.CustomSkill, variables []*mcp_service.Variable) string {
	var buf strings.Builder

	// 技能基本信息
	buf.WriteString("# 当前工作空间技能信息\n\n")
	if customSkill.Name != "" {
		fmt.Fprintf(&buf, "**技能名称**: %s\n", customSkill.Name)
	}
	if customSkill.Desc != "" {
		fmt.Fprintf(&buf, "**技能描述**: %s\n", customSkill.Desc)
	}
	buf.WriteString("\n")

	// 用户定义的变量
	if len(variables) > 0 {
		buf.WriteString("## 用户定义的变量\n\n")
		buf.WriteString("以下变量已为当前技能配置：\n\n")
		for _, v := range variables {
			if v == nil {
				continue
			}
			// 转义反引号
			escapedKey := strings.ReplaceAll(v.VariableKey, "`", "\\`")
			escapedValue := strings.ReplaceAll(v.VariableValue, "`", "\\`")

			if v.Desc != "" {
				escapedDesc := strings.ReplaceAll(v.Desc, "`", "\\`")
				fmt.Fprintf(&buf, "- **%s** (%s): `%s` = `%s`\n",
					v.Name, escapedDesc, escapedKey, escapedValue)
			} else {
				fmt.Fprintf(&buf, "- **%s**: `%s` = `%s`\n",
					v.Name, escapedKey, escapedValue)
			}
		}
	}

	return buf.String()
}

// buildBuiltinSkillMessage 获取用户所有内置技能的配置变量，构建系统消息
func buildBuiltinSkillMessage(ctx *gin.Context, userId, orgId string) (*schema.Message, error) {
	resp, err := mcp.GetBuiltinSkillsVars(ctx.Request.Context(), &mcp_service.GetBuiltinSkillsVarsReq{
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Skills) == 0 {
		return nil, nil
	}

	skillCfgMap := make(map[string]*config.SkillsConfig)
	for _, s := range config.Cfg().AgentSkills {
		skillCfgMap[s.SkillId] = s
	}

	var buf strings.Builder
	buf.WriteString("# 内置技能配置变量\n\n")
	for _, skillVars := range resp.Skills {
		if skillVars == nil || len(skillVars.Variables) == 0 {
			continue
		}
		if cfg, ok := skillCfgMap[skillVars.SkillId]; ok {
			if cfg.Name != "" {
				fmt.Fprintf(&buf, "## %s(%s)\n", cfg.Name, skillVars.SkillId)
			}
			if cfg.Desc != "" {
				fmt.Fprintf(&buf, "%s\n\n", cfg.Desc)
			}
		}
		buf.WriteString("以下变量已为当前技能配置：\n\n")
		for _, v := range skillVars.Variables {
			if v == nil {
				continue
			}
			escapedKey := strings.ReplaceAll(v.VariableKey, "`", "\\`")
			escapedValue := strings.ReplaceAll(v.VariableValue, "`", "\\`")
			if v.Desc != "" {
				escapedDesc := strings.ReplaceAll(v.Desc, "`", "\\`")
				fmt.Fprintf(&buf, "- **%s** (%s): `%s` = `%s`\n", v.Name, escapedDesc, escapedKey, escapedValue)
			} else {
				fmt.Fprintf(&buf, "- **%s**: `%s` = `%s`\n", v.Name, escapedKey, escapedValue)
			}
		}
		buf.WriteString("\n")
	}

	return &schema.Message{
		Role:    schema.System,
		Content: buf.String(),
	}, nil
}
