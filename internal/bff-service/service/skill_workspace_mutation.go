package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
	"github.com/gin-gonic/gin"
)

const (
	workspaceUploadField     = "files"
	workspaceUploadBodyLimit = SkillWorkspaceUploadBodyLimit
)

func workspaceMutationError(key string) error {
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, key)
}

func resolveAuthorizedWorkspace(ctx *gin.Context, userID, orgID, skillID string) (*skillWorkspaceContext, error) {
	if err := authorizeSkillWorkspaceEdit(ctx, userID, orgID, skillID); err != nil {
		return nil, err
	}
	return resolveSkillWorkspace(skillID)
}

func ensureWorkspaceRootLocked(ws *skillWorkspaceContext) error {
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, ws.workspaceDir, true); err != nil {
		return err
	}
	if err := os.MkdirAll(ws.workspaceDir, 0755); err != nil {
		return err
	}
	return path_util.EnsureNoSymlinkInPath(ws.workspaceDir, ws.workspaceDir, true)
}

func ensureMutationTargetPath(ws *skillWorkspaceContext, rel string) (string, string, error) {
	full, clean, err := path_util.JoinWithinBase(ws.workspaceDir, rel, false)
	if err != nil {
		return "", "", err
	}
	if clean == "." || isWorkspaceMetadataPath(clean) {
		return "", "", fmt.Errorf("workspace metadata path is not allowed")
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, full, true); err != nil {
		return "", "", err
	}
	return full, clean, nil
}

// ensureMutationDirectoryPath 解析目标目录。与条目路径不同，工作区根目录
// 也可以作为上传目标。
func ensureMutationDirectoryPath(ws *skillWorkspaceContext, rel string) (string, string, error) {
	full, clean, err := path_util.JoinWithinBase(ws.workspaceDir, rel, true)
	if err != nil {
		return "", "", err
	}
	if clean == "." {
		clean = ""
	}
	if isWorkspaceMetadataPath(clean) {
		return "", "", fmt.Errorf("workspace metadata path is not allowed")
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, full, true); err != nil {
		return "", "", err
	}
	return full, clean, nil
}

func regularWorkspaceInfo(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || info.Mode()&os.ModeNamedPipe != 0 || info.Mode()&os.ModeDevice != 0 || info.Mode()&os.ModeSocket != 0 {
		return nil, fmt.Errorf("special workspace entry is not allowed")
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("non-regular workspace entry is not allowed")
	}
	return info, nil
}

func workspaceEntryNode(root, fullPath, relPath string, info os.FileInfo) *response.FileNode {
	name := filepath.Base(filepath.ToSlash(relPath))
	node := &response.FileNode{
		Name:    name,
		Path:    filepath.ToSlash(relPath),
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime().UnixMilli(),
	}
	if info.IsDir() {
		node.Children = []*response.FileNode{}
	}
	return node
}

// CreateSkillWorkspaceFile 创建新文件，不覆盖已有条目。仓库互斥锁会串行化
// 当前工作区的路径校验和文件系统变更。
func CreateSkillWorkspaceFile(ctx *gin.Context, userID, orgID string, req request.CreateSkillWorkspaceFileReq) (*response.CreateSkillWorkspaceEntryResp, error) {
	if len(req.Content) > maxFileSize {
		return nil, workspaceMutationError("bff_skill_workspace_upload_too_large")
	}
	ws, err := resolveAuthorizedWorkspace(ctx, userID, orgID, req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	full, clean, err := ensureMutationTargetPath(ws, req.Path)
	if err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	if err := validateWorkspaceName(filepath.Base(clean)); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_name_invalid")
	}
	mu := ws.repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()
	if err := ensureWorkspaceRootLocked(ws); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_create_dir_failed")
	}
	parent := filepath.Dir(full)
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, parent, true); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	parentInfo, err := regularWorkspaceInfo(parent)
	if err != nil || !parentInfo.IsDir() {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_found")
	}
	if _, err := os.Lstat(full); err == nil {
		return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
	} else if !os.IsNotExist(err) {
		return nil, workspaceMutationError("bff_skill_workspace_stat_file_failed")
	}
	f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
		}
		return nil, workspaceMutationError("bff_skill_workspace_write_file_failed")
	}
	if _, err = f.WriteString(req.Content); err != nil {
		_ = f.Close()
		_ = os.Remove(full)
		return nil, workspaceMutationError("bff_skill_workspace_write_file_failed")
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(full)
		return nil, workspaceMutationError("bff_skill_workspace_write_file_failed")
	}
	info, err := regularWorkspaceInfo(full)
	if err != nil {
		_ = os.Remove(full)
		return nil, workspaceMutationError("bff_skill_workspace_write_file_failed")
	}
	return &response.CreateSkillWorkspaceEntryResp{Entry: workspaceEntryNode(ws.workspaceDir, full, clean, info)}, nil
}

