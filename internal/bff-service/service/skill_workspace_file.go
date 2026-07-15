package service

import (
	"fmt"
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
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// resolveDiskPath 把前端传入的 UTF-8 相对路径映射到磁盘上的真实文件路径。
//
// 写入端修复后新导入的 skill 文件名已是 UTF-8，但磁盘上仍可能存在历史 GBK 文件名
// （修复前导入的存量 skill）。前端从 files 接口拿到的是经 DecodeGBKToUTF8 转码后的
// UTF-8 path，回传时若直接拼磁盘路径，对存量 GBK 文件会找不到。
//
// 策略：先按 UTF-8 原样 Lstat（命中新文件）；失败则把 path 反向 GB18030 编码后再 Lstat
// （命中存量 GBK 文件）；都不存在则返回 UTF-8 原路径，让调用方按原逻辑报 not_found。
// 反向编码后的路径必须仍在 basePath 内，复用 path_util 的越界判定逻辑防止逃逸。
func resolveDiskPath(basePath, utf8RelPath string) (string, error) {
	fullPath, cleanRel, err := path_util.JoinWithinBase(basePath, utf8RelPath, false)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(fullPath); err == nil {
		return fullPath, nil // UTF-8 路径直接命中
	}
	// 存量 GBK 兜底：把 UTF-8 path 反向编码为 GB18030
	gbkRel := util.EncodeUTF8ToGBK(cleanRel)
	if gbkRel == cleanRel {
		return fullPath, nil // 编码无变化（ASCII 或含不可表示字符），放弃兜底
	}
	gbkFullPath := filepath.Join(basePath, filepath.FromSlash(gbkRel))
	// 安全校验：反向编码后的路径须仍在 basePath 内
	absGbk, err := filepath.Abs(gbkFullPath)
	if err != nil {
		return fullPath, nil
	}
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return fullPath, nil
	}
	rel, err := filepath.Rel(absBase, absGbk)
	if err != nil {
		return fullPath, nil
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") || filepath.IsAbs(rel) {
		return fullPath, nil // 越界，放弃兜底
	}
	if _, err := os.Lstat(absGbk); err == nil {
		return absGbk, nil // GBK 路径命中
	}
	return fullPath, nil // 都没命中，返回 UTF-8 原路径（调用方按原逻辑报 not_found）
}

// buildFileTree 递归构建文件树，跳过隐藏文件。
func buildFileTree(basePath, currentPath string, depth int, count *int) ([]*response.FileNode, error) {
	if depth > maxFileTreeDepth || *count >= maxFileTreeNodes {
		return []*response.FileNode{}, nil
	}
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	nodes := make([]*response.FileNode, 0)
	for _, entry := range entries {
		if *count >= maxFileTreeNodes {
			break
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		fullPath := filepath.Join(currentPath, entry.Name())
		relPath, err := filepath.Rel(basePath, fullPath)
		if err != nil {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// 磁盘上可能存在历史 GBK 文件名（写入端修复前导入的存量 skill）。
		// entry.Name() 把磁盘字节当 UTF-8 解码，GBK 非法序列会变成不可逆的 U+FFFD。
		// 这里对返回前端的 name/path 做兜底解码：非合法 UTF-8 则按 GBK/GB18030 解。
		// 注意 fullPath 始终用原始 entry.Name()（磁盘字节），保证能访问磁盘文件。
		name := util.DecodeGBKToUTF8(entry.Name())
		relPathUTF8 := util.DecodeGBKToUTF8(filepath.ToSlash(relPath))
		node := &response.FileNode{
			Name:    name,
			Path:    relPathUTF8,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().UnixMilli(),
		}
		*count = *count + 1
		if entry.IsDir() {
			children, err := buildFileTree(basePath, fullPath, depth+1, count)
			if err == nil {
				node.Children = children
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetSkillWorkspaceFiles 获取工作区文件树。
func GetSkillWorkspaceFiles(ctx *gin.Context, userId, orgId string, req request.GetSkillWorkspaceFilesReq) (*response.SkillWorkspaceFilesResp, error) {
	log.Infof("[Workspace] GetSkillWorkspaceFiles customSkillID: %v", req.CustomSkillID)

	ws, err := resolveSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, ws.workspaceDir, true); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if info, err := os.Stat(ws.workspaceDir); err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("[Workspace] stat workspace %s error: %v", ws.workspaceDir, err)
		}
		return &response.SkillWorkspaceFilesResp{Files: []*response.FileNode{}}, nil
	} else if !info.IsDir() {
		return &response.SkillWorkspaceFilesResp{Files: []*response.FileNode{}}, nil
	}

	count := 0
	files, err := buildFileTree(ws.workspaceDir, ws.workspaceDir, 0, &count)
	if err != nil {
		log.Errorf("[Workspace] buildFileTree error: %v", err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_build_file_tree_failed")
	}
	return &response.SkillWorkspaceFilesResp{Files: files}, nil
}

// GetSkillWorkspaceFile 读取工作区文件内容。
func GetSkillWorkspaceFile(ctx *gin.Context, userId, orgId string, req request.GetSkillWorkspaceFileReq) (*response.SkillWorkspaceFileResp, error) {
	ws, err := resolveSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	fullPath, err := resolveDiskPath(ws.workspaceDir, req.Path)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}

	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, fullPath, true); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	info, err := os.Lstat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_not_found")
		}
		log.Errorf("[Workspace] GetFile stat %s err: %v", fullPath, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_stat_file_failed")
	}
	if info.IsDir() {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_path_is_directory")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "symlink path not allowed")
	}
	if info.Size() > maxFileSize {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_too_large")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		log.Errorf("[Workspace] GetFile read %s err: %v", fullPath, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_read_file_failed")
	}
	return &response.SkillWorkspaceFileResp{
		Content: string(content),
		Size:    info.Size(),
		ModTime: info.ModTime().UnixMilli(),
	}, nil
}

