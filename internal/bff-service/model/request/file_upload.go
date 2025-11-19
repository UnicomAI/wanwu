package request

type CheckFileReq struct {
	FileName  string `json:"fileName" form:"fileName" validate:"required"`   //原始文件名 [EN] original file name
	Sequence  int    `json:"sequence" form:"sequence" validate:"gt=0"`       //分片文件序号 [EN] Fragment file serial number
	ChunkName string `json:"chunkName" form:"chunkName" validate:"required"` //上传批次标识 [EN] Upload batch ID
}

type UploadFileReq struct {
	FileName  string `json:"fileName" form:"fileName" validate:"required"`   //原始文件名 [EN] original file name
	Sequence  int    `json:"sequence" form:"sequence" validate:"gt=0"`       //分片文件序号 [EN] Fragment file serial number
	ChunkName string `json:"chunkName" form:"chunkName" validate:"required"` //上传批次标识 [EN] Upload batch ID
}

type MergeFileReq struct {
	FileName   string `json:"fileName" form:"fileName" validate:"required"`   //原始文件名 [EN] original file name
	FileSize   int64  `json:"fileSize" form:"fileSize" validate:"gt=0"`       //原始文件大小 [EN] Original file size
	ChunkName  string `json:"chunkName" form:"chunkName" validate:"required"` //上传批次标识 [EN] Upload batch ID
	ChunkTotal int    `json:"chunkTotal" form:"chunkTotal" validate:"gt=0"`   //分片总数 [EN] Total number of shards
	IsExpired  bool   `json:"isExpired" form:"isExpired"`                     //minio存储文件是否过期 0:过期，1:不过期 [EN] Whether the minio storage file has expired 0: expired, 1: not expired
}

type CleanFileReq struct {
	ChunkName string `json:"chunkName" form:"chunkName" validate:"required"` //上传批次标识 [EN] Upload batch ID
}

type CheckFileListReq struct {
	ChunkName string `json:"chunkName" form:"chunkName" validate:"required"` //上传批次标识 [EN] Upload batch ID
}

type DeleteFileReq struct {
	FileList  []string `json:"fileList" form:"fileList"`   //文件列表 [EN] file list
	IsExpired bool     `json:"isExpired" form:"isExpired"` //minio存储文件是否过期 0:过期，1:不过期 [EN] Whether the minio storage file has expired 0: expired, 1: not expired
}

type ProxyUploadFileReq struct {
	FileName string `json:"file_name" form:"file_name" validate:"required"` //原始文件名 [EN] original file name
}

func (c *CheckFileReq) Check() error {
	return nil
}
func (c *CheckFileListReq) Check() error {
	return nil
}
func (c *UploadFileReq) Check() error {
	return nil
}
func (c *MergeFileReq) Check() error {
	return nil
}
func (c *CleanFileReq) Check() error {
	return nil
}
func (c *DeleteFileReq) Check() error {
	return nil
}
func (c *ProxyUploadFileReq) Check() error {
	return nil
}
