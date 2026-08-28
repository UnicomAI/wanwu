package response

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	queue_util "github.com/UnicomAI/wanwu/pkg/queue-util"
	pkg_util "github.com/UnicomAI/wanwu/pkg/util"
)

// mdDownloadLinkRegex 匹配 markdown 下载链接格式：[文件名](URL)
// 排除图片链接 ![...](...) ，只匹配非图片的链接
var mdDownloadLinkRegex = regexp.MustCompile(`(?s)\[([^\]]*?)\]\((https?://[^\s\)]+?)\)`)

type AgentFile struct {
	Name     string     `json:"name"`
	Size     int        `json:"size"`
	FileUrl  string     `json:"fileUrl"`
	FileType string     `json:"fileType"`
	Metadata *AgentMeta `json:"metadata"`
}

type AgentMeta struct {
	Desc     string `json:"desc"`
	CreateAt string `json:"createAt"`
	Name     string `json:"name"`
}

type DownloadFileInfo struct {
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	FileSize int64  `json:"fileSize"`
	CreateAt string `json:"createAt"`
}

// ParseDownloadFileInfoList 从 interface{} 转换为 []*DownloadFileInfo
// 支持以下输入类型:
// 1. []*DownloadFileInfo - 直接返回
// 2. []interface{} - 每个 element 是 map[string]interface{} 或 DownloadFileInfo
// 3. []map[string]interface{} - JSON 反序列化后的类型
func ParseDownloadFileInfoList(data interface{}) []*DownloadFileInfo {
	if data == nil {
		return nil
	}

	// 尝试直接类型断言
	if fileList, ok := data.([]*DownloadFileInfo); ok {
		return fileList
	}

	// 尝试 []interface{} 类型 (JSON 反序列化常见情况)
	if slice, ok := data.([]interface{}); ok {
		var result []*DownloadFileInfo
		for _, item := range slice {
			info := parseDownloadFileInfo(item)
			if info != nil {
				result = append(result, info)
			}
		}
		return result
	}

	// 尝试 []map[string]interface{} 类型
	if slice, ok := data.([]map[string]interface{}); ok {
		var result []*DownloadFileInfo
		for _, item := range slice {
			info := parseDownloadFileInfoFromMap(item)
			if info != nil {
				result = append(result, info)
			}
		}
		return result
	}

	log.Warnf("ParseDownloadFileInfoList: unsupported type %T", data)
	return nil
}

// parseDownloadFileInfo 从单个 interface{} 解析 DownloadFileInfo
func parseDownloadFileInfo(data interface{}) *DownloadFileInfo {
	if data == nil {
		return nil
	}

	// 尝试直接类型断言
	if info, ok := data.(*DownloadFileInfo); ok {
		return info
	}

	// 尝试 map[string]interface{} 类型 (JSON 反序列化)
	if m, ok := data.(map[string]interface{}); ok {
		return parseDownloadFileInfoFromMap(m)
	}

	log.Warnf("parseDownloadFileInfo: unsupported item type %T", data)
	return nil
}

// parseDownloadFileInfoFromMap 从 map[string]interface{} 解析 DownloadFileInfo
func parseDownloadFileInfoFromMap(m map[string]interface{}) *DownloadFileInfo {
	if m == nil {
		return nil
	}

	info := &DownloadFileInfo{}

	// FileName
	if v, ok := m["fileName"]; ok {
		if s, ok := v.(string); ok {
			info.FileName = s
		}
	}

	// FilePath
	if v, ok := m["filePath"]; ok {
		if s, ok := v.(string); ok {
			info.FilePath = s
		}
	}

	// FileSize
	if v, ok := m["fileSize"]; ok {
		switch val := v.(type) {
		case int64:
			info.FileSize = val
		case int:
			info.FileSize = int64(val)
		case float64:
			info.FileSize = int64(val)
		case float32:
			info.FileSize = int64(val)
		}
	}

	// CreateAt
	if v, ok := m["createAt"]; ok {
		if s, ok := v.(string); ok {
			info.CreateAt = s
		}
	}

	return info
}

type DownloadContext struct {
	DownloadMap         map[string][]*DownloadFileInfo
	DownloadList        [][]*DownloadFileInfo
	AllDownloadSize     int
	ContentDownloadList []*DownloadFileInfo                  //正文内容下载文件列表
	ContentQueue        *queue_util.OverridableCircularQueue //正文内容缓存队列，用于提取markdown下载文件
}

func NewDownloadContext() *DownloadContext {
	return &DownloadContext{
		DownloadMap:  make(map[string][]*DownloadFileInfo),
		DownloadList: make([][]*DownloadFileInfo, 0),
		ContentQueue: queue_util.NewOverridableQueue(defaultContentQueueSize),
	}
}

func (d *DownloadContext) AddDownloadFile(toolId string, fileList []*DownloadFileInfo) {
	if fileList == nil {
		fileList = make([]*DownloadFileInfo, 0)
	}
	d.DownloadMap[toolId] = fileList
	d.DownloadList = append(d.DownloadList, fileList)
	d.AllDownloadSize += len(fileList)
}