// DownloadSkillWorkspace 下载工作区文件或目录。
func DownloadSkillWorkspace(ctx *gin.Context, userId, orgId string, req request.DownloadSkillWorkspaceReq) (string, []byte, error) {
	ws, err := resolveSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return "", nil, err
	}

	fullPath, err := resolveDiskPath(ws.workspaceDir, req.Path)
	if err != nil {
		return "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, fullPath, true); err != nil {
		return "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	// 下载文件名用前端 UTF-8 path 派生，避免存量 GBK 文件名导致的下载名乱码
	downloadName := filepath.Base(filepath.ToSlash(req.Path))

	info, err := os.Lstat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_not_found")
		}
		log.Errorf("[Workspace] Download stat %s err: %v", fullPath, err)
		return "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_stat_file_failed")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "symlink path not allowed")
	}
	if info.IsDir() {
		data, err := util.ZipDir(filepath.Join(fullPath, "."))
		if err != nil {
			log.Errorf("[Workspace] Download zip %s err: %v", fullPath, err)
			return "", nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("failed to create zip: %v", err))
		}
		return fmt.Sprintf("workspace_%s_%s.zip", req.CustomSkillID, downloadName), data, nil
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Errorf("[Workspace] Download read %s err: %v", fullPath, err)
		return "", nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_read_file_failed")
	}
	return downloadName, data, nil
}

// DeleteSkillWorkspaceFile 删除工作区文件或目录。
func DeleteSkillWorkspaceFile(ctx *gin.Context, userId, orgId string, req request.DeleteSkillWorkspaceFileReq) error {
	log.Infof("[Workspace] DeleteFile user=%s org=%s skill=%s path=%s", userId, orgId, req.CustomSkillID, req.Path)
	ws, err := resolveSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return err
	}

	fullPath, err := resolveDiskPath(ws.workspaceDir, req.Path)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, fullPath, true); err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_file_not_found")
		}
		log.Errorf("[Workspace] DeleteFile stat %s err: %v", fullPath, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_stat_file_failed")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, "symlink path not allowed")
	}

	if info.IsDir() {
		if err := os.RemoveAll(fullPath); err != nil {
			log.Errorf("[Workspace] DeleteFile remove dir %s err: %v", fullPath, err)
			return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_delete_file_failed")
		}
		return nil
	}
	if err := os.Remove(fullPath); err != nil {
		log.Errorf("[Workspace] DeleteFile remove file %s err: %v", fullPath, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_delete_file_failed")
	}
	return nil
}

// UpdateSkillWorkspaceFile 更新工作区文件内容。
func UpdateSkillWorkspaceFile(ctx *gin.Context, userId, orgId string, req request.UpdateSkillWorkspaceFileReq) (*response.UpdateSkillWorkspaceFileResp, error) {
	log.Infof("[Workspace] UpdateFile user=%s org=%s skill=%s path=%s size=%d", userId, orgId, req.CustomSkillID, req.Path, len(req.Content))
	ws, err := resolveSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}

	fullPath, err := resolveDiskPath(ws.workspaceDir, req.Path)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}

	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, ws.workspaceDir, true); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if err := os.MkdirAll(ws.workspaceDir, 0755); err != nil {
		log.Errorf("[Workspace] UpdateFile mkdir workspace %s err: %v", ws.workspaceDir, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_create_dir_failed")
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, filepath.Dir(fullPath), true); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Errorf("[Workspace] UpdateFile mkdir %s err: %v", filepath.Dir(fullPath), err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_create_dir_failed")
	}
	if info, err := os.Lstat(fullPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "symlink path not allowed")
		}
		if info.IsDir() {
			return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_path_is_directory")
		}
	}
	if err := util.WriteFileAtomic(fullPath, []byte(req.Content)); err != nil {
		log.Errorf("[Workspace] UpdateFile write %s err: %v", fullPath, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_write_file_failed")
	}

	return &response.UpdateSkillWorkspaceFileResp{}, nil
}
