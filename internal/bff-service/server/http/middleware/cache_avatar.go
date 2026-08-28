package middleware

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/imaging"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/gin-gonic/gin"
)

var (
	avatarCacheMu               sync.Mutex
	avatarCacheLocalDir         = "cache/avatar"
	mcpAvatarCacheLocalDir      = "cache/avatar/mcp"
	workflowAvatarCacheLocalDir = "cache/avatar/workflow"

	workflowMinioPresign     = "workflow/minio/presign/"
	workflowAvatarStaticIcon = "/api/static/icon/"

	avatarHTTPClient  = &http.Client{Timeout: 10 * time.Second}
	avatarMaxBodySize = int64(10 << 20) // 10MB
)

// CacheAvatar 缓存中间件，拦截 /v1/cache/avatar 路径的请求
// 1. 从URL中提取 relativePath（去掉前缀 /v1/cache/avatar）
// 2. 根据路径前缀分发处理：mcp/、workflow/、custom/、其他
// 3. 检查本地缓存是否已存在，存在则直接返回
// 4. 不存在则从远程下载并缓存到本地
func CacheAvatar(ctx *gin.Context) {
	urlPath := ctx.Request.URL.Path
	cacheIndex := strings.Index(urlPath, service.AvatarCachePrefix)
	prefixLen := len(service.AvatarCachePrefix)
	if cacheIndex == -1 {
		ctx.Next()
		return
	}
	// 提取 AvatarCachePrefix 后面的相对路径
	relativePath := urlPath[cacheIndex+prefixLen:]
	relativePath = strings.TrimPrefix(relativePath, "/")
	if relativePath == "" {
		ctx.Next()
		return
	}
	// 保留查询参数（如MinIO预签名URL的签名）
	if ctx.Request.URL.RawQuery != "" {
		relativePath += "?" + ctx.Request.URL.RawQuery
	}
	// 根据路径前缀分发处理
	if strings.HasPrefix(relativePath, "mcp/") {
		// MCP服务图片，去掉mcp/前缀后处理
		avatarPath := strings.TrimPrefix(relativePath, "mcp/")
		if handleMCPAvatar(ctx, avatarPath) {
			return
		}
	} else if strings.HasPrefix(relativePath, "workflow/") {
		// 工作流/对话流图片，去掉workflow/前缀后处理
		avatarPath := strings.TrimPrefix(relativePath, "workflow/")
		if handleWorkflowAvatarCache(ctx, avatarPath) {
			return
		}
	} else {
		// MinIO通用图片或用户自定义图片
		isResize := true
		realPath := relativePath
		if strings.HasPrefix(realPath, "custom/") {
			// custom/前缀表示用户自定义图片，不压缩
			realPath = strings.TrimPrefix(realPath, "custom/")
			isResize = false
		}
		avatarPath := downloadAndCacheAvatar(ctx, realPath, isResize)
		if avatarPath == "" {
			ctx.Next()
			return
		}
		if _, err := os.Stat(avatarPath); err == nil {
			ctx.File(avatarPath)
			ctx.Abort()
			return
		}
	}
	ctx.Next()
}

// handleWorkflowAvatarCache 处理工作流图片缓存请求
// 1. 判断URL类型：预签名URL 或 静态图标URL
// 2. 构造完整的请求URL
// 3. 调用 cacheWorkflowAvatar 下载并缓存
// 4. 返回本地缓存文件
func handleWorkflowAvatarCache(ctx *gin.Context, relativePath string) bool {
	var completeURL string
	// 判断是否为预签名URL
	if strings.Contains(relativePath, workflowMinioPresign) {
		// 预签名URL：直接拼接WebBaseUrl
		completeURL = config.Cfg().Server.WebBaseUrl + "/" + relativePath
	} else {
		// 静态图标URL：拼接WebBaseUrl和图标路径前缀
		completeURL, _ = url.JoinPath(config.Cfg().Server.WebBaseUrl, workflowAvatarStaticIcon, relativePath)
	}
	// 下载并缓存图片
	avatarPath := cacheWorkflowAvatar(ctx, completeURL, constant.AppTypeWorkflow)
	if avatarPath == "" {
		return false
	}
	// 去掉v1/前缀获取本地文件路径
	localPath := strings.TrimPrefix(avatarPath, "v1/")
	localPath = strings.TrimPrefix(localPath, "/")
	if _, err := os.Stat(localPath); err == nil {
		ctx.File(localPath)
		ctx.Abort()
		return true
	}
	return false
}

