package response

// SkillContentMeta 标识本次返回内容的来源。
// 已发布 skill 返回其最新发布版本的内容（source=published，version 为版本号）；
// 从未发布、或存量数据缺少对应版本 tag 时回退返回草稿工作区内容（source=draft）。
type SkillContentMeta struct {
	Source  string `json:"source"`            // published | draft
	Version string `json:"version,omitempty"` // 已发布版本号（draft 时为空）
}

const (
	SkillContentSourcePublished = "published"
	SkillContentSourceDraft     = "draft"
)

type SkillContentFilesResp struct {
	SkillContentMeta
	Files []*FileNode `json:"files"` // 文件树
}

type SkillContentFileResp struct {
	SkillContentMeta
	SkillWorkspaceFileResp
}


