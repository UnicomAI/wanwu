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
	"runtime"
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

// IsSafePath 检查用户提供的路径是否安全，防止路径遍历攻击
//
// 该函数验证 userPath 是否在 baseDir 范围内，防止恶意用户通过 "../" 等方式访问预期目录之外的文件
// 主要用于处理用户可控的文件路径输入场景，如文件上传、下载、解压缩等操作
//
// 参数:
//   - baseDir: 基础目录路径（安全边界），userPath 必须在此目录及其子目录内
//   - userPath: 用户提供的相对路径或文件名，可以是包含目录结构的相对路径（如 "subdir/file.txt"）
//
// 返回值:
//   - bool: 路径是否安全（true=安全，false=不安全）
//   - string: 安全的绝对路径（如果验证通过），可以直接用于文件操作
//   - error: 错误信息（如果路径不安全或处理失败）
//
// 安全检查包括：
//  1. 路径遍历检查：禁止 ".." 等目录遍历字符
//  2. 边界检查：确保最终路径在 baseDir 范围内
//  3. 符号链接检查：解析符号链接后再次验证是否逃逸出基础目录
//  4. 跨平台兼容：正确处理 Windows 和 Unix 系统的路径差异
//
// 使用示例:
//
//	safe, path, err := IsSafePath("/data/uploads", "user123/avatar.jpg")
//	if !safe {
//	    return errors.New("invalid file path")
//	}
//	// 使用 path 进行文件操作
func IsSafePath(baseDir, userPath string) (bool, string, error) {
	if userPath == "" {
		return false, "", fmt.Errorf("path cannot be empty")
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false, "", fmt.Errorf("failed to get absolute base path: %w", err)
	}

	cleaned := filepath.Clean(userPath)

	// 防止路径遍历：检查清理后的路径是否包含".."
	// 注意：filepath.Clean 会处理掉多余的".."，但如果路径以".."开头，清理后仍然会有".."
	if strings.Contains(cleaned, "..") {
		return false, "", fmt.Errorf("path contains traversal sequences")
	}

	fullPath := filepath.Join(absBase, cleaned)

	// 使用 filepath.Abs 获取绝对路径，这不会要求路径存在
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false, "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 先规范化路径，处理大小写（Windows）
	normalizedPath := filepath.Clean(absPath)
	if runtime.GOOS == "windows" {
		normalizedPath = strings.ToLower(normalizedPath)
		absBase = strings.ToLower(absBase)
	}

	// 检查路径是否在基础目录内
	if !isPathWithinBase(absBase, normalizedPath) {
		return false, "", fmt.Errorf("path escapes base directory")
	}

	// 尝试解析符号链接，如果文件不存在则跳过
	var resolvedPath string
	if _, err := filepath.EvalSymlinks(absPath); err == nil {
		resolvedPath, err = filepath.EvalSymlinks(absPath)
		if err != nil {
			return false, "", fmt.Errorf("failed to resolve symlinks: %w", err)
		}

		// 重新检查解析后的路径
		if runtime.GOOS == "windows" {
			resolvedPath = strings.ToLower(resolvedPath)
		}
		if !isPathWithinBase(absBase, resolvedPath) {
			return false, "", fmt.Errorf("symlink escapes base directory")
		}
		resolvedPath = filepath.Clean(resolvedPath)
	} else {
		// 文件不存在，使用原始路径
		resolvedPath = absPath
	}

	return true, resolvedPath, nil
}