func (d *DownloadContext) AddContent(content string) {
	defer pkg_util.PrintPanicStack()
	d.ContentQueue.EnQueue(content)
	//有中间产物且队列已经满了，则尝试提取markdown下载文件
	if d.AllDownloadSize > 0 && d.ContentQueue.IsFull() {
		downloads := d.extractMarkdownDownloads()
		if len(downloads) > 0 {
			d.ContentDownloadList = append(d.ContentDownloadList, downloads...)
		}
	}
}

func (d *DownloadContext) ClearContent() {
	d.ContentQueue.Clear()
}

func (d *DownloadContext) ResponseFiles(finish int) []*AgentFile {
	defer pkg_util.PrintPanicStack()
	if finish != 1 || len(d.DownloadList) == 0 {
		return make([]*AgentFile, 0)
	}
	// 从缓存的正文内容中提取 markdown 下载文件，过滤并填充元数据
	downloadUrlList := d.filterAndFillMarkdownDownloads()
	var responseFiles []*AgentFile
	if len(downloadUrlList) > 0 {
		for _, file := range downloadUrlList {
			responseFiles = append(responseFiles, buildAgentFile(file))
		}
	}
	return responseFiles
}

// filterAndFillMarkdownDownloads 从正文缓存中提取 markdown 下载链接，
// 过滤出在 DownloadList 中存在的条目，并用 DownloadList 中的 FileSize 和 CreateAt 填充
func (d *DownloadContext) filterAndFillMarkdownDownloads() []*DownloadFileInfo {
	//最后多提取一次
	downloads := d.extractMarkdownDownloads()
	if len(downloads) > 0 {
		d.ContentDownloadList = append(d.ContentDownloadList, downloads...)
	}

	if len(d.ContentDownloadList) == 0 {
		return nil
	}

	// 按 FilePath(FileUrl) 对 ContentDownloadList 去重，保持原始顺序
	d.deduplicateContentDownloads()

	downloadUrlList := d.ContentDownloadList

	// 构建 DownloadList 中以 FilePath 为 key 的索引，用于快速查找
	existingFiles := make(map[string]*DownloadFileInfo)
	for _, batch := range d.DownloadList {
		for _, f := range batch {
			if f.FilePath != "" {
				existingFiles[f.FilePath] = f
			}
		}
	}

	var result []*DownloadFileInfo
	for _, item := range downloadUrlList {
		existing, ok := existingFiles[item.FilePath]
		if !ok {
			continue
		}
		// 用 DownloadList 中的 FileSize 和 CreateAt 填充
		item.FileSize = existing.FileSize
		item.CreateAt = existing.CreateAt
		// 如果 markdown 提取的 FileName 为空，也用 DownloadList 的填充
		if item.FileName == "" {
			item.FileName = existing.FileName
		}
		result = append(result, item)
	}
	return result
}

// deduplicateContentDownloads 按 FilePath(FileUrl) 对 ContentDownloadList 去重，保持原始顺序（首次出现的保留）
func (d *DownloadContext) deduplicateContentDownloads() {
	seen := make(map[string]struct{})
	j := 0
	for _, item := range d.ContentDownloadList {
		if item.FilePath == "" {
			d.ContentDownloadList[j] = item
			j++
			continue
		}
		if _, ok := seen[item.FilePath]; ok {
			continue
		}
		seen[item.FilePath] = struct{}{}
		d.ContentDownloadList[j] = item
		j++
	}
	d.ContentDownloadList = d.ContentDownloadList[:j]
}

// extractMarkdownDownloads 从队列缓存的正文内容中提取 markdown 下载文件并添加到 DownloadContext
func (d *DownloadContext) extractMarkdownDownloads() []*DownloadFileInfo {
	if d.ContentQueue == nil || d.ContentQueue.IsEmpty() {
		return make([]*DownloadFileInfo, 0)
	}
	content := d.ContentQueue.AllValue()
	return extractMarkdownDownloadFiles(content)
}

// extractMarkdownDownloadFiles 从文本中提取 markdown 格式的下载文件链接
// 仅提取 URL 中包含文件扩展名的链接（排除纯页面链接）
func extractMarkdownDownloadFiles(content string) []*DownloadFileInfo {
	matches := mdDownloadLinkRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	var fileList []*DownloadFileInfo
	for _, match := range matches {
		fileName := match[1]
		fileURL := match[2]

		if !util.IsValidFileURL(fileURL) {
			continue
		}

		// 如果链接文本为空，从URL中提取文件名
		if fileName == "" {
			fileName = util.ExtractFileNameFromURL(fileURL)
		}

		fileList = append(fileList, &DownloadFileInfo{
			FileName: fileName,
			FilePath: fileURL,
		})
	}
	return fileList
}

// buildAgentFile 构造AgentFile
func buildAgentFile(file *DownloadFileInfo) *AgentFile {
	fileExt := filepath.Ext(file.FileName)
	fileType := strings.TrimPrefix(fileExt, ".")
	fileNameWithoutExt := strings.TrimSuffix(file.FileName, fileExt)
	var fileName = file.FileName
	if len(fileName) == 0 {
		fileName = util.ExtractFileNameFromURL(file.FilePath)
	}

	return &AgentFile{
		Name:     fileName,
		Size:     int(file.FileSize),
		FileUrl:  file.FilePath,
		FileType: fileType,
		Metadata: &AgentMeta{
			CreateAt: file.CreateAt,
			Name:     fileNameWithoutExt,
		},
	}
}
