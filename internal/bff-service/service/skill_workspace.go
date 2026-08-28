package service

import (
	"fmt"
	"path/filepath"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	git_util "github.com/UnicomAI/wanwu/pkg/git-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
)

const (
	generalAgentWorkspaceSkillDirName = "skill"         // wga 工作区中 skill 目录名
	maxFileSize                       = 1 * 1024 * 1024 // 最大文件大小 1MB
	maxSearchResults                  = 1000            // 最大搜索结果数
	maxFileTreeNodes                  = 5000
	maxFileTreeDepth                  = 100
	maxSearchLineBytes                = 256 * 1024
	maxSearchResultContentLength      = 4096
)

type skillWorkspaceContext struct {
	customSkillID string
	skillDir      string
	workspaceDir  string
	repo          git_util.Repo
}

// resolveSkillWorkspace 校验归属并解析 Skill 工作区上下文。
func resolveSkillWorkspace(customSkillID string) (*skillWorkspaceContext, error) {
	skillDir, err := getSkillDir(customSkillID)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_dir_failed")
	}
	if skillDir == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_not_found")
	}
	return &skillWorkspaceContext{
		customSkillID: customSkillID,
		skillDir:      skillDir,
		workspaceDir:  filepath.Join(skillDir, generalAgentWorkspaceSkillDirName),
		repo:          git_util.Open(skillDir),
	}, nil
}

// resolveInitializedSkillWorkspace 解析并确保工作区已初始化 Git 仓库。
func resolveInitializedSkillWorkspace(customSkillID string) (*skillWorkspaceContext, error) {
	ws, err := resolveSkillWorkspace(customSkillID)
	if err != nil {
		return nil, err
	}
	if _, err := ensureGitInitializedAt(customSkillID, ws.skillDir); err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_init_git_failed")
	}
	if !ws.repo.IsInitialized() {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_not_initialized")
	}
	return ws, nil
}

// getSkillDir 获取 skill 目录路径（包含 .git 和 skill/）。
func getSkillDir(customSkillID string) (string, error) {
	store, err := NewGeneralAgentSkillWorkspaceStore(customSkillID)
	if err != nil {
		return "", err
	}
	info := store.GetThreadDir()
	if info.Dir == "" {
		return "", nil
	}
	return info.Dir, nil
}

// ensureGitInitializedAt 在指定 skill 目录上按需初始化 Git 仓库。
func ensureGitInitializedAt(customSkillID, skillDir string) (bool, error) {
	if skillDir == "" {
		return false, nil
	}

	repo := git_util.Open(skillDir)

	// 快速路径：已初始化无需加锁
	if repo.IsInitialized() {
		return false, nil
	}

	// 慢速路径：持仓库锁，防止多请求并发初始化
	mu := repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()

	// 再次检查（另一个 goroutine 可能刚刚完成初始化）
	if repo.IsInitialized() {
		return false, nil
	}

	log.Infof("[Workspace] initializing git repo for skill %s", customSkillID)

	if err := repo.InitLocked(); err != nil {
		return false, fmt.Errorf("git init failed: %w", err)
	}

	return true, nil
}

// workspaceFilePath 解析工作区内文件路径并返回绝对路径与清理后的相对路径。
func workspaceFilePath(basePath, relativePath string) (string, string, error) {
	fullPath, cleanPath, err := path_util.JoinWithinBase(basePath, relativePath, false)
	if err != nil {
		return "", "", err
	}
	if isWorkspaceMetadataPath(cleanPath) {
		return "", "", fmt.Errorf("workspace metadata path is not allowed")
	}
	return fullPath, cleanPath, nil
}

// commitSkillWorkspaceIfFirstCommit 在 skill git 工作区"从未提交(无 HEAD)"时，
// 将 skill 子目录全部变更提交。用于 import/convert/chat 首轮产物落盘后的首次提交。
// 仅当从未提交时才提交；已存在 HEAD 则跳过，避免覆盖用户后续手动提交历史。
// best-effort：任何失败仅记日志，不阻断主流程、不回滚已落盘文件。
func commitSkillWorkspaceIfFirstCommit(customSkillID, message string) {
	ws, err := resolveInitializedSkillWorkspace(customSkillID)
	if err != nil {
		log.Warnf("[skill-autocommit] customSkillID %v resolve workspace err: %v", customSkillID, err)
		return
	}
	// 仅"从未提交"才提交；已有 HEAD 直接跳过
	if ws.repo.HasHead() {
		return
	}
	hasChanges, err := ws.repo.HasChangesInSubDir(generalAgentWorkspaceSkillDirName)
	if err != nil {
		log.Warnf("[skill-autocommit] customSkillID %v check changes err: %v", customSkillID, err)
		return
	}
	if !hasChanges {
		return
	}
	mu := ws.repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()
	commitHash, err := ws.repo.CommitAllInSubDirLocked(generalAgentWorkspaceSkillDirName, message)
	if err != nil {
		log.Warnf("[skill-autocommit] customSkillID %v commit err: %v", customSkillID, err)
		return
	}
	log.Infof("[skill-autocommit] customSkillID %v first commit %s msg=%q", customSkillID, commitHash, message)
}
