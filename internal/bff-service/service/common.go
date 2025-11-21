package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/imaging"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

var (
	avatarCacheMu       sync.Mutex
	avatarCacheLocalDir = "cache"

	mcpAvatarCacheLocalDir = "cache/mcp"
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
	})
	if err != nil {
		return nil, err
	}
	return &response.UserPermission{
		OrgPermission:    toOrgPermission(ctx, resp),
		Language:         getLanguageByCode(user.Language),
		IsUpdatePassword: resp.LastUpdatePasswordAt != 0,
		Avatar:           cacheUserAvatar(ctx, user.AvatarPath),
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

// UploadAvatar returns the objectPath of avatar in minio
func UploadAvatar(ctx *gin.Context, fileHeader *multipart.FileHeader) (string, error) {
	// Verify file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png":
	default:
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_type_error")
	}

	// Read file contents
	file, err := fileHeader.Open()
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	defer file.Close()

	// Read the image into the memory buffer
	imgBuf := new(bytes.Buffer)
	if _, err := io.Copy(imgBuf, file); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	fileName := fmt.Sprintf("%s%s", util.GenUUID(), ext)
	// Generate the storage path, the first two letters of avatar/fileName/fileName
	objectName := path.Join("avatar", fileName[:2], fileName)
	objectPath := path.Join(minio.BucketCustom, objectName)

	if _, err = minio.Custom().PutObject(ctx.Request.Context(), minio.BucketCustom, objectName, imgBuf.Bytes()); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFInvalidArg, "bff_avatar_upload_error", err.Error())
	}
	return objectPath, nil
}

// CacheAvatar converts avatar's objectPath in minio to an address accessible to the front end and caches avatar locally.
// For example custom-upload/avatar/abc/def.png => /v1/static/avatar/abc/def.png
func CacheAvatar(ctx *gin.Context, avatarObjectPath string, isResize bool) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		return avatar
	}
	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	avatar.Key = avatarObjectPath

	parts := strings.SplitN(avatarObjectPath, "/", 2)
	if len(parts) <= 1 {
		log.Errorf("cache avatar %v err: invalid objectPath", avatarObjectPath)
		return avatar
	}
	bucketName := parts[0]
	objectName := parts[1]
	filePath := filepath.Join(avatarCacheLocalDir, objectName)

	_, err := os.Stat(filePath)
	// 1 file exists
	if err == nil {
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	// 2 system error
	if !os.IsNotExist(err) {
		log.Errorf("cache avatar %v check %v exist err: %v", avatarObjectPath, filePath, err)
		return avatar
	}
	// 3 file does not exist
	// 3.1 Create directory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache avatar %v mkdir %v err: %v", avatarObjectPath, filepath.Dir(filePath), err)
		return avatar
	}
	// 3.2 Download files
	b, err := minio.Custom().GetObject(ctx.Request.Context(), bucketName, objectName)
	if err != nil {
		log.Errorf("cache avatar %v minio download err: %v", avatarObjectPath, err)
		return avatar
	}
	// 3.3 Compressed images
	if isResize {
		compressedData, err := resizeImage(b)
		if err != nil {
			log.Warnf("cache avatar %v compress failed, using original: %v", avatarObjectPath, err)
			compressedData = b
		}
		// 3.3.1 Writing compressed files
		if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
			log.Errorf("cache avatar %v write file %v err: %v", avatarObjectPath, filePath, err)
			return avatar
		}
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	// 3.4 Write original file
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		log.Errorf("cache avatar %v write file %v err: %v", avatarObjectPath, filePath, err)
		return avatar
	}
	avatar.Path = filepath.Join("/v1", filePath)
	return avatar
}

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
	return CacheAvatar(ctx, avatarObjectPath, true)
}

func cacheUserAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.UserIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// tool builtin & custom
func cacheToolAvatar(ctx *gin.Context, toolType string, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	switch toolType {
	case constant.ToolTypeCustom:
		if avatarObjectPath == "" {
			avatar.Path = config.Cfg().DefaultIcon.ToolIcon
			return avatar
		}
		return CacheAvatar(ctx, avatarObjectPath, true)
	case constant.ToolTypeBuiltIn:
		return cacheMCPServiceAvatar(ctx, avatarObjectPath)
	}
	return avatar
}

// mcp square & custom
func cacheMCPAvatar(ctx *gin.Context, squareObjectPath, customObjectPath string) request.Avatar {
	if squareObjectPath == "" {
		avatar := request.Avatar{}
		if customObjectPath == "" {
			avatar.Path = config.Cfg().DefaultIcon.McpCustomIcon
			return avatar
		}
		return CacheAvatar(ctx, customObjectPath, true)
	}
	return cacheMCPServiceAvatar(ctx, squareObjectPath)
}

// mcp server
func cacheMCPServerAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.McpServerIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// Used to cache images of built-in tools and MCP Square (from mcp-service)
func cacheMCPServiceAvatar(ctx *gin.Context, avatarPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarPath == "" {
		return avatar
	}
	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	filePath := filepath.Join(mcpAvatarCacheLocalDir, avatarPath)

	_, err := os.Stat(filePath)
	// 1 file exists
	if err == nil {
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	// 2 system error
	if !os.IsNotExist(err) {
		log.Errorf("cache mcp avatar %v check %v exist err: %v", avatarPath, filePath, err)
		return avatar
	}
	// 3. The file does not exist
	// 3.1 Create directory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache mcp avatar %v mkdir %v err: %v", avatarPath, filepath.Dir(filePath))
		return avatar
	}
	// 3.2 Download files
	resp, err := mcp.GetMCPAvatar(ctx.Request.Context(), &mcp_service.GetMCPAvatarReq{AvatarPath: avatarPath})
	if err != nil {
		log.Errorf("cache mcp avatar %v download err: %v", avatarPath, err)
		return avatar
	}
	// 3.3 Writing files
	if err := os.WriteFile(filePath, resp.Data, 0644); err != nil {
		log.Errorf("cache mcp avatar %v write file %v err: %v", avatarPath, filePath, err)
		return avatar
	}
	avatar.Path = filepath.Join("/v1", filePath)
	return avatar
}

// cacheWorkflowAvatar converts the avatar http request address into a format for unified access by the front end, and caches avatar locally.
// For example http://IP:port/api/static/abc/def.jpg => /v1/static/avatar/abc/def.png
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

	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	avatar.Key = avatarURL

	// Extract the file name: remove the query parameters first, then take the last part
	baseURL := avatarURL
	if idx := strings.Index(avatarURL, "?"); idx != -1 {
		baseURL = avatarURL[:idx]
	}
	// Extract filename from path
	lastSlash := strings.LastIndex(baseURL, "/")
	fileName := baseURL[lastSlash+1:]
	filePath := filepath.Join(avatarCacheLocalDir, fileName)
	// Check if the file is cached
	if _, err := os.Stat(filePath); err == nil {
		avatar.Path = filepath.Join("/v1", filePath)
		return avatar
	}
	var newAvatarURL string
	if strings.Contains(avatarURL, config.Cfg().Workflow.MinioProxyPrefix) {
		// Parse original URL
		parsedURL, err := url.Parse(avatarURL)
		if err != nil {
			log.Errorf("parse avatar URL %v failed: %v", avatarURL, err)
			return avatar
		}
		// Remove the /workflow/minio/presign/ prefix
		path := parsedURL.Path
		path = strings.TrimPrefix(path, config.Cfg().Workflow.MinioProxyPrefix)
		// Use url.JoinPath to build new URLs
		newAvatarURL, err = url.JoinPath(config.Cfg().Workflow.MinioProxyEndpoint, path)
		if err != nil {
			log.Errorf("join path failed: %v", err)
			avatar.Path = avatarURL
			return avatar
		}
		// Add query parameters
		if parsedURL.RawQuery != "" {
			newAvatarURL += "?" + parsedURL.RawQuery
		}
	} else {
		// Use the original URL directly (such as http://localhost:8081/api/static/icon/icon-HTTP.png)
		newAvatarURL = avatarURL
	}
	// Download files from HTTP URL
	resp, err := http.Get(newAvatarURL)
	if err != nil {
		log.Errorf("cache avatar %v download err: %v", avatarURL, err)
		return avatar
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("cache avatar %v HTTP error: %v", avatarURL, resp.Status)
		return avatar
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("cache avatar %v read response err: %v", avatarURL, err)
		return avatar
	}
	// Compress images
	compressedData, err := resizeImage(body)
	if err != nil {
		log.Warnf("cache avatar %v compress failed, using original: %v", avatarURL, err)
		// Use original data when compression fails
		compressedData = body
	}
	// Create directory
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache avatar %v mkdir %v err: %v", avatarURL, filepath.Dir(filePath), err)
		return avatar
	}
	// write file
	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		log.Errorf("cache avatar %v write file %v err: %v", avatarURL, filePath, err)
		return avatar
	}
	avatar.Path = filepath.Join("/v1", filePath)
	return avatar
}

