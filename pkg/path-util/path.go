package path_util

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var windowsDrivePathRE = regexp.MustCompile(`^[A-Za-z]:[\\/]?`)

// CleanRelPath 将用户传入的相对路径规范化为斜杠形式。
func CleanRelPath(p string, allowEmpty bool) (string, error) {
	if p == "" {
		if allowEmpty {
			return "", nil
		}
		return "", errors.New("path is required")
	}
	if strings.ContainsRune(p, '\x00') {
		return "", errors.New("path contains null byte")
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") || windowsDrivePathRE.MatchString(p) || filepath.IsAbs(p) {
		return "", errors.New("absolute path not allowed")
	}

	normalized := strings.ReplaceAll(p, "\\", "/")
	cleaned := path.Clean(normalized)
	if cleaned == "." {
		if allowEmpty {
			return "", nil
		}
		return "", errors.New("path is required")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("path traversal not allowed")
	}
	return cleaned, nil
}

// JoinWithinBase 将相对路径拼到 base 下，并校验结果仍在 base 内部。
func JoinWithinBase(basePath, relPath string, allowEmpty bool) (string, string, error) {
	if basePath == "" {
		return "", "", errors.New("base path is required")
	}
	cleanRel, err := CleanRelPath(relPath, allowEmpty)
	if err != nil {
		return "", "", err
	}

	fullPath := filepath.Join(basePath, filepath.FromSlash(cleanRel))
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", "", err
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", "", err
	}
	rel, err := filepath.Rel(absBase, absFull)
	if err != nil {
		return "", "", err
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") || filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("path outside workspace")
	}
	if err := validateRealPathWithinBase(absBase, absFull); err != nil {
		return "", "", err
	}
	return absFull, cleanRel, nil
}

// validateRealPathWithinBase 校验真实路径解析后仍位于 base 内。
func validateRealPathWithinBase(absBase, absFull string) error {
	realBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	existingFull, err := nearestExistingPath(absFull)
	if err != nil {
		return err
	}
	if existingFull == "" {
		return nil
	}
	realFull, err := filepath.EvalSymlinks(existingFull)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	rel, err := filepath.Rel(realBase, realFull)
	if err != nil {
		return err
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") || filepath.IsAbs(rel) {
		return fmt.Errorf("path outside workspace")
	}
	return nil
}

// nearestExistingPath 向上查找离目标最近的已存在路径。
func nearestExistingPath(absPath string) (string, error) {
	current := absPath
	for {
		if _, err := os.Lstat(current); err == nil {
			return current, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", nil
		}
		current = parent
	}
}

// EnsureNoSymlinkInPath 校验 basePath -> targetPath 之间每一段路径都不是 symlink。
// includeTarget=true 时把 targetPath 本身也纳入检查；
// 当 basePath 与 targetPath 相等时退化为只检查 basePath 自身。
func EnsureNoSymlinkInPath(basePath, targetPath string, includeTarget bool) error {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}
	if err := EnsureNoSymlinkInAncestors(absBase, true); err != nil {
		return err
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return err
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == "." {
		return nil
	}
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") || filepath.IsAbs(rel) {
		return fmt.Errorf("path outside workspace")
	}

	parts := strings.Split(relSlash, "/")
	if !includeTarget && len(parts) > 0 {
		parts = parts[:len(parts)-1]
	}
	current := absBase
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink path not allowed")
		}
	}
	return nil
}

// EnsureNoSymlinkInAncestors 校验 absPath 的所有祖先目录都不是 symlink。
// includePath=true 时 absPath 自身也纳入检查。
func EnsureNoSymlinkInAncestors(absPath string, includePath bool) error {
	current := absPath
	if !includePath {
		current = filepath.Dir(current)
	}
	for {
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink path not allowed")
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
		current = parent
	}
}
