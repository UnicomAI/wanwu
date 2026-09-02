package util

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/google/uuid"
)

const (
	kb               = 1024
	mb               = kb * 1024
	MaxScanTokenSize = 1024 * 1024 // Set the maximum token size to 1 MB
)

var specialFileExtList = []string{".tar.gz", ".tar.bz2"}

type FileInfo struct {
	IsDir    bool
	FilePath string
}

type FileMergeResult struct {
	TotalSuccessCount int64
	TotalLineCount    int64
	TotalByteCount    int64
	FilePath          string
}

// ============================================================================
// Public API Functions (公开接口函数)
// ============================================================================

func FileExt(filePath string) string {
	if len(filePath) == 0 {
		return ""
	}
	lower := strings.ToLower(filePath)
	for _, ext := range specialFileExtList {
		if strings.HasSuffix(lower, ext) {
			return ext
		}
	}
	return filepath.Ext(filePath)
}

func NewRandomFile(fileName string) string {
	extension := ExtractExtension(fileName)
	return uuid.New().String() + extension
}

// ExtractExtension 从路径或 URL 中安全提取扩展名（包含点号）
func ExtractExtension(raw string) string {
	// 1. 尝试解析为 URL
	u, err := url.Parse(raw)
	if err == nil && (u.Scheme != "" || u.Host != "" || strings.Contains(raw, "?")) {
		// 这是一个 URL，取其路径部分（忽略查询参数和片段）
		p := u.Path
		// 获取路径最后一段（文件名）
		base := path.Base(p)
		// 提取扩展名（path.Ext 能正确处理 .tar.gz 等多级扩展名）
		return path.Ext(base)
	}
	// 2. 非 URL，按普通文件路径处理（使用 filepath.Ext 适配 Windows 反斜杠）
	return filepath.Ext(raw)
}

// ToFileSizeStr fileSize单位是B，转换规则：小于1M为KB，大于等于1M，单位为M，保留两位小数
func ToFileSizeStr(fileSize int64) string {
	if fileSize < mb {
		return fmt.Sprintf("%.2f KB", float64(fileSize)/float64(kb))
	} else {
		return fmt.Sprintf("%.2f MB", float64(fileSize)/float64(mb))
	}
}

func FileExist(filePath string) (bool, error) {
	if len(filePath) == 0 {
		return false, nil
	}
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DirFileList(dir string, subDir bool, fullPath bool) ([]string, error) {
	fileList, err := FindDirAndFileList(dir, subDir, fullPath, false)
	if err != nil {
		return nil, err
	}
	var fileNameList []string
	for _, info := range fileList {
		fileNameList = append(fileNameList, info.FilePath)
	}
	return fileNameList, nil
}

func FindDirAndFileList(dir string, subDir bool, fullPath bool, dirPath bool) ([]*FileInfo, error) {
	var fileNameList []*FileInfo
	// 读取目录
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir (%v) err: %v", dir, err)
	}

	// 遍历目录下的所有文件和子目录
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// 处理错误
			log.Errorf("read dir (%v) entry err: %v", dir, err)
			continue
		}

		// 判断是否是文件
		if !info.IsDir() {
			if fullPath {
				fileNameList = append(fileNameList, &FileInfo{
					IsDir:    false,
					FilePath: dir + "/" + entry.Name(),
				})
			} else {
				fileNameList = append(fileNameList, &FileInfo{
					IsDir:    false,
					FilePath: entry.Name(),
				})
			}
			continue
		}

		if dirPath { // 需要返回目录名称
			if fullPath {
				fileNameList = append(fileNameList, &FileInfo{
					IsDir:    true,
					FilePath: dir + "/" + entry.Name(),
				})
			} else {
				fileNameList = append(fileNameList, &FileInfo{
					IsDir:    true,
					FilePath: entry.Name(),
				})
			}
		}

		if subDir { //需要校验底层目录
			list, err := FindDirAndFileList(dir+"/"+entry.Name(), subDir, fullPath, dirPath)
			if err != nil {
				return nil, err
			} else {
				fileNameList = append(fileNameList, list...)
			}
		}
	}

	return fileNameList, nil
}

// MergeFile 合并文件
func MergeFile(filePathList []string, mergeFilePath string) (*FileMergeResult, error) {
	// 创建或打开文件
	//0644，表示文件所有者可读写，同组用户及其他用户只可读
	dir := filepath.Dir(mergeFilePath)
	exist, err := FileExist(dir)
	if err != nil {
		return nil, err
	}
	if !exist {
		err = os.MkdirAll(filepath.Dir(mergeFilePath), 0755)
		if err != nil {
			return nil, err
		}
	}

	destinationFile, err := os.OpenFile(mergeFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open merge file (%v) err: %v", mergeFilePath, err)
	}
	defer func() {
		if err := destinationFile.Close(); err != nil {
			log.Errorf("close merge file (%v) err: %v", mergeFilePath, err)
		}
	}()

	var totalByteCount int64
	for _, fileInfo := range filePathList {
		byteCount, err := AppendFileStream(fileInfo, destinationFile)
		if err != nil {
			return nil, fmt.Errorf("merge file (%v) err: %v", mergeFilePath, err)
		}
		totalByteCount += byteCount
	}
	return &FileMergeResult{
		TotalByteCount: totalByteCount,
		FilePath:       mergeFilePath,
	}, nil
}

func MkDir(fileDir string) error {
	err := os.MkdirAll(fileDir, 0755)
	if err != nil {
		return fmt.Errorf("mkdir (%v) err: %v", fileDir, err)
	}
	return nil
}

func DeleteDir(fileDir string) error {
	err := os.RemoveAll(fileDir)
	if err != nil {
		return fmt.Errorf("delete dir (%v) err: %v", fileDir, err)
	}
	return nil
}

