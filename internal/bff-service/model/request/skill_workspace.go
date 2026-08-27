package request

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	gitutil "github.com/UnicomAI/wanwu/pkg/git-util"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
)

// checkRelPath 校验相对路径安全性（统一实现，避免各 Check() 重复内联）。
func checkRelPath(p string) error {
	_, err := path_util.CleanRelPath(p, false)
	return err
}

// checkWorkspaceEntryPath 校验工作区条目路径，并禁止直接操作根目录。
func checkWorkspaceEntryPath(p string) error {
	cleanPath, err := path_util.CleanRelPath(p, false)
	if err != nil {
		return err
	}
	if cleanPath == "." {
		return fmt.Errorf("path must not be workspace root")
	}
	return nil
}

// maxWriteFileSize 与 service 层 maxFileSize 保持一致（1MB）。
const maxWriteFileSize = 1 * 1024 * 1024
const maxPathLength = 512

// 变更请求统一使用相对于工作区的路径。
// 服务层会再次执行归属、符号链接和文件系统校验，
// 因为这些结构体也会被非 HTTP 调用方使用。
type CreateSkillWorkspaceFileReq struct {
	CustomSkillID string `json:"customSkillId" validate:"required"`
	Path          string `json:"path"`
	Content       string `json:"content"`
}

type CreateSkillWorkspaceDirectoryReq struct {
	CustomSkillID string `json:"customSkillId" validate:"required"`
	Path          string `json:"path"`
}

type RenameSkillWorkspaceFileReq struct {
	CustomSkillID string `json:"customSkillId" validate:"required"`
	Path          string `json:"path"`
	NewName       string `json:"newName"`
}

// --- 文件管理请求结构体 ---
// GET 接口使用 form tag（query parameter）
// POST/PUT 接口使用 json tag（request body）

type GetSkillWorkspaceFilesReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
}

type GetSkillWorkspaceFileReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	Path          string `form:"path"`                              // 文件路径（相对 workspace 根目录）
}

type DownloadSkillWorkspaceReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"`
	Path          string `form:"path"`
}

type DeleteSkillWorkspaceFileReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"`
	Path          string `form:"path"`
}

type UpdateSkillWorkspaceFileReq struct {
	CustomSkillID string `json:"customSkillId" validate:"required"` // Skill ID（request body）
	Path          string `json:"path"`                              // 文件路径（相对 workspace 根目录）
	Content       string `json:"content"`                           // 文件内容（全量覆盖）
}

type SearchSkillWorkspaceReq struct {
	CustomSkillID  string `json:"customSkillId" validate:"required"` // Skill ID（request body）
	Keyword        string `json:"keyword"`                           // 搜索关键词
	CaseSensitive  bool   `json:"caseSensitive"`                     // 区分大小写
	WholeWord      bool   `json:"wholeWord"`                         // 全字匹配
	UseRegex       bool   `json:"useRegex"`                          // 将 keyword 作为正则
	IncludePattern string `json:"includePattern"`                    // 包含文件模式 (glob)
	ExcludePattern string `json:"excludePattern"`                    // 排除文件模式 (glob)

	// CompiledRegex 由 Check() 在 UseRegex=true 时编译并缓存，service 层复用以避免二次编译（G3）。
	// json:"-" 确保不参与序列化，不暴露到 swagger。
	CompiledRegex *regexp.Regexp `json:"-"`
}

// Check 校验获取工作区文件树请求。
func (r *GetSkillWorkspaceFilesReq) Check() error {
	return nil
}

// Check 校验读取工作区文件请求。
func (r *GetSkillWorkspaceFileReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	return checkRelPath(r.Path)
}

// Check 校验下载工作区条目请求。
func (r *DownloadSkillWorkspaceReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	return checkWorkspaceEntryPath(r.Path)
}

// Check 校验删除工作区条目请求。
func (r *DeleteSkillWorkspaceFileReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	return checkWorkspaceEntryPath(r.Path)
}

// Check 校验更新工作区文件请求。
func (r *UpdateSkillWorkspaceFileReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if err := checkRelPath(r.Path); err != nil {
		return err
	}
	// G7: 写入侧限制文件大小和路径长度
	if len(r.Path) > maxPathLength {
		return fmt.Errorf("path too long (max %d characters)", maxPathLength)
	}
	if len(r.Content) > maxWriteFileSize {
		return fmt.Errorf("content too large (max 1MB)")
	}
	return nil
}

func (r *CreateSkillWorkspaceFileReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if err := checkWorkspaceEntryPath(r.Path); err != nil {
		return err
	}
	if len(r.Path) > maxPathLength {
		return fmt.Errorf("path too long (max %d characters)", maxPathLength)
	}
	if len(r.Content) > maxWriteFileSize {
		return fmt.Errorf("content too large (max 1MB)")
	}
	return nil
}

func (r *CreateSkillWorkspaceDirectoryReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if err := checkWorkspaceEntryPath(r.Path); err != nil {
		return err
	}
	if len(r.Path) > maxPathLength {
		return fmt.Errorf("path too long (max %d characters)", maxPathLength)
	}
	return nil
}

func (r *RenameSkillWorkspaceFileReq) Check() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if err := checkWorkspaceEntryPath(r.Path); err != nil {
		return err
	}
	if r.NewName == "" {
		return fmt.Errorf("newName is required")
	}
	if len(r.Path) > maxPathLength || len(r.NewName) > maxPathLength {
		return fmt.Errorf("path or newName too long (max %d characters)", maxPathLength)
	}
	if strings.HasSuffix(r.NewName, ".") || strings.HasSuffix(r.NewName, " ") || strings.ContainsAny(r.NewName, `/\\:`) {
		return fmt.Errorf("newName must be a safe basename")
	}
	return nil
}

// Check 校验工作区搜索请求，并在正则模式下预编译表达式。
func (r *SearchSkillWorkspaceReq) Check() error {
	if r.Keyword == "" {
		return fmt.Errorf("keyword is required")
	}
	if len(r.Keyword) > 1000 {
		return fmt.Errorf("keyword too long (max 1000 characters)")
	}
	if r.UseRegex {
		if len(r.Keyword) > 500 {
			return fmt.Errorf("regex pattern too long (max 500 characters)")
		}
		flags := ""
		if !r.CaseSensitive {
			flags = "(?i)"
		}
		compiled, err := regexp.Compile(flags + r.Keyword)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
		r.CompiledRegex = compiled // G3: 缓存，service 层直接复用
	}
	return nil
}

// --- Git 相关请求结构体 ---
// GET 接口使用 form tag（query parameter）
// POST 接口使用 json tag（request body）

type GetSkillWorkspaceGitLogReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	Count         int    `form:"count"`                             // 默认 50，最大 1000
}

type GetSkillWorkspaceGitDiffReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	FromCommit    string `form:"fromCommit"`                        // 起始 commit（默认 HEAD~1）
	ToCommit      string `form:"toCommit"`                          // 结束 commit（默认 HEAD）
}

type GetSkillWorkspaceGitFileReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	CommitHash    string `form:"commitHash"`                        // 目标 commit（默认 HEAD）
	FilePath      string `form:"filePath"`                          // 文件路径（相对 workspace 根目录）
}

type GetSkillWorkspaceGitFileDiffReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	FromCommit    string `form:"fromCommit"`                        // 起始 commit（默认 HEAD~1）
	ToCommit      string `form:"toCommit"`                          // 结束 commit（默认 HEAD）
	FilePath      string `form:"filePath"`                          // 文件路径（相对 workspace 根目录）
}

// Check 校验 Git 日志请求参数。
func (r *GetSkillWorkspaceGitLogReq) Check() error {
	if r.Count < 0 {
		return fmt.Errorf("count must be non-negative")
	}
	if r.Count > 1000 {
		return fmt.Errorf("count too large (max 1000)")
	}
	return nil
}

// Check 校验 Git diff 请求参数。
func (r *GetSkillWorkspaceGitDiffReq) Check() error {
	if err := gitutil.ValidateCommitRef(r.FromCommit); err != nil {
		return fmt.Errorf("invalid fromCommit: %v", err)
	}
	if err := gitutil.ValidateCommitRef(r.ToCommit); err != nil {
		return fmt.Errorf("invalid toCommit: %v", err)
	}
	return nil
}

// Check 校验读取 Git 历史文件请求参数。
func (r *GetSkillWorkspaceGitFileReq) Check() error {
	if r.FilePath == "" {
		return fmt.Errorf("filePath is required")
	}
	if err := checkRelPath(r.FilePath); err != nil {
		return err
	}
	if err := gitutil.ValidateCommitRef(r.CommitHash); err != nil {
		return fmt.Errorf("invalid commitHash: %v", err)
	}
	return nil
}

