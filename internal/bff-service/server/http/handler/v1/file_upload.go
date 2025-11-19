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
//	@Summary		文件校验 [EN] @Summary file verification
//	@Description	校验分片文件 [EN] @Description Verify fragmented files
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CheckFileReq	true	"文件校验参数" [EN] @Param data body request.CheckFileReq true "File verification parameters"
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
//	@Summary		文件列表校验 [EN] @Summary file list verification
//	@Description	校验分片文件列表 [EN] @Description Verify fragmented file list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CheckFileListReq	true	"文件列表校验参数" [EN] @Param data body request.CheckFileListReq true "File list verification parameters"
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
//	@Summary		文件上传 [EN] @Summary File upload
//	@Description	分片文件上传 [EN] @Description Partial file upload
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			fileName	formData	string	true	"原始文件名" [EN] @Param fileName formData string true "original file name"
//	@Param			sequence	formData	int		true	"分片文件序号" [EN] @Param sequence formData int true "Fragment file sequence number"
//	@Param			chunkName	formData	string	true	"上传批次标识" [EN] @Param chunkName formData string true "Upload batch ID"
//	@Param			files		formData	file	true	"文件" [EN] @Param files formData file true "file"
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
//	@Summary		文件合并 [EN] @Summary file merge
//	@Description	合并分片文件，并上传minio [EN] @Description Merge shard files and upload minio
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MergeFileReq	true	"文件合并参数" [EN] @Param data body request.MergeFileReq true "File merge parameters"
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
//	@Summary		文件清除 [EN] @Summary File Clearance
//	@Description	清除已上传的分片文件 [EN] @Description Clear uploaded fragment files
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CleanFileReq	true	"文件清除参数" [EN] @Param data body request.CleanFileReq true "File clearing parameter"
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
//	@Summary		文件删除 [EN] @Summary File deletion
//	@Description	删除已上传的文件 [EN] @Description Delete uploaded files
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteFileReq	true	"文件删除请求参数" [EN] @Param data body request.DeleteFileReq true "File deletion request parameters"
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
//	@Summary		代理文件上传 [EN] @Summary Agent file upload
//	@Description	代理文件上传 [EN] @Description proxy file upload
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			fileName	formData	string	true	"原始文件名" [EN] @Param fileName formData string true "original file name"
//	@Param			file		formData	file	true	"文件" [EN] @Param file formData file true "file"
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
