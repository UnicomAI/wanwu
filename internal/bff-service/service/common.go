package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

var (
	// AvatarCachePrefix 前端访问avatar的URL前缀
	AvatarCachePrefix = "/v1/cache/avatar"
	// workflowAvatarCachePrefix 工作流avatar缓存路径前缀
	workflowAvatarCachePrefix = "/v1/cache/avatar/workflow"
	// mcpAvatarCachePrefix MCP服务avatar缓存路径前缀
	mcpAvatarCachePrefix = "/v1/cache/avatar/mcp"
	// customAvatarCachePrefix 系统自定义avatar缓存路径前缀
	customAvatarCachePrefix = "/v1/cache/avatar/custom"
)

func GetUserPermission(ctx *gin.Context, userID, orgID string) (*response.UserPermission, error) {
	resp, err := iam.GetUserPermission(ctx.Request.Context(), &iam_service.GetUserPermissionReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	user, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	return &response.UserPermission{
		UserInfo:         *toUserInfo(ctx, user),
		OrgPermission:    toOrgPermission(ctx, resp),
		IsUpdatePassword: resp.LastUpdatePasswordAt != 0,
	}, nil
}

func GetOrgSelect(ctx *gin.Context, userID string) (*response.Select, error) {
	resp, err := iam.GetOrgSelect(ctx.Request.Context(), &iam_service.GetOrgSelectReq{UserId: userID})
	if err != nil {
		return nil, err
	}
	return &response.Select{
		Select: toOrgIDNames(ctx, resp.Selects, userID == config.SystemAdminUserID),
	}, nil
}

// UploadAvatar 上传用户头像到MinIO
// 1. 校验文件类型（仅支持jpg/jpeg/png）
// 2. 读取文件内容到内存
// 3. 生成UUID文件名，构造存储路径 avatar/{前两位字母}/{文件名}
// 4. 上传到MinIO的custom-upload桶
// 返回MinIO中的objectPath，格式：custom-upload/avatar/xx/xxx.jpg
func UploadAvatar(ctx *gin.Context, fileHeader *multipart.FileHeader) (string, error) {
	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png":
	default:
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_type_error")
	}

	// 读取文件内容
	file, err := fileHeader.Open()
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	defer func() { _ = file.Close() }()

	// 读取图片到内存缓冲区
	imgBuf := new(bytes.Buffer)
	if _, err := io.Copy(imgBuf, file); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	fileName := fmt.Sprintf("%s%s", util.GenUUID(), ext)
	// 生成存储路径，avatar/fileName前两位字母/fileName
	objectName := path.Join("avatar", fileName[:2], fileName)
	objectPath := path.Join(minio.BucketCustom, objectName)

	if _, err = minio.Custom().PutObject(ctx.Request.Context(), minio.BucketCustom, objectName, imgBuf.Bytes()); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	return objectPath, nil
}

// CacheAvatar 将MinIO中的avatar objectPath转为前端可访问的缓存URL
// 输入：avatarObjectPath = "custom-upload/avatar/xx/xxx.jpg"
// 输出：avatar.Path = "/v1/cache/avatar/custom-upload/avatar/xx/xxx.jpg"
func CacheAvatar(avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		return avatar
	}

	avatar.Key = avatarObjectPath
	avatar.Path = path.Join(AvatarCachePrefix, avatarObjectPath)
	return avatar
}

// cacheCustomAvatar 将自定义avatar objectPath转为前端可访问的缓存URL
// 输入：avatarObjectPath = "custom-upload/avatar/xx/xxx.jpg"
// 输出：avatar.Path = "/v1/cache/avatar/custom/custom-upload/avatar/xx/xxx.jpg"
func cacheCustomAvatar(avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		return avatar
	}

	avatar.Key = avatarObjectPath
	avatar.Path = path.Join(customAvatarCachePrefix, avatarObjectPath)
	return avatar
}

