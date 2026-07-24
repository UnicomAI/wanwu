package util

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 支持的压缩格式
var supportedArchiveExts = []string{
	".zip", ".tar", ".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".gz",
}

// IsSupportedArchive 判断文件名是否为支持的压缩格式
func IsSupportedArchive(filename string) bool {
	ext := getArchiveExt(filename)
	for _, supported := range supportedArchiveExts {
		if ext == supported {
			return true
		}
	}
	return false
}

// getArchiveExt 获取压缩文件扩展名（支持 .tar.gz 等多级扩展名）
func getArchiveExt(filename string) string {
	lower := strings.ToLower(filename)
	// 先检查多级扩展名
	for _, ext := range []string{".tar.gz", ".tar.bz2"} {
		if strings.HasSuffix(lower, ext) {
			return ext
		}
	}
	return strings.ToLower(filepath.Ext(filename))
}

// Unarchive 解压压缩包到指定目录，根据文件名后缀自动判断格式
// 支持格式：.zip, .tar, .tar.gz/.tgz, .tar.bz2/.tbz2, .gz
// reader: 压缩包数据流
// filename: 原始文件名（用于判断格式）
// destDir: 目标目录
func Unarchive(ctx context.Context, reader io.Reader, filename string, destDir string) error {
	ext := getArchiveExt(filename)

	// 确保目标目录存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dest dir error: %w", err)
	}

	switch ext {
	case ".zip":
		return unzipFromReader(reader, destDir)
	case ".tar":
		return untar(reader, destDir)
	case ".tar.gz", ".tgz":
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("create gzip reader error: %w", err)
		}
		defer func() {
			if err := gzReader.Close(); err != nil {
				log.Errorf("close gzip reader error: %v", err)
			}
		}()
		return untar(gzReader, destDir)
	case ".tar.bz2", ".tbz2":
		bz2Reader := bzip2.NewReader(reader)
		return untar(bz2Reader, destDir)
	case ".gz":
		return ungz(reader, destDir, filename)
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// UnarchiveFile 解压本地压缩包文件到指定目录
func UnarchiveFile(ctx context.Context, filePath string, destDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open archive file error: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("close archive file error: %v", err)
		}
	}()

	return Unarchive(ctx, file, filepath.Base(filePath), destDir)
}

// unzipFromReader 从 io.Reader 解压 zip 格式
// 自行实现 zip 解压逻辑，不复用 UnzipDir（UnzipDir 在文件 Mode 为 0 时会创建无权限目录）
func unzipFromReader(reader io.Reader, destDir string) error {
	// zip 需要读取文件大小信息，先保存到临时文件
	tmpFile, err := os.CreateTemp("", "unzip-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file error: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			log.Errorf("remove temp file error: %v", err)
		}
	}()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file error: %w", err)
	}
	_ = tmpFile.Close()

	fileReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		return fmt.Errorf("open zip reader error: %w", err)
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			log.Errorf("close zip reader error: %v", err)
		}
	}()

	for _, f := range fileReader.File {
		// 处理 GBK 编码文件名
		var decodeFileName string
		if f.Flags == 0 {
			i := bytes.NewReader([]byte(f.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := io.ReadAll(decoder)
			decodeFileName = string(content)
		} else {
			decodeFileName = f.Name
		}

		// 安全检查：防止路径遍历
		if err := ValidateArchivePath(decodeFileName); err != nil {
			log.Errorf("skip unsafe zip entry: %s, error: %v", decodeFileName, err)
			continue
		}

		targetPath := filepath.Join(destDir, filepath.Clean(decodeFileName))

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("create directory %s error: %w", targetPath, err)
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("create parent directory error: %w", err)
		}

		// 写入文件
		if err := writeUnzipArchiveFile(f, targetPath); err != nil {
			return fmt.Errorf("write file %s error: %w", targetPath, err)
		}
	}
	return nil
}

// writeUnzipArchiveFile 将 zip 中的文件写入目标路径
func writeUnzipArchiveFile(zipFile *zip.File, targetPath string) error {
	source, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := source.Close(); err != nil {
			log.Errorf("close zip source file error: %v", err)
		}
	}()

	perm := zipFile.Mode().Perm()
	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, IfElse(perm == 0, os.FileMode(0644), perm))
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Errorf("close output file error: %v", err)
		}
	}()

	_, err = io.Copy(outFile, source)
	return err
}

