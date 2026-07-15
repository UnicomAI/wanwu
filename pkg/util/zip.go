package util

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
)

// ZipDir 将目录打包为 zip 格式数据。
// srcDir: 源目录路径，支持两种模式：
//   - "/path/to/dir"：包含最后一级目录名，zip 内容为 "dir/file1.txt"
//   - "/path/to/dir/."：不包含最后一级目录名，zip 内容为 "file1.txt"
func ZipDir(srcDir string) ([]byte, error) {
	var buf bytes.Buffer
	err := zipDirCore(srcDir, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ZipDirToLocal 将目录打包为 zip 格式文件。
// srcDir: 源目录路径，支持两种模式：
//   - "/path/to/dir"：包含最后一级目录名，zip 内容为 "dir/file1.txt"
//   - "/path/to/dir/."：不包含最后一级目录名，zip 内容为 "file1.txt"
//
// destZipPath: 目标 zip 文件路径，例如: "/path/to/output.zip"
func ZipDirToLocal(srcDir, destZipPath string) error {
	// 创建目标 zip 文件
	zipFile, err := os.Create(destZipPath)
	if err != nil {
		return fmt.Errorf("create zip file error: %w", err)
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			log.Errorf("close zip file error: %v", err)
		}
	}()
	return zipDirCore(srcDir, zipFile)
}

func UnzipDir(ctx context.Context, localFilePath string, destDir string) (extractDir string, err error) {
	fileReader, err := zip.OpenReader(localFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err1 := fileReader.Close(); err1 != nil {
			log.Errorf("UnzipDir file (%s) close error: %v", localFilePath, err1)
		}
	}()

	for _, f := range fileReader.File {
		// EFS(bit 11)未置位 → 本地编码，默认按 GBK/GB18030 解码成 UTF-8
		decodeFileName := f.Name
		if f.Flags&ZipEFSFlag == 0 {
			decodeFileName = DecodeGBKToUTF8(decodeFileName)
		}
		// 构建完整的文件路径
		destFilePath := filepath.Join(destDir, decodeFileName)
		// 检查是否为目录
		if f.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(destFilePath, f.Mode()); err != nil {
				log.Errorf("UnzipDir create directory (%s) error: %v", destFilePath, err)
			}
			continue
		}
		// 我们需要确保所有的文件夹都已经创建好
		err = os.MkdirAll(filepath.Dir(destFilePath), f.Mode())
		if err != nil {
			return "", err
		}
		//写入文件
		err = writeUnzipFile(f, destFilePath)
		if err != nil {
			return "", err
		}
	}
	return destDir, nil
}

func zipDirCore(srcDir string, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			log.Errorf("close zip writer error: %v", err)
		}
	}()

	skipBase := strings.HasSuffix(srcDir, string(os.PathSeparator)+".")
	srcDir = filepath.Clean(srcDir)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", srcDir)
	}

	var baseName string
	if skipBase {
		baseName = ""
	} else {
		baseName = filepath.Base(srcDir)
	}

	err := filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return fmt.Errorf("get relative path failed: %w", err)
		}

		var zipPath string
		if relPath == "." {
			if skipBase {
				return nil
			}
			zipPath = baseName + "/"
		} else {
			if skipBase {
				zipPath = filepath.ToSlash(relPath)
			} else {
				zipPath = filepath.ToSlash(filepath.Join(baseName, relPath))
			}
		}

		if info.IsDir() {
			zipPath += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create zip header failed: %w", err)
		}
		header.Name = zipPath
		header.Method = zip.Store

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create zip entry failed: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open file failed: %w", err)
		}
		_, err = io.Copy(writer, file)
		_ = file.Close()
		if err != nil {
			return fmt.Errorf("write file content failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk directory failed: %w", err)
	}
	return nil
}