// cacheAppAvatar 根据应用类型获取应用avatar的缓存URL
// 1. Rag类型空avatar返回默认Rag图标
// 2. Agent类型空avatar返回默认Agent图标
// 3. 其他类型调用CacheAvatar
func cacheAppAvatar(ctx *gin.Context, avatarObjectPath, appType string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" && appType == constant.AppTypeRag {
		avatar.Path = config.Cfg().DefaultIcon.RagIcon
		return avatar
	}
	if avatarObjectPath == "" && appType == constant.AppTypeAgent {
		avatar.Path = config.Cfg().DefaultIcon.AgentIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheUserAvatar 获取用户头像的缓存URL
// 1. 空avatar返回默认用户图标
// 2. 否则调用CacheAvatar转换
func cacheUserAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.UserIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheToolAvatar 根据工具类型获取工具avatar的缓存URL
// 1. Custom类型：空avatar返回默认工具图标，否则调用CacheAvatar
// 2. BuiltIn类型：调用cacheMCPServiceAvatar
func cacheToolAvatar(ctx *gin.Context, toolType string, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	switch toolType {
	case constant.ToolTypeCustom:
		if avatarObjectPath == "" {
			avatar.Path = config.Cfg().DefaultIcon.ToolIcon
			return avatar
		}
		return CacheAvatar(avatarObjectPath)
	case constant.ToolTypeBuiltIn:
		return cacheMCPServiceAvatar(avatarObjectPath)
	}
	return avatar
}

// cacheSkillAvatar 获取技能avatar的缓存URL
// 1. 空avatar返回默认技能图标
// 2. 否则调用CacheAvatar转换
func cacheSkillAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	if avatarObjectPath == "" {
		return request.Avatar{Path: config.Cfg().DefaultIcon.CustomSkillIcon}
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheMCPAvatar 获取MCP广场或自定义MCP的avatar缓存URL
// 1. squareObjectPath为空时：customObjectPath为空则返回MCP自定义默认图标，否则用customAvatarPrefix缓存
// 2. squareObjectPath非空时：调用cacheMCPServiceAvatar缓存
func cacheMCPAvatar(ctx *gin.Context, squareObjectPath, customObjectPath string) request.Avatar {
	if squareObjectPath == "" {
		avatar := request.Avatar{}
		if customObjectPath == "" {
			avatar.Path = config.Cfg().DefaultIcon.McpCustomIcon
			return avatar
		}
		return CacheAvatar(customObjectPath)
	}
	return cacheMCPServiceAvatar(squareObjectPath)
}

// cacheMCPServerAvatar 获取自定义MCP的avatar缓存URL
// 1. 空avatar返回默认自定义MCP图标
// 2. 否则调用CacheAvatar转换
func cacheMCPServerAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.McpServerIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheMCPServiceAvatar 获取MCP服务的avatar缓存URL
// 输入：avatarPath = "custom-upload/avatar/xx/xxx.jpg"
// 输出：avatar.Path = "/v1/cache/avatar/mcp/custom-upload/avatar/xx/xxx.jpg"
func cacheMCPServiceAvatar(avatarPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarPath == "" {
		return avatar
	}

	avatar.Key = avatarPath
	avatar.Path = path.Join(mcpAvatarCachePrefix, avatarPath)
	return avatar
}

// cacheWorkflowAvatar 将avatar http请求地址转为前端统一访问的格式
// 例如 http://IP:port/api/static/abc/def.png -> v1/cache/avatar/workflow/def.png
// 预签名URL会被保留完整路径，如 v1/cache/avatar/workflow/minio/presign/workflow/BIZ_BOT_WORKFLOW/1_xxx.jpg?签名
func cacheWorkflowAvatar(avatarURL, appType string) request.Avatar {
	avatar := request.Avatar{}
	switch appType {
	case constant.AppTypeWorkflow:
		if avatarURL == "" {
			avatar.Path = config.Cfg().DefaultIcon.WorkflowIcon
			return avatar
		}
	case constant.AppTypeChatflow:
		if avatarURL == "" {
			avatar.Path = config.Cfg().DefaultIcon.ChatflowIcon
			return avatar
		}
	}

	avatar.Key = avatarURL

	// 提取文件名：先去掉查询参数，再取最后一部分
	baseURL := avatarURL
	var queryParams string
	if idx := strings.Index(avatarURL, "?"); idx != -1 {
		baseURL = avatarURL[:idx]
		queryParams = avatarURL[idx:]
	}

	// 从路径中提取文件名（保留 bucket 路径）
	lastSlash := strings.LastIndex(baseURL, "/")
	fileName := baseURL[lastSlash+1:]

	// 构建缓存路径，保留完整路径信息
	// 例如：http://localhost:8081/workflow/minio/presign/workflow/BIZ_BOT_WORKFLOW/1_xxx.jpg -> /v1/cache/workflow/minio/presign/workflow/BIZ_BOT_WORKFLOW/1_xxx.jpg
	var cachePath string
	if strings.Contains(avatarURL, config.Cfg().Workflow.MinioProxyPrefix) {
		parsedURL, err := url.Parse(avatarURL)
		if err != nil {
			log.Errorf("cacheWorkflowAvatar parse URL %v failed: %v", avatarURL, err)
			avatar.Path = path.Join(AvatarCachePrefix, fileName)
			return avatar
		}
		cachePath = path.Join(workflowAvatarCachePrefix, parsedURL.Path)
	} else {
		cachePath = path.Join(workflowAvatarCachePrefix, fileName)
	}

	// 添加查询参数
	if queryParams != "" {
		cachePath += queryParams
	}
	avatar.Path = cachePath
	return avatar
}

// cachePromptAvatar 获取Prompt的avatar缓存URL
// 1. 空avatar返回默认Prompt图标
// 2. 否则调用CacheAvatar转换
func cachePromptAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.PromptIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheKnowledgeAvatar 获取知识库的avatar缓存URL
// 1. 空avatar根据knowledgeType返回对应默认图标（知识库/QA）
// 2. 否则调用CacheAvatar转换
func cacheKnowledgeAvatar(ctx *gin.Context, avatarObjectPath string, knowledgeType int32) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		switch knowledgeType {
		case constant.KnowledgeBase:
			avatar.Path = config.Cfg().DefaultIcon.KnowledgeIcon
		case constant.KnowledgeQA:
			avatar.Path = config.Cfg().DefaultIcon.QAIcon
		default:
			avatar.Path = config.Cfg().DefaultIcon.KnowledgeIcon
		}
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheOrgAvatar 获取组织头像的缓存URL
// 1. 空avatar返回默认组织图标
// 2. 否则调用CacheAvatar转换
func cacheOrgAvatar(avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.OrgIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheRoleAvatar 获取角色头像的缓存URL
// 1. 空avatar返回默认角色图标
// 2. 否则调用CacheAvatar转换
func cacheRoleAvatar(avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.RoleIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

// cacheModelAvatar 获取模型的avatar缓存URL
// 1. 空avatar返回默认模型图标
// 2. 否则调用CacheAvatar转换
func cacheModelAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.ModelIcon
		return avatar
	}
	return CacheAvatar(avatarObjectPath)
}

func convertStatisticChart(ctx *gin.Context, pbChart *common.StatisticChart) response.StatisticChart {
	if pbChart == nil {
		return response.StatisticChart{}
	}
	respChart := response.StatisticChart{
		TableName: gin_util.I18nKey(ctx, pbChart.TableName),
		Lines:     make([]response.StatisticChartLine, 0, len(pbChart.ChartLines)),
	}
	for _, pbLine := range pbChart.ChartLines {
		goLine := response.StatisticChartLine{
			LineName: gin_util.I18nKey(ctx, pbLine.LineName),
			Items:    make([]response.StatisticChartLineItem, 0, len(pbLine.Items)),
		}

		for _, pbItem := range pbLine.Items {
			goLine.Items = append(goLine.Items, response.StatisticChartLineItem{
				Key:   pbItem.Key,
				Value: pbItem.Value,
			})
		}
		respChart.Lines = append(respChart.Lines, goLine)
	}
	return respChart
}

func writeSSE(ctx *gin.Context, resp *http.Response) error {
	// 设置 SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("X-Accel-Buffering", "no") // 针对 Nginx 代理

	// 使用固定缓冲区读取
	buffer := make([]byte, 8192) // 8KB 缓冲区
	reader := bufio.NewReader(resp.Body)

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			return errors.New("writeSSE: ctx canceled")
		default:
			n, err := reader.Read(buffer)

			if n > 0 {
				if _, err := ctx.Writer.Write(buffer[:n]); err != nil {
					// 客户端可能已断开
					log.Errorf("writeSSE write err: %v", err)
					return err
				}
				ctx.Writer.Flush()
			}

			if err != nil {
				if err == io.EOF {
					return nil // 正常结束
				}
				log.Errorf("writeSSE read err: %v", err)
				return err
			}
		}
	}
}