// untar 解压 tar 格式
func untar(reader io.Reader, destDir string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry error: %w", err)
		}

		// 安全检查：防止路径遍历
		if err := ValidateArchivePath(header.Name); err != nil {
			log.Errorf("skip unsafe tar entry: %s, error: %v", header.Name, err)
			continue
		}

		targetPath := filepath.Join(destDir, filepath.Clean(header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory %s error: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create parent directory error: %w", err)
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file %s error: %w", targetPath, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return fmt.Errorf("write file %s error: %w", targetPath, err)
			}
			_ = outFile.Close()
		case tar.TypeSymlink:
			// 跳过符号链接以确保安全
			log.Errorf("skip symlink in tar: %s -> %s", header.Name, header.Linkname)
		default:
			log.Errorf("skip unsupported tar entry type %d: %s", header.Typeflag, header.Name)
		}
	}
	return nil
}

// ungz 解压 gz 文件。
// 当压缩文件的后缀传递不完整（如 .tar.gz 变为 .gz）时，
// Unarchive 会误判格式进入 ungz 分支，仅解压 gzip 层而不展开 tar，
// 导致输出为 tar 二进制文件而非解压目录。
// 因此 ungz 需要自动检测解压内容是否为 tar 流，若是则转交 untar 处理。
// 检测优先级：文件名后缀 (.tar) > 内容嗅探 (ustar 魔数)
func ungz(reader io.Reader, destDir string, originalFilename string) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("create gzip reader error: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Errorf("close gzip reader error: %v", err)
		}
	}()

	// 从 gzip header 中获取原始文件名，如果没有则去除 .gz 后缀
	outputName := gzReader.Name
	if outputName == "" {
		outputName = strings.TrimSuffix(filepath.Base(originalFilename), ".gz")
		if outputName == "" {
			outputName = "decompressed_file"
		}
	}

	// 如果去除 .gz 后缀后文件名以 .tar 结尾，直接走 tar 解压
	if strings.HasSuffix(strings.ToLower(outputName), ".tar") {
		return untar(gzReader, destDir)
	}

	// 文件名无法判断时，通过内容嗅探检测是否为 tar 流
	// tar 文件在偏移 257 处有 "ustar" 魔数，读取前 512 字节（一个 tar block）用于检测
	peekBuf := make([]byte, 512)
	n, readErr := io.ReadFull(gzReader, peekBuf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return fmt.Errorf("read gzip content for tar detection error: %w", readErr)
	}

	if n >= 262 && string(peekBuf[257:262]) == "ustar" {
		// 内容是 tar 格式，将偷窥的数据和剩余流拼接后交给 untar
		combined := io.MultiReader(bytes.NewReader(peekBuf[:n]), gzReader)
		return untar(combined, destDir)
	}

	// 非 tar 的普通 gz 文件：将偷窥的数据和剩余流拼接后写入磁盘
	outputPath := filepath.Join(destDir, outputName)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create parent directory error: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %s error: %w", outputPath, err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Errorf("close output file error: %v", err)
		}
	}()

	remaining := io.MultiReader(bytes.NewReader(peekBuf[:n]), gzReader)
	if _, err := io.Copy(outFile, remaining); err != nil {
		return fmt.Errorf("write decompressed file error: %w", err)
	}

	return nil
}

// ValidateArchivePath 验证压缩包内的路径是否安全（防止路径遍历攻击）
func ValidateArchivePath(name string) error {
	if name == "" {
		return fmt.Errorf("empty path")
	}
	// 清理路径
	cleaned := filepath.Clean(name)
	// 不允许绝对路径
	if filepath.IsAbs(cleaned) {
		return fmt.Errorf("absolute path not allowed: %s", name)
	}
	// 不允许路径遍历
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path traversal not allowed: %s", name)
	}
	return nil
}

// IsHiddenEntry 判断文件或目录名是否为应跳过的隐藏/系统条目
// 跳过规则：
//   - 以 "." 开头的文件或目录（包括 .DS_Store、.git、.vscode、._xxx 等 AppleDouble 资源分叉文件）
//   - "__MACOSX" 目录（macOS 压缩包自动生成的元数据目录）
func IsHiddenEntry(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	if name == "__MACOSX" {
		return true
	}
	return false
}
