package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CheckFile
//
//	@Tags			common.file
//	@Summary file verification
//	@Description Verify fragmented files
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CheckFileReq true "File verification parameters"
//	@Success		200		{object}	response.Response{data=response.CheckFileResp}
//	@Router			/file/check [get]
func CheckFile(ctx *gin.Context) {
	var req request.CheckFileReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.CheckFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// CheckFileList
//
//	@Tags			common.file
//	@Summary file list verification
//	@Description Verify fragmented file list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CheckFileListReq true "File list verification parameters"
//	@Success		200		{object}	response.Response{data=response.CheckFileListResp}
//	@Router			/file/check/list [get]
func CheckFileList(ctx *gin.Context) {
	var req request.CheckFileListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.CheckFileList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// UploadFile
//
//	@Tags			common.file
//	@Summary File upload
//	@Description Partial file upload
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param fileName formData string true "original file name"
//	@Param sequence formData int true "Fragment file sequence number"
//	@Param chunkName formData string true "Upload batch ID"
//	@Param files formData file true "file"
//	@Success		200			{object}	response.Response{data=response.UploadFileResp}
//	@Router			/file/upload [post]
func UploadFile(ctx *gin.Context) {
	var req request.UploadFileReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.UploadFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// MergeFile
//
//	@Tags			common.file
//	@Summary file merge
//	@Description Merge shard files and upload minio
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.MergeFileReq true "File merge parameters"
//	@Success		200		{object}	response.Response{data=response.MergeFileResp}
//	@Router			/file/merge [post]
func MergeFile(ctx *gin.Context) {
	var req request.MergeFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.MergeFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// CleanFile
//
//	@Tags			common.file
//	@Summary File Clearance
//	@Description Clear uploaded fragment files
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CleanFileReq true "File clearing parameter"
//	@Success		200		{object}	response.Response{data=response.CleanFileResp}
//	@Router			/file/clean [post]
func CleanFile(ctx *gin.Context) {
	var req request.CleanFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CleanFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// DeleteFile
//
//	@Tags			common.file
//	@Summary File deletion
//	@Description Delete uploaded files
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteFileReq true "File deletion request parameters"
//	@Success		200		{object}	response.Response{data=response.DeleteFileResp}
//	@Router			/file/delete [delete]
func DeleteFile(ctx *gin.Context) {
	var req request.DeleteFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.DeleteFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// ProxyUploadFile
//
//	@Tags			common.file
//	@Summary Agent file upload
//	@Description proxy file upload
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param fileName formData string true "original file name"
//	@Param file formData file true "file"
//	@Success		200			{object}	response.Response{data=response.ProxyUploadFileResp}
//	@Router			/proxy/file/upload [post]
func ProxyUploadFile(ctx *gin.Context) {
	var req request.ProxyUploadFileReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.ProxyUploadFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}