// ValidateFileName 验证文件名或相对路径是否安全合法，防止文件名注入和路径遍历攻击
//
// 该函数用于验证文件名或相对路径，确保符合操作系统规范和安全要求
// 适用于验证用户上传的文件名、导出文件名、资源标识符、相对路径等场景
//
// 参数:
//   - fileName: 待验证的文件名或相对路径（可以包含路径分隔符，如 "subdir/file.txt"）
//
// 返回值:
//   - error: 文件名不合法时返回错误，合法则返回 nil
//
// 验证规则：
//
//  1. 基本检查：
//     - 不能为空字符串
//     - 不能是 "." 或 ".."
//     - 路径总长度不能超过 4096 个字符
//     - 不能包含路径遍历字符 ".."（无论原始还是清理后）
//
//  2. 跨平台检查：
//     - Windows: 每个路径组件不能包含保留字符（<>:"|?*），不能以点或空格结尾，禁止保留设备名
//     - Unix/Linux: 禁止空字符（\x00）
//
// 注意事项：
//   - 该函数支持验证包含路径分隔符的相对路径
//   - 对于需要限制在特定目录内的完整路径验证，应使用 IsSafePath 函数
//   - 路径分隔符（/ 或 \）会被正确处理，但会检查每个路径组件的合法性
//
// 使用示例:
//
//	// 验证纯文件名
//	if err := ValidateFileName("document.pdf"); err != nil {
//	    return fmt.Errorf("invalid file name: %w", err)
//	}
//	// 验证相对路径
//	if err := ValidateFileName("subdir/document.pdf"); err != nil {
//	    return fmt.Errorf("invalid file path: %w", err)
//	}
func ValidateFileName(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if fileName == "." || fileName == ".." {
		return fmt.Errorf("filename cannot be '.' or '..'")
	}

	// 路径长度限制（大多数系统PATH_MAX为4096）
	if len(fileName) > 4096 {
		return fmt.Errorf("path too long")
	}

	// 检查原始路径是否包含 ".."
	if strings.Contains(fileName, "..") {
		return fmt.Errorf("path contains traversal sequences")
	}

	// 清理路径并再次检查
	cleaned := filepath.Clean(fileName)

	// 检查是否为绝对路径
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("absolute path not allowed")
	}

	// 操作系统特定检查
	if runtime.GOOS == "windows" {
		return validateWindowsPath(fileName)
	}

	return validateUnixPath(fileName)
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

// isPathWithinBase 检查 target 路径是否在 base 目录内（包含 base 本身）
// 用于 IsSafePath 的辅助函数，确保路径不会逃逸出基础目录
func isPathWithinBase(base, target string) bool {
	if target == base {
		return true
	}

	if !strings.HasPrefix(target, base) {
		return false
	}

	if len(target) > len(base) {
		sep := string(filepath.Separator)
		if !strings.HasPrefix(target[len(base):], sep) {
			return false
		}
	}

	return true
}

func validateWindowsPath(path string) error {
	// 分割路径，检查每个组件
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Windows保留字符（排除路径分隔符，因为在路径中是允许的）
		reservedChars := `<>:"|?*`
		if strings.ContainsAny(part, reservedChars) {
			return fmt.Errorf("path component contains invalid characters: %s", part)
		}

		// 检查结尾的点和空格
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("path component cannot end with dot or space: %s", part)
		}

		// 检查保留的设备名
		nameWithoutExt := strings.TrimSuffix(part, filepath.Ext(part))
		upperName := strings.ToUpper(nameWithoutExt)

		reservedNames := map[string]bool{
			"CON": true, "PRN": true, "AUX": true, "NUL": true,
			"COM1": true, "COM2": true, "COM3": true, "COM4": true,
			"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
			"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
			"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
		}

		if reservedNames[upperName] {
			return fmt.Errorf("path component is a reserved device name: %s", part)
		}

		// 单个组件长度限制（Windows MAX_PATH组件限制）
		if len(part) > 255 {
			return fmt.Errorf("path component too long: %s", part)
		}
	}

	return nil
}

func validateUnixPath(path string) error {
	// Unix/Linux 基本检查
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null character")
	}

	// 分割路径，检查每个组件
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}

		// 单个组件长度限制（大多数文件系统限制）
		if len(part) > 255 {
			return fmt.Errorf("path component too long: %s", part)
		}
	}

	return nil
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
