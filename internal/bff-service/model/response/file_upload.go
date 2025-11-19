package response

type CheckFileResp struct {
	Status int `json:"status"` //0:不存在，1：已完成 [EN] 0: Does not exist, 1: Completed
}

type UploadFileResp struct {
	Status int `json:"status"` //0:上传失败，1：上传成功 [EN] 0: Upload failed, 1: Upload successful
}

type MergeFileResp struct {
	FileName string `json:"fileName"` //合并后文件名 [EN] Merged file name
	FilePath string `json:"filePath"` //minio文件的完整路径 [EN] full path to minio file
}

type CleanFileResp struct {
	Status int `json:"status"` //0:清除失败，1：已完成 [EN] 0: Clearing failed, 1: Completed
}

type DeleteFileResp struct {
	Status int `json:"status"` //0:删除失败，1：已完成 [EN] 0: Deletion failed, 1: Completed
}

type CheckFileListResp struct {
	UploadedFileSequences []int `json:"uploadedFileSequences"` //已经上传成功的切片文件序号列表 [EN] List of slice file serial numbers that have been uploaded successfully
}

type ProxyUploadFileResp struct {
	DownloadLink string `json:"download_link"` //上传文件链接 [EN] Upload file link
}