// CreateSkillWorkspaceDirectory 创建一个目录条目。父目录必须已存在；此操作
// 只创建一个新目录，不会擅自创建任意路径前缀。
func CreateSkillWorkspaceDirectory(ctx *gin.Context, userID, orgID string, req request.CreateSkillWorkspaceDirectoryReq) (*response.CreateSkillWorkspaceEntryResp, error) {
	ws, err := resolveAuthorizedWorkspace(ctx, userID, orgID, req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	full, clean, err := ensureMutationTargetPath(ws, req.Path)
	if err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	if err := validateWorkspaceName(filepath.Base(clean)); err != nil || isWorkspaceMetadataPath(clean) {
		return nil, workspaceMutationError("bff_skill_workspace_name_invalid")
	}
	mu := ws.repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()
	if err := ensureWorkspaceRootLocked(ws); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_create_dir_failed")
	}
	parent := filepath.Dir(full)
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, parent, true); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	parentInfo, err := regularWorkspaceInfo(parent)
	if err != nil || !parentInfo.IsDir() {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_found")
	}
	if err := os.Mkdir(full, 0755); err != nil {
		if os.IsExist(err) {
			return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
		}
		return nil, workspaceMutationError("bff_skill_workspace_create_dir_failed")
	}
	info, err := regularWorkspaceInfo(full)
	if err != nil {
		_ = os.Remove(full)
		return nil, workspaceMutationError("bff_skill_workspace_create_dir_failed")
	}
	return &response.CreateSkillWorkspaceEntryResp{Entry: workspaceEntryNode(ws.workspaceDir, full, clean, info)}, nil
}

// RenameSkillWorkspaceEntry 支持文件和目录。目标始终位于同级目录，
// 因此调用方无法将条目移动到其他工作区。
func RenameSkillWorkspaceEntry(ctx *gin.Context, userID, orgID string, req request.RenameSkillWorkspaceFileReq) (*response.RenameSkillWorkspaceEntryResp, error) {
	if err := validateWorkspaceName(req.NewName); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_name_invalid")
	}
	ws, err := resolveAuthorizedWorkspace(ctx, userID, orgID, req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	source, cleanSource, err := ensureMutationTargetPath(ws, req.Path)
	if err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	if isWorkspaceMetadataPath(cleanSource) {
		return nil, workspaceMutationError("bff_skill_workspace_name_invalid")
	}
	target := filepath.Join(filepath.Dir(source), req.NewName)
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, target, false); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	_, cleanTarget, err := path_util.JoinWithinBase(ws.workspaceDir, filepath.ToSlash(filepath.Join(filepath.Dir(cleanSource), req.NewName)), false)
	if err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	mu := ws.repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()
	if err := ensureWorkspaceRootLocked(ws); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	_, err = regularWorkspaceInfo(source)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, workspaceMutationError("bff_skill_workspace_file_not_found")
		}
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	if _, err := os.Lstat(target); err == nil {
		return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
	} else if !os.IsNotExist(err) {
		return nil, workspaceMutationError("bff_skill_workspace_stat_file_failed")
	}
	if err := os.Rename(source, target); err != nil {
		if os.IsExist(err) {
			return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
		}
		return nil, workspaceMutationError("bff_skill_workspace_operation_conflict")
	}
	newInfo, err := regularWorkspaceInfo(target)
	if err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_operation_conflict")
	}
	return &response.RenameSkillWorkspaceEntryResp{Entry: workspaceEntryNode(ws.workspaceDir, target, cleanTarget, newInfo)}, nil
}

func validateUploadHeader(header *multipart.FileHeader) error {
	if header == nil {
		return fmt.Errorf("empty upload")
	}
	if header.Filename == "" || filepath.Base(filepath.ToSlash(header.Filename)) != filepath.Base(header.Filename) {
		return fmt.Errorf("upload filename must be a basename")
	}
	if strings.HasSuffix(header.Filename, "/") || strings.HasSuffix(header.Filename, "\\") {
		return fmt.Errorf("directory upload is not allowed")
	}
	if err := validateWorkspaceName(header.Filename); err != nil {
		return err
	}
	if header.Size < 0 || header.Size > maxWorkspaceUploadFileBytes {
		return fmt.Errorf("upload file too large")
	}
	return nil
}

