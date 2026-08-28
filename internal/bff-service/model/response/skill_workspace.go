package response

type FileNode struct {
	Name     string      `json:"name"`     // 文件/文件夹名称
	Path     string      `json:"path"`     // 相对路径
	IsDir    bool        `json:"isDir"`    // 是否是目录
	Size     int64       `json:"size"`     // 文件大小（字节）
	ModTime  int64       `json:"modTime"`  // 修改时间（毫秒时间戳）
	Children []*FileNode `json:"children"` // 子节点（仅目录有）
}

type SkillWorkspaceFilesResp struct {
	Files []*FileNode `json:"files"` // 文件树
}

type SkillWorkspaceFileResp struct {
	Content string `json:"content"` // 文件内容
	Size    int64  `json:"size"`    // 文件大小
	ModTime int64  `json:"modTime"` // 修改时间
}

// UpdateSkillWorkspaceFileResp 文件更新响应
type UpdateSkillWorkspaceFileResp struct{}

type CreateSkillWorkspaceEntryResp struct {
	Entry *FileNode `json:"entry"`
}

type RenameSkillWorkspaceEntryResp struct {
	Entry *FileNode `json:"entry"`
}

type SkillWorkspaceUploadResp struct {
	Files []*FileNode `json:"files"`
}

type SearchResult struct {
	Path    string `json:"path"`    // 文件路径
	Line    int    `json:"line"`    // 行号
	Content string `json:"content"` // 匹配的行内容
}

type SkillWorkspaceSearchResp struct {
	Results   []*SearchResult `json:"results"`             // 搜索结果
	Total     int             `json:"total"`               // 本次返回数量
	Truncated bool            `json:"truncated,omitempty"` // 结果是否因达到上限而被截断
}

type GitCommitInfo struct {
	Hash    string   `json:"hash"`
	Message string   `json:"message"`
	Time    int64    `json:"time"`
	Tags    []string `json:"tags"`
}

type SkillWorkspaceGitLogResp struct {
	Commits []GitCommitInfo `json:"commits"`
}

type GitFileChange struct {
	Path       string `json:"path"`              // 文件路径（相对 workspace/）
	OldPath    string `json:"oldPath,omitempty"` // 旧路径（仅重命名时有值，空时省略）
	ChangeType string `json:"changeType"`        // "added" | "modified" | "deleted" | "renamed"
}

type SkillWorkspaceGitDiffResp struct {
	FromCommit   string          `json:"fromCommit"`
	ToCommit     string          `json:"toCommit"`
	Diff         string          `json:"diff"`         // unified diff 文本
	ChangedFiles []GitFileChange `json:"changedFiles"` // 变更文件列表
	OldContent   *string         `json:"oldContent,omitempty"`
	NewContent   *string         `json:"newContent,omitempty"`
}

type SkillWorkspaceGitFileResp struct {
	FilePath   string `json:"filePath"`
	Content    string `json:"content"`
	CommitHash string `json:"commitHash"`
}

type GitStatusFile struct {
	Path       string `json:"path"`
	OldPath    string `json:"oldPath,omitempty"` // 仅重命名时有值，空时省略
	ChangeType string `json:"changeType"`
	Staged     bool   `json:"staged"`
}

type GitStatusResp struct {
	Files []GitStatusFile `json:"files"`
}