// handleMCPAvatar 处理MCP图片请求
// 1. 调用 cacheMCPAvatar 下载MCP图片到本地缓存
// 2. 返回本地缓存文件
func handleMCPAvatar(ctx *gin.Context, mcpPath string) bool {
	avatarPath := cacheMCPAvatar(ctx, mcpPath)
	if avatarPath == "" {
		return false
	}
	if _, err := os.Stat(avatarPath); err == nil {
		ctx.File(avatarPath)
		ctx.Abort()
		return true
	}
	return false
}

// downloadAndCacheAvatar 将MinIO中的avatar下载并缓存到本地
// 1. 解析 objectPath 获取 bucketName 和 objectName
// 2. 检查本地缓存是否存在（分普通路径和raw路径）
// 3. 从MinIO下载文件
// 4. 根据 isResize 决定是否压缩图片
// 5. 保存到本地缓存目录
// 返回本地缓存文件路径，失败返回空字符串
func downloadAndCacheAvatar(ctx *gin.Context, avatarObjectPath string, isResize bool) string {
	if avatarObjectPath == "" {
		return ""
	}
	// 去掉开头的斜杠
	avatarObjectPath = strings.TrimPrefix(avatarObjectPath, "/")
	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	// 解析bucketName和objectName
	parts := strings.SplitN(avatarObjectPath, "/", 2)
	if len(parts) <= 1 {
		log.Errorf("cache avatar %v err: invalid objectPath", avatarObjectPath)
		return ""
	}
	bucketName := parts[0]
	objectName := parts[1]

	// 构造本地缓存路径
	filePath := filepath.Join(avatarCacheLocalDir, objectName)
	if !isResize {
		// 用户自定义图片不压缩，使用raw子目录
		filePath = filepath.Join(avatarCacheLocalDir, "raw", objectName)
	}

	// 路径安全检查，防止../路径穿越
	if !strings.HasPrefix(filePath, avatarCacheLocalDir) {
		log.Warnf("avatar path traversal attempt: %s", avatarObjectPath)
		return ""
	}

	// 检查本地缓存是否存在
	_, err := os.Stat(filePath)
	if err == nil {
		return filePath
	}
	if !os.IsNotExist(err) {
		log.Errorf("cache avatar %v check %v exist err: %v", avatarObjectPath, filePath, err)
		return ""
	}

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache avatar %v mkdir %v err: %v", avatarObjectPath, filepath.Dir(filePath), err)
		return ""
	}

	// 从MinIO下载文件
	b, err := minio.Custom().GetObject(ctx.Request.Context(), bucketName, objectName)
	if err != nil {
		log.Errorf("cache avatar %v minio download err: %v", avatarObjectPath, err)
		return ""
	}

	// 根据需要压缩图片
	if isResize {
		compressedData, err := resizeImage(b)
		if err != nil {
			log.Warnf("cache avatar %v compress failed, using original: %v", avatarObjectPath, err)
			compressedData = b
		}
		if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
			log.Errorf("cache avatar %v write file %v err: %v", avatarObjectPath, filePath, err)
			return ""
		}
		return filePath
	}

	// 不压缩直接保存
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		log.Errorf("cache avatar %v write file %v err: %v", avatarObjectPath, filePath, err)
		return ""
	}
	return filePath
}

