package util

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
)

var urlKeys = []string{
	"output", "url", "fileUrl", "file_url", "downloadUrl", "download_url", "path", "uri",
	"link", "file", "address", "src", "callbackUrl", "callback_url", "resource",
}

var imageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".svg":  true,
	".tiff": true,
}

var zipExt = map[string]bool{
	".zip": true,
	".rar": true,
	".7z":  true,
}

// ImageFile 判断文件是否为图片
func ImageFile(fileName string) bool {
	return imageExt[strings.ToLower(filepath.Ext(fileName))]
}

// ZipFile 判断文件是否为压缩包
func ZipFile(fileName string) bool {
	if strings.HasSuffix(fileName, ".tar.gz") {
		return true
	}
	return zipExt[strings.ToLower(filepath.Ext(fileName))]
}

// ExtractFileNameFromURL 从URL中提取文件名
func ExtractFileNameFromURL(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return ""
	}

	path := parsedURL.Path
	if path == "" || path == "/" {
		return ""
	}

	// 获取路径的最后一部分
	fileName := filepath.Base(path)

	// 如果有查询参数中的文件名，优先使用
	if queryName := parsedURL.Query().Get("filename"); queryName != "" {
		fileName = queryName
	} else if queryName = parsedURL.Query().Get("name"); queryName != "" {
		fileName = queryName
	} else if queryName = parsedURL.Query().Get("file"); queryName != "" {
		fileName = queryName
	}

	return fileName
}

// ExtractURLFromJSON 通用URL提取工具
func ExtractURLFromJSON(jsonStr string) (string, error) {
	// 1. 把JSON解析为通用map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("JSON解析失败: %v", err)
	}

	// 2. 优先从常用key中查找URL
	for _, key := range urlKeys {
		if val, ok := data[key]; ok {
			if str, ok := val.(string); ok && isValidURL(str) {
				if !imageExt[filepath.Ext(str)] {
					return str, nil
				}
			}
		}
	}

	// 3. 兜底：遍历所有key，找到第一个符合URL格式的字符串
	for _, val := range data {
		if str, ok := val.(string); ok && isValidURL(str) {
			if !imageExt[filepath.Ext(str)] {
				return str, nil
			}
		}
	}

	return "", fmt.Errorf("未找到有效的URL")
}

// IsValidFileURL 检查字符串是否是有效的文件URL
func IsValidFileURL(s string) bool {
	if len(s) < 5 {
		return false
	}
	// 检查是否是HTTP/HTTPS URL
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		// 尝试解析URL
		parsedURL, err := url.Parse(s)
		if err != nil {
			return false
		}
		// 检查是否有路径部分（包含文件名）
		path := parsedURL.Path
		if path != "" && path != "/" {
			// 检查路径是否包含文件扩展名
			ext := filepath.Ext(path)
			return ext != ""
		}
	}
	return false
}

// 判断字符串是否为有效URL（宽松匹配，兼容内网IP、中文文件名）
func isValidURL(s string) bool {
	// 先尝试标准URL解析
	if _, err := url.Parse(s); err == nil {
		return true
	}
	// 额外支持：内网地址、本地文件地址、带中文的下载链接
	// 匹配 http/https 开头 + 包含路径格式
	matched := false
	reflect.ValueOf(&matched)
	return len(s) > 10 && (startsWith(s, "http://") || startsWith(s, "https://"))
}

// 字符串前缀判断
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
