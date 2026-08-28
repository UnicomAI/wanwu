package response

type UploadFileByBase64Resp struct {
	Url string `json:"url"`
	Uri string `json:"uri"`
}

type UnarchiveFileResp struct {
	ObjectPath string              `json:"objectPath"` // MinIO顶层对象路径（如 file-upload/file-expire/unarchive/uuid）
	Children   []UnarchiveFileNode `json:"children"`   // 目录树（从解压根目录的子节点开始）
	TotalFiles int                 `json:"totalFiles"` // 总文件数
	TotalSize  int64               `json:"totalSize"`  // 总文件大小（字节）
}

type UnarchiveFileNode struct {
	Name         string              `json:"name"`         // 文件或目录名
	Type         string              `json:"type"`         // "directory" 或 "file"
	ObjectPath   string              `json:"objectPath"`   // MinIO对象路径（如 file-upload/file-expire/unarchive/xxx/src/main.go）
	RelativePath string              `json:"relativePath"` // 相对路径（如 src/main.go）
	Size         int64               `json:"size"`         // 文件大小，字节（目录为0）
	MinioUrl     string              `json:"minioUrl"`     // MinIO内部访问地址（仅文件有值）
	DownloadUrl  string              `json:"downloadUrl"`  // 外部下载地址（仅文件有值）
	Children     []UnarchiveFileNode `json:"children"`     // 子节点（仅目录有）
}