// writeUnzipFile 写入文件
func writeUnzipFile(zipFile *zip.File, destFilePath string) error {
	//打开目标文件
	destFile, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
	if err != nil {
		return err
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			log.Errorf("writeUnzipFile file close error: %v", err)
		}
	}()

	//打开源压缩文件
	sourceFile, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			log.Errorf("writeUnzipFile file close error: %v", err)
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

// CleanZipEntryName 规范化 zip 内文件名（统一斜杠、清理 . 与 ..、拒绝绝对路径与盘符）。
func CleanZipEntryName(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." {
			continue
		}
		if part == ".." || strings.Contains(part, ":") {
			return "", fmt.Errorf("invalid zip entry path: %s", name)
		}
	}
	cleanName := path.Clean(name)
	if cleanName == "." || path.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, "../") {
		return "", fmt.Errorf("invalid zip entry path: %s", name)
	}
	return cleanName, nil
}

// IsZipEntryInSubDir 判断 cleanName（来自 CleanZipEntryName）是否落在 zip 内某个子目录下，
// subDir="." 视为整个压缩包根。
func IsZipEntryInSubDir(cleanName, subDir string) bool {
	if subDir == "." {
		return true
	}
	return cleanName == subDir || strings.HasPrefix(cleanName, subDir+"/")
}

// RelativeZipEntryName 返回 cleanName 相对 subDir 的剩余路径，subDir="." 时直接返回 cleanName。
func RelativeZipEntryName(cleanName, subDir string) string {
	if subDir == "." {
		return cleanName
	}
	if cleanName == subDir {
		return "."
	}
	return strings.TrimPrefix(cleanName, subDir+"/")
}

// WriteZipEntry 将 zip.File 解压到目标路径（按 file.Mode().Perm() 设权限，0 时退回 0644）。
// 不创建父目录、不做路径安全校验，由调用方保证 targetPath 已经过路径校验。
func WriteZipEntry(file *zip.File, targetPath string) error {
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	perm := file.Mode().Perm()
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, IfElse(perm == 0, os.FileMode(0644), perm))
	if err != nil {
		return err
	}
	defer func() { _ = target.Close() }()

	_, err = io.Copy(target, source)
	return err
}

// UnzipSubDir 安全地把 reader 中位于 subDir 下的所有条目解压到 destDir。
// destDir 不存在会自动创建；拒绝 symlink 条目、越界条目；保持原文件权限。
func UnzipSubDir(reader *zip.Reader, subDir, destDir string) error {
	cleanDest, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	cleanDestWithSep := cleanDest + string(os.PathSeparator)

	for _, file := range reader.File {
		// EFS(bit 11)未置位 → 本地编码，先按 GBK/GB18030 解码成 UTF-8，再做路径清理。
		// 必须先解码再 CleanZipEntryName：GBK 双字节字符第二字节可能为 0x5C('\')，
		// 否则 CleanZipEntryName 的 ReplaceAll("\\","/") 会截坏 GBK 字符。
		entryName := file.Name
		if file.Flags&ZipEFSFlag == 0 {
			entryName = DecodeGBKToUTF8(entryName)
		}
		cleanName, err := CleanZipEntryName(entryName)
		if err != nil {
			return err
		}
		if !IsZipEntryInSubDir(cleanName, subDir) {
			continue
		}
		relativeName := RelativeZipEntryName(cleanName, subDir)
		if relativeName == "." {
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("zip entry symlink is not supported: %s", file.Name)
		}

		targetPath := filepath.Join(cleanDest, filepath.FromSlash(relativeName))
		absTargetPath, err := filepath.Abs(targetPath)
		if err != nil {
			return err
		}
		if absTargetPath != cleanDest && !strings.HasPrefix(absTargetPath, cleanDestWithSep) {
			return fmt.Errorf("zip entry escapes target directory: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			perm := file.Mode().Perm()
			if err := os.MkdirAll(absTargetPath, IfElse(perm == 0, os.FileMode(0755), perm)); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(absTargetPath), 0755); err != nil {
			return err
		}
		if err := WriteZipEntry(file, absTargetPath); err != nil {
			return err
		}
	}
	return nil
}