// UploadSkillWorkspaceFiles 将每个 multipart 部分流式写入同目录临时文件，
// 然后以原子方式重命名。写入前会校验所有名称和冲突，失败时删除部分输出。
func UploadSkillWorkspaceFiles(ctx *gin.Context, userID, orgID string, customSkillID, relDir string) (*response.SkillWorkspaceUploadResp, error) {
	if ctx.Request.ContentLength > workspaceUploadBodyLimit {
		return nil, workspaceMutationError("bff_skill_workspace_upload_too_large")
	}
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			return nil, workspaceMutationError("bff_skill_workspace_upload_too_large")
		}
		return nil, workspaceMutationError("bff_skill_workspace_invalid_multipart")
	}
	form := ctx.Request.MultipartForm
	if form == nil {
		return nil, workspaceMutationError("bff_skill_workspace_invalid_multipart")
	}
	files := form.File[workspaceUploadField]
	if len(files) == 0 {
		return nil, workspaceMutationError("bff_skill_workspace_upload_empty")
	}
	if len(files) > maxWorkspaceUploadFiles {
		return nil, workspaceMutationError("bff_skill_workspace_upload_too_large")
	}
	var total int64
	seen := make(map[string]struct{}, len(files))
	for _, header := range files {
		if err := validateUploadHeader(header); err != nil {
			return nil, workspaceMutationError("bff_skill_workspace_name_invalid")
		}
		name := filepath.Base(header.Filename)
		if _, ok := seen[name]; ok {
			return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
		}
		seen[name] = struct{}{}
		total += header.Size
		if total > maxWorkspaceUploadTotalBytes {
			return nil, workspaceMutationError("bff_skill_workspace_upload_too_large")
		}
	}

	ws, err := resolveAuthorizedWorkspace(ctx, userID, orgID, customSkillID)
	if err != nil {
		return nil, err
	}
	targetDir, cleanDir, err := ensureMutationDirectoryPath(ws, relDir)
	if err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	mu := ws.repo.GetMutex()
	mu.Lock()
	defer mu.Unlock()
	if err := ensureWorkspaceRootLocked(ws); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_create_dir_failed")
	}
	info, err := regularWorkspaceInfo(targetDir)
	if err != nil || !info.IsDir() {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_found")
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, targetDir, true); err != nil {
		return nil, workspaceMutationError("bff_skill_workspace_path_not_allowed")
	}
	for _, header := range files {
		dest := filepath.Join(targetDir, filepath.Base(header.Filename))
		if _, err := os.Lstat(dest); err == nil {
			return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
		} else if !os.IsNotExist(err) {
			return nil, workspaceMutationError("bff_skill_workspace_stat_file_failed")
		}
	}

	created := make([]string, 0, len(files))
	cleanup := func() {
		for _, path := range created {
			_ = os.Remove(path)
		}
	}
	result := make([]*response.FileNode, 0, len(files))
	for _, header := range files {
		dest := filepath.Join(targetDir, filepath.Base(header.Filename))
		tmp, err := os.CreateTemp(targetDir, ".upload-*")
		if err != nil {
			cleanup()
			return nil, workspaceMutationError("bff_skill_workspace_upload_failed")
		}
		tmpName := tmp.Name()
		ok := false
		func() {
			defer func() { _ = tmp.Close() }()
			file, openErr := header.Open()
			if openErr != nil {
				return
			}
			defer func() { _ = file.Close() }()
			written, copyErr := io.CopyN(tmp, file, header.Size)
			if copyErr != nil && (copyErr != io.EOF || written != header.Size) {
				return
			}
			// 不能允许声明大小更小的部分被静默截断；在限定长度复制后再读取一个字节探测。
			var probe [1]byte
			if n, probeErr := file.Read(probe[:]); n > 0 || (probeErr != io.EOF && probeErr != nil) {
				return
			}
			if syncErr := tmp.Sync(); syncErr != nil {
				return
			}
			ok = written == header.Size
		}()
		if !ok {
			_ = os.Remove(tmpName)
			cleanup()
			return nil, workspaceMutationError("bff_skill_workspace_upload_failed")
		}
		if err := os.Chmod(tmpName, 0644); err != nil {
			_ = os.Remove(tmpName)
			cleanup()
			return nil, workspaceMutationError("bff_skill_workspace_upload_failed")
		}
		if err := os.Rename(tmpName, dest); err != nil {
			_ = os.Remove(tmpName)
			cleanup()
			if os.IsExist(err) {
				return nil, workspaceMutationError("bff_skill_workspace_name_conflict")
			}
			return nil, workspaceMutationError("bff_skill_workspace_upload_failed")
		}
		created = append(created, dest)
		fileInfo, infoErr := regularWorkspaceInfo(dest)
		if infoErr != nil {
			_ = os.Remove(dest)
			cleanup()
			return nil, workspaceMutationError("bff_skill_workspace_upload_failed")
		}
		cleanPath := filepath.ToSlash(filepath.Join(cleanDir, filepath.Base(header.Filename)))
		result = append(result, workspaceEntryNode(ws.workspaceDir, dest, cleanPath, fileInfo))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	log.Infof("[Workspace] uploaded %d files skill=%s dir=%s", len(result), customSkillID, cleanDir)
	return &response.SkillWorkspaceUploadResp{Files: result}, nil
}
