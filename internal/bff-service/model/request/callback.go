package request

import (
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
)

type FileUrlConvertBase64Req struct {
	FileUrl      string `form:"fileUrl" json:"fileUrl" validate:"required"` // 文件URL
	AddPrefix    bool   `form:"addPrefix" json:"addPrefix"`                 // 是否添加 data:xxx;base64, 前缀
	CustomPrefix string `form:"customPrefix" json:"customPrefix"`           // 自定义前缀（如 "data:video/mp4;base64,"）
}

func (f *FileUrlConvertBase64Req) Check() error {
	if f.FileUrl != "" {
		if err := url_util.ValidateURL(f.FileUrl); err != nil {
			return err
		}
	}
	return nil
}

type UploadFileByBase64Req struct {
	File     string `form:"file" json:"file" validate:"required"` // base64格式
	FileName string `form:"fileName" json:"fileName"`
	FileExt  string `form:"fileExt" json:"fileExt"` // 文件后缀名，如 "png", "pdf"
}

func (u *UploadFileByBase64Req) Check() error {
	return nil
}

type UnarchiveFileReq struct {
	FileUrl string `form:"fileUrl" json:"fileUrl" validate:"required"` // 压缩包的MinIO URL
}

func (u *UnarchiveFileReq) Check() error {
	if u.FileUrl != "" {
		if err := url_util.ValidateURL(u.FileUrl); err != nil {
			return err
		}
	}
	return nil
}

type GetUserListByUserIdsReq struct {
	UserIds []string `json:"userIds" validate:"required"` // 用户ID列表
}

func (u *GetUserListByUserIdsReq) Check() error {
	return nil
}
