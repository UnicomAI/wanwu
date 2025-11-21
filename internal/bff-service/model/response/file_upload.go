package response

type CheckFileResp struct {
	Status int `json:"status"` //0: Does not exist, 1: Completed
}

type UploadFileResp struct {
	Status int `json:"status"` //0: Upload failed, 1: Upload successful
}

type MergeFileResp struct {
	FileName string `json:"fileName"` //Merged file name
	FilePath string `json:"filePath"` //full path to minio file
}

type CleanFileResp struct {
	Status int `json:"status"` //0: Clearing failed, 1: Completed
}

type DeleteFileResp struct {
	Status int `json:"status"` //0: Deletion failed, 1: Completed
}

type CheckFileListResp struct {
	UploadedFileSequences []int `json:"uploadedFileSequences"` //List of slice file serial numbers that have been uploaded successfully
}

type ProxyUploadFileResp struct {
	DownloadLink string `json:"download_link"` //Upload file link
}
