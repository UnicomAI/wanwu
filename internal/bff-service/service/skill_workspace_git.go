package service

import (
	"os"
	"path"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	git_util "github.com/UnicomAI/wanwu/pkg/git-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
	"github.com/gin-gonic/gin"
)

// GetSkillWorkspaceGitLog 获取 git commit 历史。
func GetSkillWorkspaceGitLog(ctx *gin.Context, userId, orgId string, req request.GetSkillWorkspaceGitLogReq) (*response.SkillWorkspaceGitLogResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	count := req.Count
	if count <= 0 {
		count = 50
	}

	commits, err := ws.repo.GetCommitLog(count)
	if err != nil {
		// 空仓库时 git log 会返回错误，返回空列表而非报错
		log.Infof("[Workspace] GetGitLog skill=%s: %v (may be empty repo)", req.CustomSkillID, err)
		return &response.SkillWorkspaceGitLogResp{Commits: []response.GitCommitInfo{}}, nil
	}

	respCommits := make([]response.GitCommitInfo, len(commits))
	for i, c := range commits {
		respCommits[i] = response.GitCommitInfo{
			Hash:    c.Hash,
			Message: c.Message,
			Time:    c.Time,
			Tags:    c.Tags,
		}
	}
	return &response.SkillWorkspaceGitLogResp{Commits: respCommits}, nil
}