func cachePromptAvatar(ctx *gin.Context, avatarObjectPath string) request.Avatar {
	avatar := request.Avatar{}
	if avatarObjectPath == "" {
		avatar.Path = config.Cfg().DefaultIcon.PromptIcon
		return avatar
	}
	return CacheAvatar(ctx, avatarObjectPath, true)
}

// resizeImage compressed image
func resizeImage(imageData []byte) ([]byte, error) {
	// Decode first to get the image size
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()
	// Calculate the size after scaling
	targetWidth, targetHeight := calculateResizeParameters(originalWidth, originalHeight, 200)
	// Re-create the reader (because the previous reading position has changed)
	reader := bytes.NewReader(imageData)
	// Compress image to calculated dimensions
	compressedData, err := imaging.Resize(reader, targetWidth, targetHeight)
	if err != nil {
		return nil, fmt.Errorf("image resize failed: %w", err)
	}
	return compressedData, nil
}

// Calculate proportional scaling dimensions
func calculateResizeParameters(originalWidth, originalHeight, maxSize int) (int, int) {
	if originalWidth <= maxSize && originalHeight <= maxSize {
		// If the original image is smaller than the target size, return to the original size
		return originalWidth, originalHeight
	}
	var newWidth, newHeight int
	if originalWidth > originalHeight {
		// Wide image: based on width
		newWidth = maxSize
		newHeight = int(float64(originalHeight) * float64(maxSize) / float64(originalWidth))
	} else {
		// Heightmap or square: based on height
		newHeight = maxSize
		newWidth = int(float64(originalWidth) * float64(maxSize) / float64(originalHeight))
	}
	// Make sure the minimum size is 1
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}
	return newWidth, newHeight
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
	// Set SSE response headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("X-Accel-Buffering", "no") // For Nginx proxy

	// Read using fixed buffer
	buffer := make([]byte, 8192) // 8KB buffer
	reader := bufio.NewReader(resp.Body)

	for {
		select {
		case <-ctx.Done():
			// Client disconnects
			return errors.New("writeSSE: ctx canceled")
		default:
			n, err := reader.Read(buffer)

			if n > 0 {
				if _, err := ctx.Writer.Write(buffer[:n]); err != nil {
					// The client may have been disconnected
					log.Errorf("writeSSE write err: %v", err)
					return err
				}
				ctx.Writer.Flush()
			}

			if err != nil {
				if err == io.EOF {
					return nil // End normally
				}
				log.Errorf("writeSSE read err: %v", err)
				return err
			}
		}
	}
}
