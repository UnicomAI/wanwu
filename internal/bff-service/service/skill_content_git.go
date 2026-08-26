package service

import (
	"sort"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	git_util "github.com/UnicomAI/wanwu/pkg/git-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// 已发布内容全部从 git 对象库按 tag 读取：只读对象库、不 checkout、不碰工作树，
// 因此用户草稿的未提交更改不受影响。过滤与限额规则与草稿态（skill_workspace_file.go /
// skill_workspace_search.go）保持一致，保证两种来源下前端看到的行为相同。

// hasHiddenSegment 判断路径中是否存在以 . 开头的段，与草稿态 buildFileTree/shouldVisit 的过滤规则一致。
func hasHiddenSegment(relPath string) bool {
	for _, part := range strings.Split(relPath, "/") {
		if part != "." && strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// decodeTreeEntryPath 把 git 中存储的原始路径字节解码为 UTF-8。
// 存量仓库中可能存在 GBK 文件名，需转码后再返回前端，与草稿态 buildFileTree 的处理一致。
func decodeTreeEntryPath(rawPath string) string {
	return util.DecodeGBKToUTF8(rawPath)
}

// listSkillTreeEntries 列出指定版本 tag 下 skill 目录中的全部普通文件条目。
func listSkillTreeEntries(skillDir, version string) ([]git_util.TreeEntry, error) {
	entries, err := git_util.ListTreeFiles(skillDir, version, generalAgentWorkspaceSkillDirName)
	if err != nil {
		log.Errorf("[SkillContent] list tree files dir=%s version=%s err: %v", skillDir, version, err)
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_content_read_tree_failed")
	}
	return entries, nil
}

// buildTreeFromEntries 把扁平的 blob 列表折叠成文件树。
//
// git ls-tree -r 只返回 blob，中间目录需要从文件路径逐段推导。所有节点的 modTime
// 统一取发布提交的时间——git 不保存单文件 mtime。
func buildTreeFromEntries(entries []git_util.TreeEntry, commitTimeMs int64) []*response.FileNode {
	root := make([]*response.FileNode, 0)
	dirIndex := make(map[string]*response.FileNode) // 相对路径 -> 目录节点
	count := 0

	for _, entry := range entries {
		if !entry.IsRegularFile() { // 跳过符号链接与子模块，与草稿态跳过 symlink 一致
			continue
		}
		relPath := decodeTreeEntryPath(entry.Path)
		if relPath == "" || hasHiddenSegment(relPath) {
			continue
		}
		segments := strings.Split(relPath, "/")
		if len(segments) > maxFileTreeDepth {
			continue
		}

		// 逐段确保父目录存在，再挂载文件节点
		children := &root
		parentPath := ""
		truncated := false
		for _, segment := range segments[:len(segments)-1] {
			if parentPath == "" {
				parentPath = segment
			} else {
				parentPath += "/" + segment
			}
			dirNode, ok := dirIndex[parentPath]
			if !ok {
				if count >= maxFileTreeNodes {
					truncated = true
					break
				}
				dirNode = &response.FileNode{
					Name:     segment,
					Path:     parentPath,
					IsDir:    true,
					ModTime:  commitTimeMs,
					Children: make([]*response.FileNode, 0),
				}
				dirIndex[parentPath] = dirNode
				*children = append(*children, dirNode)
				count++
			}
			children = &dirNode.Children
		}
		if truncated {
			continue
		}
		if count >= maxFileTreeNodes {
			break
		}
		*children = append(*children, &response.FileNode{
			Name:    segments[len(segments)-1],
			Path:    relPath,
			IsDir:   false,
			Size:    entry.Size,
			ModTime: commitTimeMs,
		})
		count++
	}

	sortFileNodes(root)
	return root
}

// sortFileNodes 递归排序：目录优先，同类按名称升序，与草稿态 buildFileTree 的排序一致。
func sortFileNodes(nodes []*response.FileNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return nodes[i].Name < nodes[j].Name
	})
	for _, node := range nodes {
		if node.IsDir {
			sortFileNodes(node.Children)
		}
	}
}

// readSkillTreeBlob 读取指定版本 tag 下 skill 目录中的单个文件。
//
// 前端传入的是 files 接口解码后的 UTF-8 路径，而 git 中存的可能是存量 GBK 字节，
// 因此按 UTF-8 未命中时再用 GBK 编码重试一次，与草稿态 resolveDiskPath 的兜底策略同构。
func readSkillTreeBlob(skillDir, version, relPath string) (git_util.BlobInfo, error) {
	blob, err := git_util.ReadBlobAtTreeish(skillDir, version, joinSkillTreePath(relPath), maxFileSize)
	if err != nil {
		log.Errorf("[SkillContent] read blob dir=%s version=%s path=%s err: %v", skillDir, version, relPath, err)
		return git_util.BlobInfo{}, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_read_file_failed")
	}
	if blob.Exists {
		return blob, nil
	}
	gbkPath := util.EncodeUTF8ToGBK(relPath)
	if gbkPath == relPath {
		return blob, nil // 编码无变化（纯 ASCII 或含不可表示字符），无需兜底
	}
	gbkBlob, err := git_util.ReadBlobAtTreeish(skillDir, version, joinSkillTreePath(gbkPath), maxFileSize)
	if err != nil {
		log.Errorf("[SkillContent] read blob (gbk) dir=%s version=%s path=%s err: %v", skillDir, version, relPath, err)
		return blob, nil // GBK 兜底失败时按原路径的 not found 处理
	}
	if gbkBlob.Exists {
		return gbkBlob, nil
	}
	return blob, nil
}

// joinSkillTreePath 把相对 skill 内容根的路径拼成相对仓库根的路径。
func joinSkillTreePath(relPath string) string {
	return generalAgentWorkspaceSkillDirName + "/" + relPath
}