// Check 校验 Git 单文件 diff 请求参数。
func (r *GetSkillWorkspaceGitFileDiffReq) Check() error {
	if r.FilePath == "" {
		return fmt.Errorf("filePath is required")
	}
	if err := checkRelPath(r.FilePath); err != nil {
		return err
	}
	if err := gitutil.ValidateCommitRef(r.FromCommit); err != nil {
		return fmt.Errorf("invalid fromCommit: %v", err)
	}
	if err := gitutil.ValidateCommitRef(r.ToCommit); err != nil {
		return fmt.Errorf("invalid toCommit: %v", err)
	}
	return nil
}

type GitStatusReq struct {
	CustomSkillID string `form:"customSkillId" validate:"required"` // Skill ID（query parameter）
	FilePath      string `form:"filePath"`                          // 可选：限制 diff 到单个文件
}

type GitAddReq struct {
	CustomSkillID string   `json:"customSkillId" validate:"required"` // Skill ID（request body）
	Paths         []string `json:"paths"`
}

type GitResetReq struct {
	CustomSkillID string   `json:"customSkillId" validate:"required"` // Skill ID（request body）
	Paths         []string `json:"paths"`
}

type GitRestoreReq struct {
	CustomSkillID string `json:"customSkillId" validate:"required"` // Skill ID（request body）
	Commit        string `json:"commit" validate:"required"`
}

type GitDiscardReq struct {
	CustomSkillID string   `json:"customSkillId" validate:"required"`
	Paths         []string `json:"paths"`
}

type GitCommitReq struct {
	CustomSkillID string `json:"customSkillId" validate:"required"` // Skill ID（request body）
	Message       string `json:"message"`
}

// Check 校验 Git 状态请求参数。
func (r *GitStatusReq) Check() error {
	if r.FilePath == "" {
		return nil
	}
	return checkRelPath(r.FilePath)
}

// Check 校验 Git 暂存请求路径。
func (r *GitAddReq) Check() error {
	for _, p := range r.Paths {
		if p == "" {
			return fmt.Errorf("path is required")
		}
		if err := checkRelPath(p); err != nil {
			return fmt.Errorf("path %q: %w", p, err)
		}
	}
	return nil
}

// Check 校验 Git 取消暂存请求路径。
func (r *GitResetReq) Check() error {
	for _, p := range r.Paths {
		if p == "" {
			return fmt.Errorf("path is required")
		}
		if err := checkRelPath(p); err != nil {
			return fmt.Errorf("path %q: %w", p, err)
		}
	}
	return nil
}

// Check 校验 Git 恢复请求。
func (r *GitRestoreReq) Check() error {
	if r.Commit == "" {
		return fmt.Errorf("commit is required")
	}
	if err := gitutil.ValidateCommitRef(r.Commit); err != nil {
		return fmt.Errorf("invalid commit: %v", err)
	}
	return nil
}

// Check 校验 Git 放弃工作区更改请求路径。
func (r *GitDiscardReq) Check() error {
	for _, p := range r.Paths {
		if p == "" {
			return fmt.Errorf("path is required")
		}
		if err := checkRelPath(p); err != nil {
			return fmt.Errorf("path %q: %w", p, err)
		}
	}
	return nil
}

// Check 校验 Git 提交信息。
func (r *GitCommitReq) Check() error {
	if r.Message == "" {
		return fmt.Errorf("commit message is required")
	}
	if len(r.Message) > 5000 {
		return fmt.Errorf("commit message too long (max 5000 characters)")
	}
	// 拒绝 NUL 字节和其他 ASCII 控制字符（保留 \t \n \r）
	for _, ch := range r.Message {
		if ch == '\x00' || (ch < 0x20 && ch != '\t' && ch != '\n' && ch != '\r') {
			return fmt.Errorf("commit message contains invalid control character")
		}
		if unicode.Is(unicode.Cc, ch) && ch > 0x7e {
			return fmt.Errorf("commit message contains invalid control character")
		}
	}
	firstLine := strings.SplitN(r.Message, "\n", 2)[0]
	if len(firstLine) > 200 {
		return fmt.Errorf("commit message first line too long (max 200 characters)")
	}
	if strings.Count(r.Message, "\n") > 20 {
		return fmt.Errorf("commit message too many lines (max 20)")
	}
	return nil
}
