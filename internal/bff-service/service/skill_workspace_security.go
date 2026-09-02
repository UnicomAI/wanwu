package service

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
	"github.com/gin-gonic/gin"
)

const (
	// 限制标识符长度，因为 wga-persistent 会将其拼接到目录名中。当前 Skill
	// 标识由 UUID 或命名空间/名称组成，因此这里的限制比通用文件名更严格。
	maxSkillWorkspaceIDLength = 256

	maxWorkspaceUploadFiles       = 20
	maxWorkspaceUploadFileBytes   = 50 << 20  // 50 MiB per file
	maxWorkspaceUploadTotalBytes  = 100 << 20 // 100 MiB per request
	SkillWorkspaceUploadBodyLimit = maxWorkspaceUploadTotalBytes + (1 << 20)
)

var skillWorkspaceIDSegmentRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// validateSkillWorkspaceID 在值进入持久化存储前进行校验。存储层会拼接该值
// 构造路径，因此仅校验请求结构体并不充分（后台调用方也会直接使用存储层）。
func validateSkillWorkspaceID(id string) error {
	if id == "" {
		return fmt.Errorf("skill id is required")
	}
	if len(id) > maxSkillWorkspaceIDLength {
		return fmt.Errorf("skill id too long")
	}
	if strings.TrimSpace(id) != id {
		return fmt.Errorf("skill id contains surrounding whitespace")
	}
	if strings.ContainsRune(id, '\\') || strings.ContainsRune(id, '\x00') {
		return fmt.Errorf("skill id contains an invalid separator")
	}
	for _, r := range id {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return fmt.Errorf("skill id contains control characters")
		}
	}
	// 内置标识（例如 anthropics/docx）保留斜杠，但每个片段都必须是普通路径
	// 组件；空片段、点和双点片段一律不接受。
	segments := strings.Split(id, "/")
	if len(segments) == 0 {
		return fmt.Errorf("skill id is required")
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("skill id path segment is invalid")
		}
		if !skillWorkspaceIDSegmentRE.MatchString(segment) {
			return fmt.Errorf("skill id contains an invalid character")
		}
	}
	return nil
}

// validateSkillWorkspaceStorePath 执行第二层文件系统边界校验。它会拒绝带有
// 符号链接的基础目录、祖先目录或目标路径，避免后续 mkdir 或 git init 被重定向到配置根目录之外。
func validateSkillWorkspaceStorePath(baseDir, skillID string) error {
	if err := validateSkillWorkspaceID(skillID); err != nil {
		return err
	}
	if strings.TrimSpace(baseDir) == "" {
		return fmt.Errorf("skill workspace base directory is required")
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}
	target := filepath.Join(absBase, "thread-overwrite_"+skillID)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return err
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") || filepath.IsAbs(rel) {
		return fmt.Errorf("skill workspace path outside configured base")
	}
	if err := path_util.EnsureNoSymlinkInPath(absBase, absTarget, true); err != nil {
		return err
	}
	return nil
}

func workspaceNoPermissionError() error {
	// 响应中不包含归属查询错误或请求的标识，避免调用方利用该接口探测 Skill 是否存在。
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_no_permission")
}

// authorizeSkillWorkspaceEdit 有意限制为仅所有者可编辑。路由级 resource.skill
// 权限属于模块级权限而非对象级权限，依赖它会导致自定义 Skill 出现越权访问；
// 内置 Skill 明确为只读，即使 SkillBiz 的归属查询会将其映射到当前请求身份。
func authorizeSkillWorkspaceEdit(ctx *gin.Context, userID, orgID, skillID string) error {
	if err := validateSkillWorkspaceID(skillID); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_path_not_allowed")
	}
	if strings.TrimSpace(userID) == "" {
		return workspaceNoPermissionError()
	}
	if _, builtin := config.Cfg().AgentSkill(skillID); builtin {
		return workspaceNoPermissionError()
	}

	ownerUserID, ownerOrgID, err := OwnerInfo(ctx, constant.BizModuleResourceSkill, skillID)
	if err != nil {
		log.Warnf("[Workspace] owner lookup failed for skill %s: %v", skillID, err)
		return workspaceNoPermissionError()
	}
	if ownerUserID == "" || ownerUserID != userID {
		return workspaceNoPermissionError()
	}
	// 所有者身份是最终依据。当双方都提供组织信息时，还要求请求属于同一组织，
	// 防止伪造或过期的组织上下文跨租户访问，同时兼容不携带组织 ID 的内部调用方。
	if ownerOrgID != "" && orgID != "" && ownerOrgID != orgID {
		return workspaceNoPermissionError()
	}
	return nil
}

func isWorkspaceMetadataPath(clean string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(clean), "/") {
		if strings.EqualFold(segment, ".git") {
			return true
		}
	}
	return false
}