// GetSkillWorkspaceGitDiff 获取两个 commit 之间的 diff。
func GetSkillWorkspaceGitDiff(ctx *gin.Context, userId, orgId string, req request.GetSkillWorkspaceGitDiffReq) (*response.SkillWorkspaceGitDiffResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	fromCommit := req.FromCommit
	toCommit := req.ToCommit
	if toCommit == "" {
		toCommit = "HEAD"
	}
	defaultedFromCommit := fromCommit == ""
	if fromCommit == "" {
		fromCommit = "HEAD~1"
	}

	useRoot := false
	responseFromCommit := fromCommit
	diff, err := ws.repo.GetDiff(fromCommit, toCommit, generalAgentWorkspaceSkillDirName)
	if err != nil {
		if !shouldFallbackGitDiffToRoot(defaultedFromCommit, fromCommit, toCommit) {
			log.Errorf("[Workspace] GetGitDiff skill=%s fromCommit=%s toCommit=%s err=%v", req.CustomSkillID, fromCommit, toCommit, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_diff_failed")
		}
		log.Infof("[Workspace] GetGitDiff fromCommit=%s failed, trying --root: %v", fromCommit, err)
		useRoot = true
		responseFromCommit = "" // 空字符串表示从初始提交开始，git-util 中空 fromCommit 触发 git show --root
		diff, err = ws.repo.GetDiff("", toCommit, generalAgentWorkspaceSkillDirName)
		if err != nil {
			log.Errorf("[Workspace] GetGitDiff skill=%s err=%v", req.CustomSkillID, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_diff_failed")
		}
	}

	var changedFileInfos []git_util.FileChangeInfo
	if useRoot {
		changedFileInfos, err = ws.repo.GetChangedFiles("", toCommit, generalAgentWorkspaceSkillDirName)
	} else {
		changedFileInfos, err = ws.repo.GetChangedFiles(fromCommit, toCommit, generalAgentWorkspaceSkillDirName)
	}
	if err != nil {
		log.Warnf("[Workspace] GetGitDiff get changed files err: %v", err)
		changedFileInfos = []git_util.FileChangeInfo{}
	}

	return &response.SkillWorkspaceGitDiffResp{
		FromCommit:   responseFromCommit,
		ToCommit:     toCommit,
		Diff:         diff,
		ChangedFiles: toGitFileChanges(changedFileInfos),
	}, nil
}

// toGitFileChanges 将 git-util 文件变更信息转换为接口响应结构。
func toGitFileChanges(files []git_util.FileChangeInfo) []response.GitFileChange {
	changes := make([]response.GitFileChange, len(files))
	for i, f := range files {
		changes[i] = response.GitFileChange{
			Path:       f.Path,
			OldPath:    f.OldPath,
			ChangeType: f.ChangeType,
		}
	}
	return changes
}

// GetSkillWorkspaceGitFile 获取指定 commit 中某文件的内容。
func GetSkillWorkspaceGitFile(ctx *gin.Context, userId, orgId string, req request.GetSkillWorkspaceGitFileReq) (*response.SkillWorkspaceGitFileResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	commitHash := req.CommitHash
	if commitHash == "" {
		commitHash, err = ws.repo.GetHeadCommit()
		if err != nil {
			log.Errorf("[Workspace] GetGitFile GetHeadCommit skill=%s err=%v", req.CustomSkillID, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_head_commit_failed")
		}
	}

	_, cleanRelPath, err := workspaceFilePath(ws.workspaceDir, req.FilePath)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, err.Error())
	}
	relGitFilePath := path.Join(generalAgentWorkspaceSkillDirName, cleanRelPath)

	content, err := ws.repo.GetFileContentAtCommit(commitHash, relGitFilePath)
	if err != nil {
		log.Errorf("[Workspace] GetGitFile GetFileContentAtCommit skill=%s commit=%s path=%s err=%v", req.CustomSkillID, commitHash, relGitFilePath, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_file_content_failed")
	}
	return &response.SkillWorkspaceGitFileResp{
		FilePath:   req.FilePath,
		Content:    content,
		CommitHash: commitHash,
	}, nil
}

// GetSkillWorkspaceGitFileDiff 获取单个文件在两个 commit 之间的 diff。
func GetSkillWorkspaceGitFileDiff(ctx *gin.Context, userId, orgId string, req request.GetSkillWorkspaceGitFileDiffReq) (*response.SkillWorkspaceGitDiffResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	fromCommit := req.FromCommit
	toCommit := req.ToCommit
	if toCommit == "" {
		toCommit = "HEAD"
	}
	defaultedFromCommit := fromCommit == ""
	if fromCommit == "" {
		fromCommit = "HEAD~1"
	}

	_, cleanRelPath, err := workspaceFilePath(ws.workspaceDir, req.FilePath)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, err.Error())
	}
	relGitFilePath := path.Join(generalAgentWorkspaceSkillDirName, cleanRelPath)

	responseFromCommit := fromCommit
	diff, err := ws.repo.GetFileDiff(fromCommit, toCommit, relGitFilePath)
	if err != nil {
		if !shouldFallbackGitDiffToRoot(defaultedFromCommit, fromCommit, toCommit) {
			log.Errorf("[Workspace] GetGitFileDiff skill=%s fromCommit=%s toCommit=%s path=%s err=%v", req.CustomSkillID, fromCommit, toCommit, relGitFilePath, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_file_diff_failed")
		}
		log.Infof("[Workspace] GetGitFileDiff fromCommit=%s failed, trying --root: %v", fromCommit, err)
		responseFromCommit = "" // 空字符串表示从初始提交开始，git-util 中空 fromCommit 触发 git show --root
		diff, err = ws.repo.GetFileDiff("", toCommit, relGitFilePath)
		if err != nil {
			log.Errorf("[Workspace] GetGitFileDiff skill=%s err=%v", req.CustomSkillID, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_get_file_diff_failed")
		}
	}
	return &response.SkillWorkspaceGitDiffResp{
		FromCommit:   responseFromCommit,
		ToCommit:     toCommit,
		Diff:         diff,
		ChangedFiles: []response.GitFileChange{{Path: req.FilePath, ChangeType: "modified"}},
	}, nil
}

// shouldFallbackGitDiffToRoot 判断默认提交范围是否需要回退到根提交。
func shouldFallbackGitDiffToRoot(defaultedFromCommit bool, fromCommit, toCommit string) bool {
	if defaultedFromCommit {
		return true
	}
	if fromCommit == "" || toCommit == "" {
		return false
	}
	for _, suffix := range []string{"~1", "~", "^1", "^"} {
		if strings.HasSuffix(fromCommit, suffix) {
			return strings.TrimSuffix(fromCommit, suffix) == toCommit
		}
	}
	return false
}

// GetGitStatus 获取工作区 git 状态。
func GetGitStatus(ctx *gin.Context, userId, orgId string, req request.GitStatusReq) (*response.GitStatusResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	statusFiles, err := ws.repo.Status(generalAgentWorkspaceSkillDirName)
	if err != nil {
		log.Errorf("[Workspace] GetGitStatus GitStatus skill=%s err=%v", req.CustomSkillID, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_status_failed")
	}

	respFiles := make([]response.GitStatusFile, len(statusFiles))
	for i, f := range statusFiles {
		respFiles[i] = response.GitStatusFile{
			Path:       f.Path,
			OldPath:    f.OldPath,
			ChangeType: f.ChangeType,
			Staged:     f.Staged,
		}
	}
	return &response.GitStatusResp{Files: respFiles}, nil
}

// GitAdd 暂存文件。
func GitAdd(ctx *gin.Context, userId, orgId string, req request.GitAddReq) error {
	log.Infof("[Workspace] GitAdd user=%s org=%s skill=%s paths=%v", userId, orgId, req.CustomSkillID, req.Paths)
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return err
	}

	if err := ws.repo.Add(req.Paths, generalAgentWorkspaceSkillDirName); err != nil {
		log.Errorf("[Workspace] GitAdd skill=%s paths=%v err=%v", req.CustomSkillID, req.Paths, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_add_failed")
	}
	return nil
}

// GitReset 取消暂存文件。
func GitReset(ctx *gin.Context, userId, orgId string, req request.GitResetReq) error {
	log.Infof("[Workspace] GitReset user=%s org=%s skill=%s paths=%v", userId, orgId, req.CustomSkillID, req.Paths)
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return err
	}

	if err := ws.repo.Reset(req.Paths, generalAgentWorkspaceSkillDirName); err != nil {
		log.Errorf("[Workspace] GitReset skill=%s paths=%v err=%v", req.CustomSkillID, req.Paths, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_reset_failed")
	}
	return nil
}

// GitRestore 恢复整个 Skill 工作区到指定 commit。
func GitRestore(ctx *gin.Context, userId, orgId string, req request.GitRestoreReq) error {
	log.Infof("[Workspace] GitRestore user=%s org=%s skill=%s commit=%s", userId, orgId, req.CustomSkillID, req.Commit)
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return err
	}

	if err := ws.repo.Restore(req.Commit, generalAgentWorkspaceSkillDirName); err != nil {
		log.Errorf("[Workspace] GitRestore skill=%s commit=%s err=%v", req.CustomSkillID, req.Commit, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_restore_failed")
	}
	return nil
}

// GitDiscardWorkingTree 放弃未暂存的工作区变更。
func GitDiscardWorkingTree(ctx *gin.Context, userId, orgId string, req request.GitDiscardReq) error {
	log.Infof("[Workspace] GitDiscard user=%s org=%s skill=%s paths=%v", userId, orgId, req.CustomSkillID, req.Paths)
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return err
	}

	if err := ws.repo.DiscardWorkingTree(req.Paths, generalAgentWorkspaceSkillDirName); err != nil {
		log.Errorf("[Workspace] GitDiscard skill=%s paths=%v err=%v", req.CustomSkillID, req.Paths, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_discard_failed")
	}
	return nil
}

// GitCommitAction 提交已暂存的变更。
func GitCommitAction(ctx *gin.Context, userId, orgId string, req request.GitCommitReq) (string, error) {
	log.Infof("[Workspace] GitCommit user=%s org=%s skill=%s msgLen=%d", userId, orgId, req.CustomSkillID, len(req.Message))
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return "", err
	}

	commitHash, err := ws.repo.Commit(req.Message)
	if err != nil {
		log.Errorf("[Workspace] GitCommit skill=%s err=%v", req.CustomSkillID, err)
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_commit_failed")
	}
	return commitHash, nil
}

// GetGitDiffWorkingTree 获取未暂存的 diff（工作目录 vs 暂存区）。
func GetGitDiffWorkingTree(ctx *gin.Context, userId, orgId string, req request.GitStatusReq) (*response.SkillWorkspaceGitDiffResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	diff, err := ws.repo.DiffWorkingTree(generalAgentWorkspaceSkillDirName, req.FilePath)
	if err != nil {
		log.Errorf("[Workspace] GetGitDiffWorkingTree skill=%s err=%v", req.CustomSkillID, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_diff_failed")
	}
	resp := &response.SkillWorkspaceGitDiffResp{Diff: diff}
	if req.FilePath != "" {
		oldContent, newContent, err := getGitDiffWorkingTreeContents(ws, req.FilePath)
		if err != nil {
			log.Errorf("[Workspace] GetGitDiffWorkingTree content skill=%s path=%s err=%v", req.CustomSkillID, req.FilePath, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_diff_content_failed")
		}
		resp.OldContent = &oldContent
		resp.NewContent = &newContent
	}
	return resp, nil
}

// GetGitDiffStaged 获取已暂存的 diff（暂存区 vs HEAD）。
func GetGitDiffStaged(ctx *gin.Context, userId, orgId string, req request.GitStatusReq) (*response.SkillWorkspaceGitDiffResp, error) {
	ws, err := resolveInitializedSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	diff, err := ws.repo.DiffStaged(generalAgentWorkspaceSkillDirName, req.FilePath)
	if err != nil {
		log.Errorf("[Workspace] GetGitDiffStaged skill=%s err=%v", req.CustomSkillID, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_diff_staged_failed")
	}
	resp := &response.SkillWorkspaceGitDiffResp{Diff: diff}
	if req.FilePath != "" {
		oldContent, newContent, err := getGitDiffStagedContents(ws, req.FilePath)
		if err != nil {
			log.Errorf("[Workspace] GetGitDiffStaged content skill=%s path=%s err=%v", req.CustomSkillID, req.FilePath, err)
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_git_diff_content_failed")
		}
		resp.OldContent = &oldContent
		resp.NewContent = &newContent
	}
	return resp, nil
}

// getGitDiffStagedContents 读取单文件 HEAD 与暂存区内容。
func getGitDiffStagedContents(ws *skillWorkspaceContext, filePath string) (string, string, error) {
	_, cleanRelPath, err := workspaceFilePath(ws.workspaceDir, filePath)
	if err != nil {
		return "", "", err
	}
	relGitFilePath := path.Join(generalAgentWorkspaceSkillDirName, cleanRelPath)

	oldSnapshot, err := ws.repo.GetFileSnapshotAtCommit("HEAD", relGitFilePath)
	if err != nil {
		return "", "", err
	}
	newSnapshot, err := ws.repo.GetFileSnapshotAtIndex(relGitFilePath)
	if err != nil {
		return "", "", err
	}
	return oldSnapshot.Content, newSnapshot.Content, nil
}

// getGitDiffWorkingTreeContents 读取单文件暂存区与工作区内容。
func getGitDiffWorkingTreeContents(ws *skillWorkspaceContext, filePath string) (string, string, error) {
	_, cleanRelPath, err := workspaceFilePath(ws.workspaceDir, filePath)
	if err != nil {
		return "", "", err
	}
	relGitFilePath := path.Join(generalAgentWorkspaceSkillDirName, cleanRelPath)

	oldSnapshot, err := ws.repo.GetFileSnapshotAtIndex(relGitFilePath)
	if err != nil {
		return "", "", err
	}
	newContent, err := readWorkspaceFileContentOrEmpty(ws, filePath)
	if err != nil {
		return "", "", err
	}
	return oldSnapshot.Content, newContent, nil
}

// readWorkspaceFileContentOrEmpty 读取工作区文件，不存在时返回空内容。
func readWorkspaceFileContentOrEmpty(ws *skillWorkspaceContext, filePath string) (string, error) {
	fullPath, _, err := workspaceFilePath(ws.workspaceDir, filePath)
	if err != nil {
		return "", err
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, fullPath, true); err != nil {
		return "", err
	}
	info, err := os.Lstat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if info.IsDir() {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_path_is_directory")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, "symlink path not allowed")
	}
	if info.Size() > maxFileSize {
		return "", grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_too_large")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