// cacheMCPAvatar 从MCP服务下载图片并缓存到本地
// 1. 检查本地缓存是否存在
// 2. 调用 MCP服务的 API 获取图片数据
// 3. 保存到本地缓存目录
// 返回本地缓存文件路径，失败返回空字符串
func cacheMCPAvatar(ctx *gin.Context, avatarPath string) string {
	if avatarPath == "" {
		return ""
	}
	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	filePath := filepath.Join(mcpAvatarCacheLocalDir, avatarPath)

	// 路径安全检查
	if !strings.HasPrefix(filePath, mcpAvatarCacheLocalDir) {
		log.Warnf("mcp avatar path traversal attempt: %s", avatarPath)
		return ""
	}

	// 检查缓存是否存在
	_, err := os.Stat(filePath)
	if err == nil {
		return filePath
	}
	if !os.IsNotExist(err) {
		log.Errorf("cache mcp avatar %v check %v exist err: %v", avatarPath, filePath, err)
		return ""
	}

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache mcp avatar %v mkdir %v err: %v", avatarPath, filepath.Dir(filePath))
		return ""
	}

	// 调用MCP服务API获取图片
	resp, err := service.GetMCPAvatar(ctx, avatarPath)
	if err != nil {
		log.Errorf("cache mcp avatar %v download err: %v", avatarPath, err)
		return ""
	}

	// 保存到本地
	if err := os.WriteFile(filePath, resp.Data, 0644); err != nil {
		log.Errorf("cache mcp avatar %v write file %v err: %v", avatarPath, filePath, err)
		return ""
	}
	return filePath
}

// cacheWorkflowAvatar 从远程URL下载工作流/对话流图片并缓存
// 1. 空URL返回默认图标
// 2. 提取文件名作为缓存key
// 3. 检查本地缓存是否存在
// 4. 转换URL为容器内可访问地址
// 5. 下载图片并压缩
// 6. 保存到本地缓存目录
// 返回本地缓存文件路径，失败返回空字符串
func cacheWorkflowAvatar(ctx *gin.Context, avatarURL, appType string) string {
	// 空URL返回默认图标
	switch appType {
	case constant.AppTypeWorkflow:
		if avatarURL == "" {
			return config.Cfg().DefaultIcon.WorkflowIcon
		}
	case constant.AppTypeChatflow:
		if avatarURL == "" {
			return config.Cfg().DefaultIcon.ChatflowIcon
		}
	}

	avatarCacheMu.Lock()
	defer avatarCacheMu.Unlock()

	// 提取文件名作为缓存key
	baseURL := avatarURL
	if idx := strings.Index(avatarURL, "?"); idx != -1 {
		baseURL = avatarURL[:idx]
	}
	lastSlash := strings.LastIndex(baseURL, "/")
	fileName := baseURL[lastSlash+1:]
	filePath := filepath.Join(workflowAvatarCacheLocalDir, fileName)

	// 检查缓存是否存在
	if _, err := os.Stat(filePath); err == nil {
		return filePath
	}

	// 转换URL为容器内可访问地址
	var newAvatarURL string
	if strings.Contains(avatarURL, config.Cfg().Workflow.MinioProxyPrefix) {
		// 预签名URL转换：替换MinioProxyPrefix为MinioProxyEndpoint
		parsedURL, err := url.Parse(avatarURL)
		if err != nil {
			log.Errorf("cache avatar parse URL %v failed: %v", avatarURL, err)
			return ""
		}
		path := parsedURL.Path
		path = strings.TrimPrefix(path, config.Cfg().Workflow.MinioProxyPrefix)
		newAvatarURL, err = url.JoinPath(config.Cfg().Workflow.MinioProxyEndpoint, path)
		if err != nil {
			log.Errorf("join path failed: %v", err)
			return ""
		}
		// 保留签名等查询参数
		if parsedURL.RawQuery != "" {
			newAvatarURL = newAvatarURL + "?" + parsedURL.RawQuery
		}
	} else if strings.Contains(avatarURL, config.Cfg().Server.WebBaseUrl) {
		// 普通URL，替换host为workflow内部Endpoint
		parsedURL, err := url.Parse(avatarURL)
		if err != nil {
			log.Errorf("cache avatar parse URL %v failed: %v", avatarURL, err)
			return ""
		}
		internalURL, err := url.Parse(config.Cfg().Workflow.Endpoint)
		if err != nil {
			log.Errorf("cache avatar invalid Workflow.Endpoint %s", config.Cfg().Workflow.Endpoint)
			return ""
		}
		parsedURL.Host = internalURL.Host
		parsedURL.Scheme = internalURL.Scheme
		newAvatarURL = parsedURL.String()
	} else {
		newAvatarURL = avatarURL
	}

	// 下载图片
	resp, err := avatarHTTPClient.Get(newAvatarURL)
	if err != nil {
		log.Errorf("cache avatar %v download %v err: %v", avatarURL, newAvatarURL, err)
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("cache avatar %v download %v HTTP error: %v", avatarURL, newAvatarURL, resp.Status)
		return ""
	}

	// 读取响应体，限制最大体积
	body, err := io.ReadAll(io.LimitReader(resp.Body, avatarMaxBodySize))
	if err != nil {
		log.Errorf("cache avatar %v download %v read response err: %v", avatarURL, newAvatarURL, err)
		return ""
	}
	if len(body) >= int(avatarMaxBodySize) {
		log.Warnf("cache avatar %v download %v body too large, truncated", avatarURL, newAvatarURL)
		return ""
	}

	// 压缩图片
	compressedData, err := resizeImage(body)
	if err != nil {
		log.Warnf("cache avatar %v compress failed, using original: %v", avatarURL, err)
		compressedData = body
	}

	// 保存到本地
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Errorf("cache avatar %v mkdir %v err: %v", avatarURL, filepath.Dir(filePath), err)
		return ""
	}
	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		log.Errorf("cache avatar %v write file %v err: %v", avatarURL, filePath, err)
		return ""
	}
	return filePath
}