func DeleteFile(file string) error {
	err := os.Remove(file)
	if err != nil {
		return fmt.Errorf("delete file (%v) err: %v", file, err)
	}
	return nil
}

func AppendFileStream(filePath string, destinationFile *os.File) (int64, error) {
	// Open the source file for reading
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open append file (%v) err: %v", filePath, err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Errorf("close append file (%v) err: %v", filePath, err)
		}
	}()
	fileReader := bufio.NewReader(sourceFile)
	byteCount, err := appendFile(fileReader, destinationFile)
	if err != nil {
		return 0, fmt.Errorf("append file (%v) to (%v) err: %v", filePath, destinationFile.Name(), err)
	}
	log.Infof("append file (%v) to (%v) succeed, bytes: %v", filePath, destinationFile.Name(), byteCount)
	return byteCount, nil
}

func FileEOF(err error) bool {
	return errors.Is(err, io.EOF) || (err != nil && err.Error() == "EOF")
}

func File2Base64(filePath string, customPrefix string) (base64Str string, base64StrWithPrefix string, err error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}
	return FileData2Base64(fileData, customPrefix)
}

func FileData2Base64(fileData []byte, customPrefix string) (base64Str string, base64StrWithPrefix string, err error) {
	if len(fileData) == 0 {
		return "", "", errors.New("empty file data")
	}

	base64Str = base64.StdEncoding.EncodeToString(fileData)

	var prefix string
	if customPrefix != "" {
		prefix = customPrefix
	} else {
		// 自动检测 MIME 类型
		mimeType := http.DetectContentType(fileData)
		prefix = "data:" + mimeType + ";base64"
	}
	if !strings.Contains(prefix, ",") {
		prefix += ","
	}
	base64StrWithPrefix = prefix + base64Str

	return base64Str, base64StrWithPrefix, nil
}

// FileData2FileHeader
//
//	@Description: 将字节数组转换为multipart.FileHeader
//	@Author zhangzekai
//	@Time 2026-01-21 11:11:20
//	@param filename
//	@param fileData
//	@return *multipart.FileHeader
//	@return error
func FileData2FileHeader(filename string, fileData []byte) (*multipart.FileHeader, error) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
	header.Set("Content-Type", "application/octet-stream") // 可根据实际文件类型修改（如audio/wav）

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, fmt.Errorf("创建form字段失败: %w", err)
	}
	_, err = part.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("写入文件数据失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭form writer失败: %w", err)
	}

	reader := multipart.NewReader(buf, writer.Boundary())
	form, err := reader.ReadForm(int64(len(fileData)) + 1024)
	if err != nil {
		return nil, fmt.Errorf("解析form数据失败: %w", err)
	}

	fileHeaders := form.File["file"]
	if len(fileHeaders) == 0 {
		return nil, fmt.Errorf("form中未找到file字段")
	}

	return fileHeaders[0], nil
}

// ============================================================================
// Internal Helper Functions (内部辅助函数)
// ============================================================================

func appendFile(reader *bufio.Reader, destinationFile *os.File) (byteCount int64, error error) {
	buf := make([]byte, MaxScanTokenSize)
	for {
		n, err := reader.Read(buf)
		if FileEOF(err) { // 检查是否到达文件末尾
			break
		}
		if err != nil {
			log.Errorf("Error reading file: %s", err)
			return -1, err
		}
		line := buf[:n]
		bytesWritten, err := destinationFile.Write(line)
		if err != nil {
			log.Errorf("appendFile error %s", err)
			return -1, err
		}
		byteCount += int64(bytesWritten)
	}
	return byteCount, nil
}

// WriteFileAtomic 通过临时文件+rename 原子写入文件内容，失败时清理临时文件。
func WriteFileAtomic(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpName, filePath); err != nil {
		if removeErr := os.Remove(filePath); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		if renameErr := os.Rename(tmpName, filePath); renameErr != nil {
			return err
		}
	}
	return nil
}

// binaryDetectHeadBytes 是二进制启发式判断读取的头部字节数。
const binaryDetectHeadBytes = 8192

// IsLikelyBinaryFile 通过读取文件头部 8KB 是否含 NUL 字节，启发式判断是否为二进制文件。
func IsLikelyBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = file.Close() }()

	buf := make([]byte, binaryDetectHeadBytes)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	return IsLikelyBinaryData(buf[:n]), nil
}

// IsLikelyBinaryData 按与 IsLikelyBinaryFile 相同的规则判断内存中的字节是否为二进制内容，
// 供已经读入内存（如从 git 对象库取出）的数据复用。
func IsLikelyBinaryData(data []byte) bool {
	if len(data) > binaryDetectHeadBytes {
		data = data[:binaryDetectHeadBytes]
	}
	return slices.Contains(data, 0)
}

// TruncateUTF8 按字节上限安全截断 UTF-8 字符串，保证不切坏多字节序列。
func TruncateUTF8(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	truncated := s[:maxBytes]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated
}

// MatchGlobPatterns 检查路径是否匹配以逗号分隔的多个 glob 模式（任一命中即返回 true）。
// 同时尝试匹配整段路径与 basename，行为与 ripgrep --glob 类似。
func MatchGlobPatterns(relPath, pattern string) bool {
	if pattern == "" {
		return true
	}
	for _, p := range strings.Split(pattern, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if matched, err := path.Match(p, filepath.Base(relPath)); err == nil && matched {
			return true
		}
		if matched, err := path.Match(p, relPath); err == nil && matched {
			return true
		}
	}
	return false
}

// RecreateDir 删除已存在的目录然后重新创建，权限 0755。
func RecreateDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