// resizeImage 将图片压缩到最大200px（宽高都不超过200），保持宽高比
// 1. 解码图片获取原始尺寸
// 2. 计算等比例缩放后的尺寸
// 3. 使用imaging库进行缩放
// 返回压缩后的图片数据
func resizeImage(imageData []byte) ([]byte, error) {
	// 先解码获取图像尺寸
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()
	// 计算等比例缩放后的尺寸
	targetWidth, targetHeight := calculateResizeParameters(originalWidth, originalHeight, 200)
	// 重新创建 reader（因为之前的读取位置已经改变）
	reader := bytes.NewReader(imageData)
	// 压缩图像到计算后的尺寸
	compressedData, err := imaging.Resize(reader, targetWidth, targetHeight)
	if err != nil {
		return nil, fmt.Errorf("image resize failed: %w", err)
	}
	return compressedData, nil
}

// calculateResizeParameters 计算等比例缩放的目标尺寸
// 1. 如果原图宽高都小于等于maxSize，返回原尺寸
// 2. 否则按比例缩放，确保宽高都不超过maxSize
// 3. 最小尺寸为1x1
// 返回目标宽和高
func calculateResizeParameters(originalWidth, originalHeight, maxSize int) (int, int) {
	if originalWidth <= maxSize && originalHeight <= maxSize {
		// 如果原图已经小于目标尺寸，返回原尺寸
		return originalWidth, originalHeight
	}
	var newWidth, newHeight int
	if originalWidth > originalHeight {
		// 宽图：以宽度为基准
		newWidth = maxSize
		newHeight = int(float64(originalHeight) * float64(maxSize) / float64(originalWidth))
	} else {
		// 高图或正方形：以高度为基准
		newHeight = maxSize
		newWidth = int(float64(originalWidth) * float64(maxSize) / float64(originalHeight))
	}
	// 确保最小尺寸为1
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}
	return newWidth, newHeight
}
